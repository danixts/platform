package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var ErrNoIdentity = errors.New("platform/middleware: no identity in context (Tenant middleware not in chain)")

type identityKey struct{}

type Options struct {
	RequireValid   bool
	RequireAccount bool
	OnReject       func(c fiber.Ctx, reason string) error
}

func DefaultOptions() Options {
	return Options{
		RequireValid:   true,
		RequireAccount: true,
	}
}

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

func FromContext(c fiber.Ctx) (*Identity, error) {
	id, ok := c.Locals(identityKey{}).(*Identity)
	if !ok || id == nil {
		return nil, ErrNoIdentity
	}
	return id, nil
}

func MustFromContext(c fiber.Ctx) *Identity {
	id, err := FromContext(c)
	if err != nil {
		panic(err)
	}
	return id
}

func GetIdentity(c fiber.Ctx) *Identity { return MustFromContext(c) }

func GetAccountUID(c fiber.Ctx) string { return MustFromContext(c).AccountUID }
func GetUserUID(c fiber.Ctx) string    { return MustFromContext(c).UserUID }
func GetTimezone(c fiber.Ctx) string   { return MustFromContext(c).Timezone }
func GetRole(c fiber.Ctx) string       { return MustFromContext(c).Role }

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
		OrgUID:         c.Get(HeaderOrgUID),
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
