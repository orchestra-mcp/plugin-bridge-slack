package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SetConfigSchema returns the JSON schema for the slack_set_config tool.
func SetConfigSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"enabled":        map[string]any{"type": "boolean", "description": "Enable/disable bot"},
			"bot_token":      map[string]any{"type": "string", "description": "Slack bot token (xoxb-...)"},
			"app_token":      map[string]any{"type": "string", "description": "Slack app-level token (xapp-...)"},
			"signing_secret": map[string]any{"type": "string", "description": "Slack signing secret"},
			"app_id":         map[string]any{"type": "string", "description": "Slack app ID"},
			"channel_id":     map[string]any{"type": "string", "description": "Default channel ID"},
			"command_prefix": map[string]any{"type": "string", "description": "Command prefix (default: !)"},
			"webhook_url":    map[string]any{"type": "string", "description": "Webhook URL for notifications"},
			"allowed_users":  map[string]any{"type": "string", "description": "Comma-separated Slack user IDs"},
			"team_id":        map[string]any{"type": "string", "description": "Slack workspace team ID"},
		},
	})
	return s
}

// SetConfig returns a tool handler that updates Slack bot configuration.
func SetConfig(bridge *SlackBridge) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		cfg := internal.LoadConfig()

		if v := helpers.GetString(req.Arguments, "bot_token"); v != "" {
			cfg.BotToken = v
		}
		if v := helpers.GetString(req.Arguments, "app_token"); v != "" {
			cfg.AppToken = v
		}
		if v := helpers.GetString(req.Arguments, "signing_secret"); v != "" {
			cfg.SigningSecret = v
		}
		if v := helpers.GetString(req.Arguments, "app_id"); v != "" {
			cfg.AppID = v
		}
		if v := helpers.GetString(req.Arguments, "channel_id"); v != "" {
			cfg.ChannelID = v
		}
		if v := helpers.GetString(req.Arguments, "command_prefix"); v != "" {
			cfg.CommandPrefix = v
		}
		if v := helpers.GetString(req.Arguments, "webhook_url"); v != "" {
			cfg.WebhookURL = v
		}
		if v := helpers.GetString(req.Arguments, "team_id"); v != "" {
			cfg.TeamID = v
		}

		if req.Arguments != nil {
			if v, ok := req.Arguments.Fields["enabled"]; ok {
				cfg.Enabled = v.GetBoolValue()
			}
		}

		if v := helpers.GetString(req.Arguments, "allowed_users"); v != "" {
			var users []string
			for _, u := range splitCSV(v) {
				if u != "" {
					users = append(users, u)
				}
			}
			cfg.AllowedUsers = users
		}

		if err := cfg.Save(); err != nil {
			return helpers.ErrorResult("save_error", err.Error()), nil
		}

		return helpers.TextResult("Slack configuration saved. Restart the bot for changes to take effect."), nil
	}
}

func splitCSV(s string) []string {
	var result []string
	for _, part := range splitBy(s, ',') {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitBy(s string, sep byte) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
