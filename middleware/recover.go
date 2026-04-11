package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"

	"github.com/danixts/platform/logger"
	"github.com/danixts/platform/response"
)

// Recover returns a Fiber handler that traps panics in the downstream
// chain, logs the panic value plus a stack trace, and replies with a
// generic 500. It is safe to place at the top of the middleware stack.
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
