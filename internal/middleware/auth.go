package middleware

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const UserIDKey = "user_id"

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-User-ID")
		fmt.Printf("Middleware: X-User-ID from header = '%s'\n", userIDStr)
		if userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing X-User-ID"})
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			fmt.Printf("Middleware: Parse error for '%s': %v\n", userIDStr, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user id"})
		}

		c.Locals(UserIDKey, userID)
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) int64 {
	if id, ok := c.Locals(UserIDKey).(int64); ok && id != 0 {
		return id
	}
	// fallback for public routes
	id, _ := strconv.ParseInt(c.Get("X-User-ID"), 10, 64)
	return id
}
