package middleware

import "github.com/gofiber/fiber/v3"

type identityKey struct{}

func FromContext(c fiber.Ctx) (*Identity, error) {
	id, ok := c.Locals(identityKey{}).(*Identity)
	if !ok || id == nil {
		return nil, ErrNoIdentity
	}
	return id, nil
}

func SetIdentity(c fiber.Ctx, id *Identity) {
	c.Locals(identityKey{}, id)
}

func MustFromContext(c fiber.Ctx) *Identity {
	id, err := FromContext(c)
	if err != nil {
		panic(err)
	}
	return id
}

func GetIdentity(c fiber.Ctx) *Identity    { return MustFromContext(c) }
func GetIsValid(c fiber.Ctx) bool          { return MustFromContext(c).IsValid }
func GetAuthType(c fiber.Ctx) string       { return MustFromContext(c).AuthType }
func GetUserUID(c fiber.Ctx) string        { return MustFromContext(c).UserUID }
func GetUsername(c fiber.Ctx) string       { return MustFromContext(c).Username }
func GetEmail(c fiber.Ctx) string          { return MustFromContext(c).Email }
func GetRole(c fiber.Ctx) string           { return MustFromContext(c).Role }
func GetIsSuperAdmin(c fiber.Ctx) bool     { return MustFromContext(c).IsSuperAdmin }
func GetIsInternal(c fiber.Ctx) bool       { return MustFromContext(c).IsInternal }
func GetUserStatus(c fiber.Ctx) string     { return MustFromContext(c).UserStatus }
func GetTimezone(c fiber.Ctx) string       { return MustFromContext(c).Timezone }
func GetTimezoneOffset(c fiber.Ctx) string { return MustFromContext(c).TimezoneOffset }
func GetAccountUID(c fiber.Ctx) string     { return MustFromContext(c).AccountUID }
func GetOrgUID(c fiber.Ctx) string         { return MustFromContext(c).OrgUID }
func GetProductSlugs(c fiber.Ctx) []string { return MustFromContext(c).ProductSlugs }
func GetRequestID(c fiber.Ctx) string      { return MustFromContext(c).RequestID }
