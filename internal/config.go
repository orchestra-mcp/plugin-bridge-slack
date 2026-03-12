package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds Slack bot configuration.
type Config struct {
	Enabled          bool     `json:"enabled"`
	BotToken         string   `json:"bot_token"`        // xoxb- bot token
	AppToken         string   `json:"app_token"`         // xapp- app-level token for Socket Mode
	SigningSecret    string   `json:"signing_secret"`
	AppID            string   `json:"app_id"`
	ChannelID        string   `json:"channel_id"`
	CommandPrefix    string   `json:"command_prefix"`
	WebhookURL       string   `json:"webhook_url"`
	AllowedUsers     []string `json:"allowed_users"`
	TeamID           string   `json:"team_id"`
	APIURL           string   `json:"api_url"`            // Orchestra web server URL (e.g. https://orchestra-mcp.dev)
	APIToken         string   `json:"api_token"`           // API auth token for web server
	DefaultWorkspace string   `json:"default_workspace"`   // Default workspace ID for chat routing
}

// DefaultConfig returns default Slack configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:       false,
		CommandPrefix: "!",
		AllowedUsers:  []string{},
	}
}

// IsAllowed checks if a Slack user ID is in the allowed list.
// If AllowedUsers is empty, all users are allowed.
func (c *Config) IsAllowed(userID string) bool {
	if len(c.AllowedUsers) == 0 {
		return true
	}
	for _, id := range c.AllowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}

// IsValid checks if minimum required fields are present.
func (c *Config) IsValid() bool {
	if c.BotToken != "" && c.AppToken != "" {
		return true
	}
	if c.WebhookURL != "" {
		return true
	}
	return false
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".orchestra", "slack.json")
}

// LoadConfig loads config from the default path.
func LoadConfig() *Config {
	return LoadConfigFromFile(ConfigPath())
}

// LoadConfigFromFile loads config from a specific path.
func LoadConfigFromFile(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig()
	}
	if cfg.CommandPrefix == "" {
		cfg.CommandPrefix = "!"
	}
	return cfg
}

// Save writes config to the default path.
func (c *Config) Save() error {
	return c.SaveToFile(ConfigPath())
}

// SaveToFile writes config to a specific path.
func (c *Config) SaveToFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
