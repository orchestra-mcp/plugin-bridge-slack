package tools

import (
	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// SlackBridge provides access to the Slack bot for MCP tools.
type SlackBridge struct {
	Plugin *internal.BridgePlugin
}
