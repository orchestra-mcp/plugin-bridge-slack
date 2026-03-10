package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// McpHandler executes MCP tools from Slack.
type McpHandler struct{}

// NewMcpHandler creates a new MCP handler.
func NewMcpHandler() *McpHandler { return &McpHandler{} }

func (h *McpHandler) Name() string                      { return "mcp" }
func (h *McpHandler) MatchesPrefix(content string) bool { return strings.HasPrefix(strings.ToLower(content), "mcp ") }
func (h *McpHandler) MatchesSlash(command string) bool  { return command == "/mcp" }
func (h *McpHandler) SlashCommand() string              { return "/mcp" }

// HandleMessage handles a prefix command message.
func (h *McpHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	parts := strings.SplitN(strings.TrimPrefix(msg.Text, "mcp "), " ", 2)
	toolName := parts[0]
	var argsJSON string
	if len(parts) > 1 {
		argsJSON = parts[1]
	}
	h.doMcp(msg.Channel, toolName, argsJSON, api)
}

// HandleSlashCommand handles a slash command.
func (h *McpHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	parts := strings.SplitN(cmd.Text, " ", 2)
	toolName := ""
	var argsJSON string
	if len(parts) > 0 {
		toolName = parts[0]
	}
	if len(parts) > 1 {
		argsJSON = parts[1]
	}
	api.RespondToURL(cmd.ResponseURL, "Running...", nil, false)
	h.doMcp(cmd.ChannelID, toolName, argsJSON, api)
}

func (h *McpHandler) doMcp(channelID, toolName, argsJSON string, api internal.HandlerAPI) {
	if toolName == "" {
		blocks, attachments := internal.InfoBlocks("Usage", "`!mcp <tool> [json-args]`")
		api.SendToChannel(channelID, "Usage", blocks, attachments)
		return
	}

	args := make(map[string]any)
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			blocks, attachments := internal.ErrorBlocks("Invalid JSON", err.Error())
			api.SendToChannel(channelID, "Invalid JSON", blocks, attachments)
			return
		}
	}

	blocks, attachments := internal.InfoBlocks("Running", fmt.Sprintf("`%s`", toolName))
	api.SendToChannel(channelID, "Running "+toolName, blocks, attachments)

	result, err := api.CallTool(toolName, args)
	if err != nil {
		blocks, attachments = internal.ErrorBlocks("Tool Error", err.Error())
		api.SendToChannel(channelID, "Tool Error", blocks, attachments)
		return
	}

	if len(result) <= internal.SafeBlockText {
		blocks, attachments = internal.SuccessBlocks(toolName, result)
		api.SendToChannel(channelID, toolName, blocks, attachments)
	} else {
		blocks, attachments = internal.SuccessBlocks(toolName, internal.Truncate(result, internal.SafeBlockText))
		api.SendToChannel(channelID, toolName, blocks, attachments)
	}
}
