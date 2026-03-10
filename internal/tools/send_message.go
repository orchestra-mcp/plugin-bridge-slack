package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SendMessageSchema returns the JSON schema for the slack_send_message tool.
func SendMessageSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"channel_id": map[string]any{
				"type":        "string",
				"description": "Slack channel ID to send to (defaults to configured channel)",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Message content",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Header title (optional, sends as Block Kit message if provided)",
			},
			"color": map[string]any{
				"type":        "string",
				"description": "Attachment color: success, error, info, warning (default: info)",
			},
		},
		"required": []any{"content"},
	})
	return s
}

// SendMessage returns a tool handler that sends a message to a Slack channel.
func SendMessage(bridge *SlackBridge) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		if bridge.Plugin.Bot == nil || !bridge.Plugin.Bot.IsRunning() {
			return helpers.ErrorResult("bot_error", "Slack bot is not running"), nil
		}

		content := helpers.GetString(req.Arguments, "content")
		channelID := helpers.GetString(req.Arguments, "channel_id")
		title := helpers.GetString(req.Arguments, "title")
		color := helpers.GetString(req.Arguments, "color")

		if channelID == "" {
			channelID = bridge.Plugin.Bot.Config().ChannelID
		}
		if channelID == "" {
			return helpers.ErrorResult("config_error", "No channel ID configured"), nil
		}

		if title != "" {
			var blocks []internal.Block
			var attachments []internal.Attachment
			switch color {
			case "success":
				blocks, attachments = internal.SuccessBlocks(title, content)
			case "error":
				blocks, attachments = internal.ErrorBlocks(title, content)
			case "warning":
				blocks, attachments = internal.WarningBlocks(title, content)
			default:
				blocks, attachments = internal.InfoBlocks(title, content)
			}
			bridge.Plugin.Bot.SendToChannel(channelID, title, blocks, attachments)
		} else {
			bridge.Plugin.Bot.SendToChannel(channelID, content, nil, nil)
		}

		return helpers.TextResult(fmt.Sprintf("Message sent to channel %s", channelID)), nil
	}
}
