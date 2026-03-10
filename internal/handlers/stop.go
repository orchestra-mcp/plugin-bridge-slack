package handlers

import (
	"strings"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// StopHandler stops Claude sessions.
type StopHandler struct{}

// NewStopHandler creates a new stop handler.
func NewStopHandler() *StopHandler { return &StopHandler{} }

func (h *StopHandler) Name() string                      { return "stop" }
func (h *StopHandler) MatchesPrefix(content string) bool { return strings.HasPrefix(strings.ToLower(content), "stop") }
func (h *StopHandler) MatchesSlash(command string) bool  { return command == "/stop" }
func (h *StopHandler) SlashCommand() string              { return "/stop" }

// HandleMessage handles a prefix command message.
func (h *StopHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	parts := strings.Fields(msg.Text)
	sessionID := ""
	if len(parts) > 1 {
		sessionID = parts[1]
	}
	h.doStop(msg.Channel, sessionID, api)
}

// HandleSlashCommand handles a slash command.
func (h *StopHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	api.RespondToURL(cmd.ResponseURL, "Processing...", nil, false)
	h.doStop(cmd.ChannelID, strings.TrimSpace(cmd.Text), api)
}

func (h *StopHandler) doStop(channelID, sessionID string, api internal.HandlerAPI) {
	if sessionID == "" {
		// List active sessions first
		result, err := api.CallTool("list_active", map[string]any{})
		if err != nil {
			blocks, attachments := internal.ErrorBlocks("Error", err.Error())
			api.SendToChannel(channelID, "Error", blocks, attachments)
			return
		}
		blocks, attachments := internal.InfoBlocks("Active Sessions", result+"\n\nUse `!stop <session_id>` to stop a session")
		api.SendToChannel(channelID, "Active Sessions", blocks, attachments)
		return
	}

	result, err := api.CallTool("kill_session", map[string]any{"session_id": sessionID})
	if err != nil {
		blocks, attachments := internal.ErrorBlocks("Stop Error", err.Error())
		api.SendToChannel(channelID, "Stop Error", blocks, attachments)
		return
	}
	blocks, attachments := internal.SuccessBlocks("Session Stopped", result)
	api.SendToChannel(channelID, "Session Stopped", blocks, attachments)
}
