package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/danixts/platform/logger"
)

// AccessLog returns a Fiber handler that logs a single structured line
// per request once the handler chain returns. It uses RequestIDFromContext
// to include the request id and, when Tenant() ran earlier in the chain,
// the account_uid / user_uid from the Identity.
//
// The log record is emitted at Info level via logger.Get(). Services that
// want more customisation can write their own middleware and drop this one.
func AccessLog() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		ev := logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Int("size", len(c.Response().Body())).
			Str("ip", c.IP()).
			Dur("latency", latency).
			Str("request_id", RequestIDFromContext(c))

		if id, ok := c.Locals(identityKey{}).(*Identity); ok && id != nil {
			ev = ev.Str("account_uid", id.AccountUID).Str("user_uid", id.UserUID)
		}
		if err != nil {
			ev = ev.Err(err)
		}
		ev.Msg("http")
		return err
	}
}
