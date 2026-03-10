package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Enabled {
		t.Error("default config should not be enabled")
	}
	if cfg.CommandPrefix != "!" {
		t.Errorf("default prefix = %q, want !", cfg.CommandPrefix)
	}
	if cfg.AllowedUsers == nil {
		t.Error("AllowedUsers should not be nil")
	}
	if len(cfg.AllowedUsers) != 0 {
		t.Error("default AllowedUsers should be empty")
	}
}

func TestConfig_IsAllowed_EmptyList(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.IsAllowed("U12345") {
		t.Error("empty allowed list should allow all users")
	}
}

func TestConfig_IsAllowed_WithList(t *testing.T) {
	cfg := &Config{AllowedUsers: []string{"U111", "U222"}}
	if !cfg.IsAllowed("U111") {
		t.Error("expected U111 to be allowed")
	}
	if !cfg.IsAllowed("U222") {
		t.Error("expected U222 to be allowed")
	}
	if cfg.IsAllowed("U333") {
		t.Error("expected U333 to NOT be allowed")
	}
}

func TestConfig_IsValid_BotAndApp(t *testing.T) {
	cfg := &Config{BotToken: "xoxb-test", AppToken: "xapp-test"}
	if !cfg.IsValid() {
		t.Error("config with bot_token + app_token should be valid")
	}
}

func TestConfig_IsValid_WebhookOnly(t *testing.T) {
	cfg := &Config{WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	if !cfg.IsValid() {
		t.Error("config with only webhook_url should be valid")
	}
}

func TestConfig_IsValid_Empty(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.IsValid() {
		t.Error("default config should NOT be valid")
	}
}

func TestConfig_IsValid_BotOnly(t *testing.T) {
	cfg := &Config{BotToken: "xoxb-test"}
	if cfg.IsValid() {
		t.Error("config with only bot_token (no app_token) should NOT be valid")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "slack.json")

	cfg := &Config{
		Enabled:       true,
		BotToken:      "xoxb-test-token",
		AppToken:      "xapp-test-token",
		SigningSecret:  "abc123",
		AppID:         "A12345",
		ChannelID:     "C67890",
		CommandPrefix: "!",
		TeamID:        "T11111",
		AllowedUsers:  []string{"U111", "U222"},
	}

	if err := cfg.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	loaded := LoadConfigFromFile(path)
	if loaded.BotToken != "xoxb-test-token" {
		t.Errorf("BotToken = %q, want xoxb-test-token", loaded.BotToken)
	}
	if loaded.AppToken != "xapp-test-token" {
		t.Errorf("AppToken = %q, want xapp-test-token", loaded.AppToken)
	}
	if loaded.SigningSecret != "abc123" {
		t.Errorf("SigningSecret = %q, want abc123", loaded.SigningSecret)
	}
	if loaded.TeamID != "T11111" {
		t.Errorf("TeamID = %q, want T11111", loaded.TeamID)
	}
	if len(loaded.AllowedUsers) != 2 {
		t.Errorf("AllowedUsers len = %d, want 2", len(loaded.AllowedUsers))
	}
	if !loaded.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestConfig_LoadMissing(t *testing.T) {
	cfg := LoadConfigFromFile("/nonexistent/path/slack.json")
	if cfg.Enabled {
		t.Error("missing config should not be enabled")
	}
	if cfg.CommandPrefix != "!" {
		t.Errorf("missing config prefix = %q, want !", cfg.CommandPrefix)
	}
}

func TestConfig_LoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "slack.json")
	os.WriteFile(path, []byte("not json"), 0644)

	cfg := LoadConfigFromFile(path)
	if cfg.Enabled {
		t.Error("invalid json should return default config")
	}
}

func TestConfig_EmptyPrefix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "slack.json")
	os.WriteFile(path, []byte(`{"command_prefix":""}`), 0644)

	cfg := LoadConfigFromFile(path)
	if cfg.CommandPrefix != "!" {
		t.Errorf("empty prefix should default to !, got %q", cfg.CommandPrefix)
	}
}

func TestConfigPath(t *testing.T) {
	p := ConfigPath()
	if !filepath.IsAbs(p) {
		t.Errorf("ConfigPath should be absolute, got %q", p)
	}
	if filepath.Base(p) != "slack.json" {
		t.Errorf("ConfigPath base = %q, want slack.json", filepath.Base(p))
	}
}
