package handler

import (
	"log"
	"post-service/internal/dto"
	"post-service/internal/middleware"
	"post-service/internal/service"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type PostHandler struct {
	svc      service.PostService
	validate *validator.Validate
}

func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{
		svc:      svc,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (h *PostHandler) Create(c *fiber.Ctx) error {
	var req dto.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	log.Printf("CreatePost: title=%q, content_len=%d, author_id=%d", req.Title, len(req.Content), req.AuthorID)
	if err := h.validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userID := middleware.GetUserID(c)
	log.Printf("CreatePost: userID from header=%d", userID)
	if req.AuthorID == 0 {
		req.AuthorID = userID
	}
	username := c.Get("X-Username")
	avatarURL := c.Get("X-User-Avatar")

	post, err := h.svc.Create(c.Context(), userID, username, avatarURL, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}
	userID := middleware.GetUserID(c)

	post, err := h.svc.GetPost(c.Context(), id, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}
	if userID != 0 {
		_ = h.svc.IncrementView(c.Context(), id, userID)
	}
	return c.JSON(post)
}

func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	q := dto.ListPostsQuery{
		Limit:  c.QueryInt("limit", 20),
		Offset: c.QueryInt("offset", 0),
		Sort:   c.Query("sort", "new"),
		Author: c.Query("author"),
		Tag:    c.Query("tag"),
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	userID := middleware.GetUserID(c)
	posts, err := h.svc.ListPosts(c.Context(), q, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(posts)
}

func (h *PostHandler) HotPosts(c *fiber.Ctx) error {
	c.Request().URI().SetQueryStringBytes(append(c.Request().URI().QueryString(), []byte("&sort=hot")...))
	return h.ListPosts(c)
}

func (h *PostHandler) TopPosts(c *fiber.Ctx) error {
	c.Request().URI().SetQueryStringBytes(append(c.Request().URI().QueryString(), []byte("&sort=top")...))
	return h.ListPosts(c)
}

func (h *PostHandler) Search(c *fiber.Ctx) error {
	q := c.Query("q")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	posts, err := h.svc.SearchPosts(c.Context(), q, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(posts)
}

func (h *PostHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}
	var req dto.UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	userID := middleware.GetUserID(c)

	post, err := h.svc.UpdatePost(c.Context(), id, userID, &req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(post)
}

func (h *PostHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}
	userID := middleware.GetUserID(c)

	if err := h.svc.DeletePost(c.Context(), id, userID); err != nil {
		return handleServiceError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostHandler) Like(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}
	userID := middleware.GetUserID(c)

	if err := h.svc.Like(c.Context(), id, userID); err != nil {
		if err.Error() == "already liked" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already liked"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"action": "liked"})
}

func (h *PostHandler) Unlike(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid post id"})
	}
	userID := middleware.GetUserID(c)

	if err := h.svc.Unlike(c.Context(), id, userID); err != nil {
		if err.Error() == "not liked" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "not liked"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"action": "unliked"})
}

// --- Хелперы ---

func parseID(c *fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("id"), 10, 64)
}

func handleServiceError(c *fiber.Ctx, err error) error {
	switch err.Error() {
	case "forbidden":
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	case "post not found":
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}
