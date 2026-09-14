package config

import "testing"

func TestParseDefaultsEmpty(t *testing.T) {
	c := parse(func(string) string { return "" })
	if c.Debug {
		t.Error("Debug should default to false")
	}
	if c.NoColor {
		t.Error("NoColor should default to false")
	}
	if c.HasAIKey() {
		t.Error("HasAIKey should be false when key is empty")
	}
	if c.AIModel != "" {
		t.Errorf("AIModel should default to empty, got %q", c.AIModel)
	}
}

func TestParseReadsEnv(t *testing.T) {
	env := map[string]string{
		"NEXUS_DEBUG":          "1",
		"NEXUS_NO_COLOR":       "true",
		"NEXUS_AI_MODEL":       "gpt-4o-mini",
		"NEXUS_OPENAI_API_KEY": "secret",
	}
	c := parse(func(k string) string { return env[k] })
	if !c.Debug {
		t.Error("Debug should be true")
	}
	if !c.NoColor {
		t.Error("NoColor should be true")
	}
	if c.AIModel != "gpt-4o-mini" {
		t.Errorf("AIModel = %q", c.AIModel)
	}
	if !c.HasAIKey() {
		t.Error("HasAIKey should be true")
	}
}

func TestParseRespectsStandardNoColor(t *testing.T) {
	c := parse(func(k string) string {
		if k == "NO_COLOR" {
			return "1"
		}
		return ""
	})
	if !c.NoColor {
		t.Error("NO_COLOR=1 should imply NoColor")
	}
}
