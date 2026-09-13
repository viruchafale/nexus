// Package config loads NEXUS configuration from environment variables.
//
// AI functionality is optional. No secrets are hardcoded and
// config values must be passed explicitly (no global mutable state).
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration. Use Load to build it from the environment.
type Config struct {
	Debug   bool
	NoColor bool

	// AIModel is the optional model name for `nexus ask`.
	// Empty means "use provider default".
	AIModel string
	// OpenAIKey enables `nexus ask`. Empty means AI is disabled.
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
	return Config{
		Debug:     parseBool(getenv("NEXUS_DEBUG")),
		NoColor:   parseBool(getenv("NEXUS_NO_COLOR")) || parseBool(getenv("NO_COLOR")),
		AIModel:   strings.TrimSpace(getenv("NEXUS_AI_MODEL")),
		OpenAIKey: strings.TrimSpace(getenv("NEXUS_OPENAI_API_KEY")),
	}
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
