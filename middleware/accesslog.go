package middleware

import (
	"bytes"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/danixts/platform/logger"
)

// silentPaths suppresses access-log entries for high-frequency, low-signal
// endpoints. Health/readiness probes are scraped every few seconds by the
// kubelet and Prometheus scrapes /metrics every 30s — logging each request
// floods the logs without adding diagnostic value.
var silentPaths = map[string]struct{}{
	"/":               {},
	"/health":         {},
	"/healthz":        {},
	"/ready":          {},
	"/readyz":         {},
	"/metrics":        {},
	"/api/v1/health":  {},
	"/api/v1/healthz": {},
	"/api/v1/ready":   {},
	"/api/v1/readyz":  {},
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
		// written the status code, derive the real status from the error so the
		// log doesn't show status=200 with error="<something>". Domain errors
		// carry their own status: without asking them, a 422 or a 409 gets logged
		// as 500 and reads like a server fault when hunting a bug.
		status := c.Response().StatusCode()
		if err != nil && status < 400 {
			status = statusFromError(err)
		}

		// For a streaming response (SSE) the body is written incrementally after
		// this middleware returns; Response().Body() would materialize the whole
		// stream and block forever, so the client never receives anything. Skip
		// the size read for text/event-stream.
		size := 0
		if !bytes.Contains(c.Response().Header.ContentType(), []byte("text/event-stream")) {
			size = len(c.Response().Body())
		}

		logEvent := logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Int("size", size).
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

// httpStatuser is implemented by domain errors that already know which HTTP
// status they map to. Kept as an interface so services keep their own error
// types without this package importing any of them.
type httpStatuser interface{ HTTPStatus() int }

func statusFromError(err error) int {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code
	}
	var hs httpStatuser
	if errors.As(err, &hs) {
		if code := hs.HTTPStatus(); code >= 100 {
			return code
		}
	}
	return fiber.StatusInternalServerError
}
