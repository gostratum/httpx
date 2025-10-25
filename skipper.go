package httpx

import (
	"regexp"
	"strings"
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
func NewSkipper(cfg Config) (func(method, path string) bool, error) {
	// Get user-defined rules from config
	userRules := cfg.Request.Logging.DisabledURLs

	// Get configurable health endpoint paths
	healthzPath := cfg.Health.ReadinessPath
	livezPath := cfg.Health.LivenessPath
	infoPath := cfg.Health.InfoPath

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
