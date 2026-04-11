// Package middleware provides the gateway-aware identity middleware shared
// across XMart Cloud services. It parses the X-* headers emitted by
// core-manager and exposes a typed *Identity via FromContext.
package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// identityKey is an unexported type used as the Locals key so no other
// package can collide with it. This is the Go stdlib-recommended pattern
// for request-scoped context values.
type identityKey struct{}

// ErrNoIdentity is returned by FromContext when Tenant() was not run on
// the request (i.e. the middleware chain is misconfigured).
var ErrNoIdentity = errors.New("xmart-platform/middleware: no identity in context (Tenant middleware not in chain)")

// Options configures the Tenant middleware behaviour.
type Options struct {
	// RequireValid rejects requests with X-Is-Valid != "true" with 401.
	// Default true. Set to false only for public endpoints that still want
	// to read the identity when present (buyer-facing pay endpoints, webhooks).
	RequireValid bool

	// RequireAccount rejects requests without X-Account-Uid with 401.
	// Default true. Set to false for global endpoints that operate
	// cross-account (admin/platform management).
	RequireAccount bool

	// OnReject is called when the middleware rejects a request. It should
	// write the response body. If nil, a minimal 401 JSON body is emitted.
	OnReject func(c fiber.Ctx, reason string) error
}

// DefaultOptions returns the strict defaults: require valid + account.
func DefaultOptions() Options {
	return Options{
		RequireValid:   true,
		RequireAccount: true,
	}
}

// Tenant returns a Fiber v3 handler that parses the gateway identity
// headers into an *Identity and stores it in request locals. Downstream
// handlers retrieve it via FromContext.
//
// With the default options a request is rejected with 401 when:
//   - X-Is-Valid is not "true"
//   - X-User-Uid is empty
//   - X-Account-Uid is empty
func Tenant(opts ...Options) fiber.Handler {
	o := DefaultOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	reject := o.OnReject
	if reject == nil {
		reject = defaultReject
	}

	return func(c fiber.Ctx) error {
		id := parseIdentity(c)

		if o.RequireValid && !id.IsValid {
			return reject(c, "invalid_identity")
		}
		if id.UserUID == "" {
			return reject(c, "missing_user_uid")
		}
		if o.RequireAccount && id.AccountUID == "" {
			return reject(c, "missing_account_uid")
		}

		c.Locals(identityKey{}, id)
		return c.Next()
	}
}

// FromContext returns the *Identity attached to the request by Tenant().
// It returns ErrNoIdentity if the middleware was not registered.
func FromContext(c fiber.Ctx) (*Identity, error) {
	id, ok := c.Locals(identityKey{}).(*Identity)
	if !ok || id == nil {
		return nil, ErrNoIdentity
	}
	return id, nil
}

// MustFromContext is FromContext that panics on ErrNoIdentity. Use only in
// handlers protected by Tenant() — the panic indicates a programming error
// (middleware chain misconfiguration), not a runtime condition.
func MustFromContext(c fiber.Ctx) *Identity {
	id, err := FromContext(c)
	if err != nil {
		panic(err)
	}
	return id
}

// GetIdentity is an alias for MustFromContext, kept for call-site
// ergonomics. Prefer FromContext in new code when you need to distinguish
// "no identity" from "handler error".
func GetIdentity(c fiber.Ctx) *Identity { return MustFromContext(c) }

// Convenience accessors — safe to call from any handler that runs after Tenant().
// They panic if Tenant() was not in the chain (programming error).

func GetAccountUID(c fiber.Ctx) string { return MustFromContext(c).AccountUID }
func GetUserUID(c fiber.Ctx) string    { return MustFromContext(c).UserUID }
func GetTimezone(c fiber.Ctx) string   { return MustFromContext(c).Timezone }
func GetRole(c fiber.Ctx) string       { return MustFromContext(c).Role }

// parseIdentity builds an *Identity from the incoming request headers.
// It is exported as a package-internal helper so tests can exercise it
// without a full Fiber app.
func parseIdentity(c fiber.Ctx) *Identity {
	return &Identity{
		IsValid:        parseBool(c.Get(HeaderIsValid)),
		AuthType:       c.Get(HeaderAuthType),
		UserUID:        c.Get(HeaderUserUID),
		Username:       c.Get(HeaderUsername),
		Email:          c.Get(HeaderEmail),
		Role:           c.Get(HeaderRole),
		IsSuperAdmin:   parseBool(c.Get(HeaderIsSuperAdmin)),
		IsInternal:     parseBool(c.Get(HeaderIsInternal)),
		UserStatus:     c.Get(HeaderUserStatus),
		Timezone:       c.Get(HeaderTimezone),
		TimezoneOffset: c.Get(HeaderTimezoneOffset),
		AccountUID:     c.Get(HeaderAccountUID),
		ProductSlugs:   parseCSV(c.Get(HeaderProductSlugs)),
		RequestID:      c.Get(HeaderRequestID),
	}
}

func parseBool(v string) bool {
	switch strings.ToLower(v) {
	case "true", "1", "yes", "y":
		return true
	}
	return false
}

func parseCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func defaultReject(c fiber.Ctx, reason string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success": false,
		"code":    "unauthorized",
		"message": reason,
	})
}
