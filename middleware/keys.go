package middleware

// Header names emitted by the core-manager gateway. Keep this file in sync
// with the gateway contract — adding/removing a header here is a breaking
// change for every consumer.
const (
	HeaderIsValid        = "X-Is-Valid"
	HeaderAuthType       = "X-Auth-Type"
	HeaderUserUID        = "X-User-Uid"
	HeaderUsername       = "X-Username"
	HeaderEmail          = "X-Email"
	HeaderRole           = "X-Role"
	HeaderIsSuperAdmin   = "X-Is-Super-Admin"
	HeaderIsInternal     = "X-Is-Internal"
	HeaderUserStatus     = "X-User-Status"
	HeaderTimezone       = "X-Timezone"
	HeaderTimezoneOffset = "X-Timezone-Offset"
	HeaderAccountUID     = "X-Account-Uid"
	HeaderProductSlugs   = "X-Product-Slugs"
	HeaderRequestID      = "X-Request-Id"
)
