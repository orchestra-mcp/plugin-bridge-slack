package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// BotStatusSchema returns the JSON schema for the slack_bot_status tool.
func BotStatusSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

// BotStatus returns a tool handler that reports Slack bot status.
func BotStatus(bridge *SlackBridge) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if bridge.Plugin.Bot == nil {
			return helpers.TextResult("Slack bot: not initialized"), nil
		}
		running := bridge.Plugin.Bot.IsRunning()
		status := "stopped"
		if running {
			status = "running"
		}
		cfg := bridge.Plugin.Bot.Config()
		result := fmt.Sprintf("## Slack Bot Status\n\n- **Status:** %s\n- **Enabled:** %v\n- **Team:** %s\n- **Channel:** %s\n- **Prefix:** %s\n- **Allowed Users:** %d",
			status, cfg.Enabled, cfg.TeamID, cfg.ChannelID, cfg.CommandPrefix, len(cfg.AllowedUsers))
		return helpers.TextResult(result), nil
	}
}
