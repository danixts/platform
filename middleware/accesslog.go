package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/danixts/platform/logger"
)

// silentPaths suppresses access-log entries for high-frequency, low-signal
// endpoints. Health/readiness probes are scraped every few seconds by the
// kubelet and Prometheus scrapes /metrics every 30s — logging each request
// floods the logs without adding diagnostic value.
var silentPaths = map[string]struct{}{
	"/":                    {},
	"/health":              {},
	"/healthz":             {},
	"/ready":               {},
	"/readyz":              {},
	"/metrics":             {},
	"/api/v1/health":       {},
	"/api/v1/healthz":      {},
	"/api/v1/ready":        {},
	"/api/v1/readyz":       {},
}

func AccessLog() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, skip := silentPaths[c.Path()]; skip {
			return c.Next()
		}

		requestStart := time.Now()
		err := c.Next()
		latency := time.Since(requestStart)

		// If the handler returned an error but the ErrorHandler hasn't yet
		// written the status code, derive the real status from the fiber.Error
		// so the log doesn't show status=200 with error="<something>".
		status := c.Response().StatusCode()
		if err != nil && status < 400 {
			if fe, ok := err.(*fiber.Error); ok {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		logEvent := logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Int("size", len(c.Response().Body())).
			Str("ip", c.IP()).
			Dur("latency", latency).
			Str("request_id", RequestIDFromContext(c))

		if id, ok := c.Locals(identityKey{}).(*Identity); ok && id != nil {
			logEvent = logEvent.Str("account_uid", id.AccountUID).Str("user_uid", id.UserUID)
		}
		if err != nil {
			logEvent = logEvent.Err(err)
		}
		logEvent.Msg("http")
		return err
	}
}
