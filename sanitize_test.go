package httpx

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSanitizeViper(t *testing.T) {
	v := viper.New()
	v.Set("http.base_path", "/api")
	v.Set("http.auth.password", "s3cret")
	v.Set("http.auth.token", "tok-123")
	v.Set("http.client.api_key", "ak-456")

	out := SanitizeViper(v)

	// Ensure non-secret preserved
	if out["base_path"] != "/api" {
		t.Fatalf("expected base_path preserved")
	}

	// secrets should be redacted
	auth, ok := out["auth"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth map")
	}
	if auth["password"] != "[redacted]" {
		t.Fatalf("password not redacted")
	}
	if auth["token"] != "[redacted]" {
		t.Fatalf("token not redacted")
	}

	client, ok := out["client"].(map[string]any)
	if !ok {
		t.Fatalf("expected client map")
	}
	if client["api_key"] != "[redacted]" {
		t.Fatalf("api_key not redacted")
	}
}
