package middleware

import "slices"

type Identity struct {
	IsValid        bool
	AuthType       string
	UserUID        string
	Username       string
	Email          string
	Role           string
	IsSuperAdmin   bool
	IsInternal     bool
	UserStatus     string
	Timezone       string
	TimezoneOffset string
	AccountUID     string
	ProductSlugs   []string
	RequestID      string
}

func (i *Identity) HasProduct(slug string) bool {
	if i == nil {
		return false
	}
	return slices.Contains(i.ProductSlugs, slug)
}

func (i *Identity) IsAuthenticated() bool {
	return i != nil && i.IsValid && i.UserUID != ""
}
