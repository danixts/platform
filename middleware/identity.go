package middleware

import "slices"

// Identity is the request-scoped identity emitted by the gateway. It is
// populated by Tenant() from the X-* headers and made available via
// FromContext. The zero value is a valid (but unauthenticated) Identity.
type Identity struct {
	IsValid        bool
	AuthType       string // "jwt" | "apikey" | ...
	UserUID        string
	Username       string
	Email          string
	Role           string
	IsSuperAdmin   bool
	IsInternal     bool
	UserStatus     string // "active" | ...
	Timezone       string // e.g. "America/La_Paz"
	TimezoneOffset string // raw value from header, e.g. "-04:00"
	AccountUID     string
	ProductSlugs   []string
	RequestID      string
}

// HasProduct reports whether the account has access to the given product
// slug. Matching is case-sensitive; product slugs are expected to be
// lowercase (e.g. "crm-meta", "wabot").
func (i *Identity) HasProduct(slug string) bool {
	if i == nil {
		return false
	}
	return slices.Contains(i.ProductSlugs, slug)
}

// IsAuthenticated reports whether the gateway flagged this request as a
// valid authenticated identity. A non-nil Identity with IsValid=false means
// the request came through but the gateway could not attest the user.
func (i *Identity) IsAuthenticated() bool {
	return i != nil && i.IsValid && i.UserUID != ""
}
