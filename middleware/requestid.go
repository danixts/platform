package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/oklog/ulid/v2"
)

type requestIDKey struct{}

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(HeaderRequestID)
		if id == "" {
			id = ulid.Make().String()
		}
		c.Set(HeaderRequestID, id)
		c.Locals(requestIDKey{}, id)
		return c.Next()
	}
}

func RequestIDFromContext(c fiber.Ctx) string {
	if v, ok := c.Locals(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}
