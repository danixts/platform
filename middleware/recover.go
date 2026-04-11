package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"

	"github.com/danixts/platform/logger"
	"github.com/danixts/platform/response"
)

func Recover() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error().
					Str("path", c.Path()).
					Str("request_id", RequestIDFromContext(c)).
					Interface("panic", r).
					Bytes("stack", debug.Stack()).
					Msg("panic recovered")
				err = response.Internal(c)
			}
		}()
		return c.Next()
	}
}
