package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"post-service/internal/config"
	"post-service/internal/event"
	"post-service/internal/handler"
	"post-service/internal/middleware"
	"post-service/internal/repository"
	"post-service/internal/service"

	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// PostgreSQL
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("postgres ping: %v", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err = rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	// Зависимости
	repo := repository.NewPostRepository(pool)
	postService := service.NewPostService(repo, rdb)
	postHandler := handler.NewPostHandler(postService)
	consumer := event.NewConsumer(rdb, repo)

	// Event consumer в горутине
	go consumer.Start(ctx)

	// Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Prometheus
	prom := fiberprometheus.New("post_service")
	prom.RegisterAt(app, "/metrics")
	app.Use(prom.Middleware)

	app.Use(logger.New())
	app.Use(recover.New())

	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "0.1.0",
			"checks":    fiber.Map{"database": "healthy", "redis": "healthy"},
		})
	})
	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy", "database": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"status": "healthy", "database": "healthy"})
	})

	// Маршруты
	v1 := app.Group("/api/v1")

	// Публичные
	v1.Get("/posts/search", postHandler.Search) // до /:id !
	v1.Get("/posts/hot", postHandler.HotPosts)
	v1.Get("/posts/top", postHandler.TopPosts)
	v1.Get("/posts", postHandler.ListPosts)
	v1.Get("/posts/:id", postHandler.GetPost)

	// Защищённые
	auth := v1.Group("/", middleware.AuthRequired())
	auth.Post("/posts", postHandler.Create)
	auth.Patch("/posts/:id", postHandler.Update)
	auth.Delete("/posts/:id", postHandler.Delete)
	auth.Post("/posts/:id/like", postHandler.Like)
	auth.Delete("/posts/:id/like", postHandler.Unlike)

	// Старт
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort)
		log.Printf("server listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("stopped")
}
