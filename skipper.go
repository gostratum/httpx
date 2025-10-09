package httpx

import (
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// DisabledURL represents a URL pattern to skip in request logging
type DisabledURL struct {
	Method     string `mapstructure:"method"`
	URLPattern string `mapstructure:"urlPattern"`
}

// compiledRule represents a compiled regex rule for URL matching
type compiledRule struct {
	method string
	re     *regexp.Regexp
}

// NewSkipper creates a function that determines whether to skip logging for a given request
// It loads patterns from config and ensures actuator/health endpoints are skipped by default
func NewSkipper(v *viper.Viper) (func(method, path string) bool, error) {
	// Load user-defined rules from config
	var userRules []DisabledURL
	if err := v.UnmarshalKey("http.request.logging.disabled_urls", &userRules); err != nil {
		// Ignore unmarshal errors for missing/empty config
		userRules = nil
	}

	// Get configurable health endpoint paths
	healthzPath := v.GetString("http.health.readiness_path")
	if healthzPath == "" {
		healthzPath = "/healthz"
	}

	livezPath := v.GetString("http.health.liveness_path")
	if livezPath == "" {
		livezPath = "/livez"
	}

	infoPath := v.GetString("http.health.info_path")
	if infoPath == "" {
		infoPath = "/actuator/info"
	}

	// Always ensure health endpoints are skipped by default (using configured paths)
	defaultRules := []DisabledURL{
		{Method: "GET", URLPattern: "^" + healthzPath + "$"},
		{Method: "GET", URLPattern: "^" + livezPath + "$"},
		{Method: "GET", URLPattern: "^" + strings.Replace(infoPath, "/info", "/.*", 1)}, // Skip all actuator endpoints
	}

	// Combine default rules with user rules
	allRules := append(defaultRules, userRules...)

	// Compile all rules
	var compiledRules []compiledRule
	for _, rule := range allRules {
		re, err := regexp.Compile(rule.URLPattern)
		if err != nil {
			return nil, err
		}
		compiledRules = append(compiledRules, compiledRule{
			method: strings.ToUpper(rule.Method),
			re:     re,
		})
	}

	// Return the skipper function
	return func(method, path string) bool {
		normalizedMethod := strings.ToUpper(method)
		for _, rule := range compiledRules {
			// If method is empty or matches, check the URL pattern
			if (rule.method == "" || rule.method == normalizedMethod) && rule.re.MatchString(path) {
				return true
			}
		}
		return false
	}, nil
}
