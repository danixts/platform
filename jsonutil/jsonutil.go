package jsonutil

import "github.com/bytedance/sonic"

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

const Redacted = "[REDACTED]"

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
