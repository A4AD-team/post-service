package router

import (
	"post-service/internal/handler"
	"post-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, h *handler.PostHandler) {
	v1 := app.Group("/api/v1")

	// Public
	v1.Get("/posts/search", h.Search)
	v1.Get("/posts/hot", h.HotPosts)
	v1.Get("/posts/top", h.TopPosts)
	v1.Get("/posts", h.ListPosts)
	v1.Get("/posts/:id", h.GetPost)

	// Protected
	auth := v1.Group("/", middleware.AuthRequired())
	auth.Post("/posts", h.Create)
	auth.Patch("/posts/:id", h.Update)
	auth.Delete("/posts/:id", h.Delete)
	auth.Post("/posts/:id/like", h.Like)
	auth.Delete("/posts/:id/like", h.Unlike)
}
