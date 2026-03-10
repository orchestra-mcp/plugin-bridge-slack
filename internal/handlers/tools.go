package handlers

import (
	"strings"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// ToolsHandler lists available MCP tools and commands.
type ToolsHandler struct{}

// NewToolsHandler creates a new tools handler.
func NewToolsHandler() *ToolsHandler { return &ToolsHandler{} }

func (h *ToolsHandler) Name() string                      { return "tools" }
func (h *ToolsHandler) MatchesPrefix(content string) bool { return strings.ToLower(strings.TrimSpace(content)) == "tools" }
func (h *ToolsHandler) MatchesSlash(command string) bool  { return command == "/tools" }
func (h *ToolsHandler) SlashCommand() string              { return "/tools" }

// HandleMessage handles a prefix command message.
func (h *ToolsHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	h.doTools(msg.Channel, api)
}

// HandleSlashCommand handles a slash command.
func (h *ToolsHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	api.RespondToURL(cmd.ResponseURL, "Loading tools...", nil, false)
	h.doTools(cmd.ChannelID, api)
}

func (h *ToolsHandler) doTools(channelID string, api internal.HandlerAPI) {
	// List features from the MCP
	result, err := api.CallTool("list_features", map[string]any{})
	if err != nil {
		// Fallback: just show available commands
		desc := "*Available Commands:*\n"
		desc += "`!chat <prompt>` - Chat with Claude\n"
		desc += "`!mcp <tool> [args]` - Execute MCP tool\n"
		desc += "`!status [project]` - Project status\n"
		desc += "`!stop [session]` - Stop session\n"
		desc += "`!ping` - Health check\n"
		desc += "`!tools` - This help\n"
		blocks, attachments := internal.InfoBlocks("Available Commands", desc)
		api.SendToChannel(channelID, "Available Commands", blocks, attachments)
		return
	}

	blocks, attachments := internal.InfoBlocks("MCP Tools", internal.Truncate(result, internal.SafeBlockText))
	api.SendToChannel(channelID, "MCP Tools", blocks, attachments)
}
