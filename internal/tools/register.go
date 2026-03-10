package tools

import (
	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// RegisterAll registers all Slack bridge tools with the plugin builder.
func RegisterAll(builder *plugin.PluginBuilder, bp *internal.BridgePlugin) {
	bridge := &SlackBridge{Plugin: bp}

	builder.RegisterTool("start_slack_bot",
		"Start the Slack bot (connects to Socket Mode, registers commands)",
		StartBotSchema(), StartBot(bridge))

	builder.RegisterTool("stop_slack_bot",
		"Stop the Slack bot (disconnects from Socket Mode)",
		StopBotSchema(), StopBot(bridge))

	builder.RegisterTool("slack_bot_status",
		"Get Slack bot status (running, config, allowed users)",
		BotStatusSchema(), BotStatus(bridge))

	builder.RegisterTool("slack_send_message",
		"Send a message to a Slack channel",
		SendMessageSchema(), SendMessage(bridge))

	builder.RegisterTool("slack_set_config",
		"Update Slack bot configuration (saved to ~/.orchestra/slack.json)",
		SetConfigSchema(), SetConfig(bridge))
}
