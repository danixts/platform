// Package jsonutil wraps bytedance/sonic for JSON (de)serialization and
// provides sanitisation helpers that redact sensitive fields from payloads
// before logging. The sensitive key list is intentionally conservative —
// callers can extend it per service.
package jsonutil

import "github.com/bytedance/sonic"

// SensitiveKeys is the set of JSON field names that are always redacted by
// Sanitize and SanitizeHeaders. It is a package-level variable so services
// can add their own keys at init time.
var SensitiveKeys = map[string]bool{
	"password":              true,
	"Password":              true,
	"userPassword":          true,
	"userName":              true,
	"authorizationId":       true,
	"accountId":             true,
	"apiKey":                true,
	"apiKeyServicio":        true,
	"Authorization":         true,
	"authorization":         true,
	"cert_passphrase":       true,
	"certPassphrase":        true,
	"public_token_enc":      true,
	"bank_password_enc":     true,
	"api_key_enc":           true,
	"token":                 true,
	"newAuthorizationId":    true,
	"actualAuthorizationId": true,
}

// Redacted is the placeholder written in place of sensitive values.
const Redacted = "[REDACTED]"

// Sanitize walks a decoded JSON value (map[string]any / []any / scalar)
// and returns a copy with SensitiveKeys redacted. The input is not mutated.
func Sanitize(v any) any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		out := make(map[string]any, len(m))
		for k, val := range m {
			if SensitiveKeys[k] {
				out[k] = Redacted
				continue
			}
			out[k] = Sanitize(val)
		}
		return out
	}
	if s, ok := v.([]any); ok {
		out := make([]any, len(s))
		for i, val := range s {
			out[i] = Sanitize(val)
		}
		return out
	}
	return v
}

// SanitizeHeaders returns a copy of h with SensitiveKeys redacted. Handy
// for logging HTTP request/response headers without leaking credentials.
func SanitizeHeaders(h map[string]string) map[string]string {
	if h == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(h))
	for k, v := range h {
		if SensitiveKeys[k] {
			out[k] = Redacted
			continue
		}
		out[k] = v
	}
	return out
}

// TryUnmarshal best-effort decodes a JSON byte slice into a dynamic value.
// If decoding fails, it returns a map containing the raw string so the
// caller can still log the payload.
func TryUnmarshal(data []byte) any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var v any
	if err := sonic.Unmarshal(data, &v); err != nil {
		return map[string]any{"raw": string(data)}
	}
	return v
}
