package httpx

import (
	"strings"

	"github.com/spf13/viper"
)

// SanitizeViper returns a sanitized copy of the `http` subtree from the provided Viper.
// It redacts values whose keys look like secrets (password, token, key, secret, private, pem).
func SanitizeViper(v *viper.Viper) map[string]any {
	sub := v.Sub("http")
	if sub == nil {
		return map[string]any{}
	}

	raw := sub.AllSettings()
	return sanitizeMap(raw)
}

func sanitizeMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		lk := strings.ToLower(k)
		switch val := v.(type) {
		case map[string]interface{}:
			// viper may return map[string]interface{}; convert and sanitize
			m := make(map[string]any, len(val))
			for kk, vv := range val {
				m[kk] = vv
			}
			out[k] = sanitizeMap(m)
		default:
			if isSecretKey(lk) {
				out[k] = "[redacted]"
			} else {
				out[k] = v
			}
		}
	}
	return out
}

func isSecretKey(k string) bool {
	// simple substring checks; intentionally conservative
	secrets := []string{"password", "passwd", "secret", "token", "key", "private", "pem", "hmac", "api_key", "apikey"}
	for _, s := range secrets {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}
