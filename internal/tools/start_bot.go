package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// StartBotSchema returns the JSON schema for the start_slack_bot tool.
func StartBotSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

// StartBot returns a tool handler that starts the Slack bot.
func StartBot(bridge *SlackBridge) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if bridge.Plugin.Bot == nil {
			return helpers.ErrorResult("bot_error", "Bot not initialized"), nil
		}
		if bridge.Plugin.Bot.IsRunning() {
			return helpers.TextResult("Slack bot is already running"), nil
		}
		go bridge.Plugin.Bot.Start(ctx)
		return helpers.TextResult("Slack bot starting..."), nil
	}
}
