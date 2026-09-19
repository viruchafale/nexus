// Package config loads NEXUS configuration from environment variables.
//
// AI functionality is optional. No secrets are hardcoded and
// config values must be passed explicitly (no global mutable state).
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Defaults for optional AI configuration.
const (
	DefaultAIBaseURL = "https://api.openai.com/v1"
	DefaultAIModel   = "gpt-4o-mini"
	DefaultAITimeout = 60 * time.Second
)

// Config holds runtime configuration. Use Load to build it from the environment.
type Config struct {
	Debug   bool
	NoColor bool

	// AIProvider selects the backend. Only "openai" (OpenAI-compatible
	// chat-completions API) is supported; anything else is rejected
	// with a clear error instead of guessing.
	AIProvider string
	// AIModel is the model name for `nexus ask`. Empty means provider default.
	AIModel string
	// AIBaseURL overrides the API endpoint (self-hosted gateways, proxies).
	AIBaseURL string
	// AITimeout bounds the whole AI request.
	AITimeout time.Duration
	// OpenAIKey enables `nexus ask`. Empty means AI is disabled.
	// It is never printed by any command.
	OpenAIKey string
}

// HasAIKey reports whether AI troubleshooting can be attempted.
func (c Config) HasAIKey() bool {
	return strings.TrimSpace(c.OpenAIKey) != ""
}

// Load reads configuration from process environment.
func Load() Config {
	return parse(os.Getenv)
}

// parse is testable core of Load without touching global env directly.
func parse(getenv func(string) string) Config {
	provider := strings.ToLower(strings.TrimSpace(getenv("NEXUS_AI_PROVIDER")))
	if provider == "" {
		provider = "openai"
	}
	baseURL := strings.TrimSpace(getenv("NEXUS_AI_BASE_URL"))
	if baseURL == "" {
		baseURL = DefaultAIBaseURL
	}
	return Config{
		Debug:      parseBool(getenv("NEXUS_DEBUG")),
		NoColor:    parseBool(getenv("NEXUS_NO_COLOR")) || parseBool(getenv("NO_COLOR")),
		AIProvider: provider,
		AIModel:    strings.TrimSpace(getenv("NEXUS_AI_MODEL")),
		AIBaseURL:  baseURL,
		AITimeout:  parseTimeout(getenv("NEXUS_AI_TIMEOUT_SECS"), DefaultAITimeout),
		OpenAIKey:  strings.TrimSpace(getenv("NEXUS_OPENAI_API_KEY")),
	}
}

// parseTimeout reads seconds from env, falling back to def on empty/invalid.
func parseTimeout(s string, def time.Duration) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return time.Duration(n) * time.Second
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return false
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		// Accept common truthy values ParseBool rejects.
		return s == "1" || s == "yes" || s == "y" || s == "on"
	}
	return b
}
