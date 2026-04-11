package jsonutil

import "github.com/bytedance/sonic"

// SensitiveKeys is the set of JSON field names that are always redacted
// by Sanitize and SanitizeHeaders. It holds the generic credential-like
// names that apply to almost any service.
//
// Services with domain-specific sensitive fields should extend this map
// in an init() block at startup, for example:
//
//	func init() {
//	    jsonutil.SensitiveKeys["my_custom_secret_field"] = true
//	}
var SensitiveKeys = map[string]bool{
	"password":        true,
	"Password":        true,
	"userPassword":    true,
	"apiKey":          true,
	"api_key":         true,
	"apikey":          true,
	"Authorization":   true,
	"authorization":   true,
	"token":           true,
	"access_token":    true,
	"refresh_token":   true,
	"secret":          true,
	"client_secret":   true,
	"private_key":     true,
	"cert_passphrase": true,
	"certPassphrase":  true,
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
