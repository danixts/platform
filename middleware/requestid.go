package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/oklog/ulid/v2"
)

// requestIDKey is the unexported Locals key for the request id. Defined
// separately from identityKey so RequestID() can be used standalone
// (e.g. on a public endpoint that does not require Tenant()).
type requestIDKey struct{}

// RequestID returns a Fiber handler that ensures every request has an
// X-Request-Id. If the header is already present (typical when the
// gateway forwarded it), that value is reused; otherwise a fresh ULID is
// generated. The id is echoed back on the response and stored in Locals
// so downstream handlers can read it with RequestIDFromContext.
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

// RequestIDFromContext returns the request id attached by RequestID. It
// returns the empty string when the middleware was not in the chain,
// which is safe to pass to loggers.
func RequestIDFromContext(c fiber.Ctx) string {
	if v, ok := c.Locals(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}
