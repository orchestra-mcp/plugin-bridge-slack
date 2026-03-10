package handlers

import (
	"strings"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// ProgressHandler watches session progress.
type ProgressHandler struct{}

// NewProgressHandler creates a new progress handler.
func NewProgressHandler() *ProgressHandler { return &ProgressHandler{} }

func (h *ProgressHandler) Name() string                      { return "watch" }
func (h *ProgressHandler) MatchesPrefix(content string) bool { return strings.HasPrefix(strings.ToLower(content), "watch") }
func (h *ProgressHandler) MatchesSlash(command string) bool  { return command == "/watch" }
func (h *ProgressHandler) SlashCommand() string              { return "/watch" }

// HandleMessage handles a prefix command message.
func (h *ProgressHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	parts := strings.Fields(msg.Text)
	sessionID := ""
	if len(parts) > 1 {
		sessionID = parts[1]
	}
	h.doWatch(msg.Channel, sessionID, api)
}

// HandleSlashCommand handles a slash command.
func (h *ProgressHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	api.RespondToURL(cmd.ResponseURL, "Checking...", nil, false)
	h.doWatch(cmd.ChannelID, cmd.Text, api)
}

func (h *ProgressHandler) doWatch(channelID, sessionID string, api internal.HandlerAPI) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		result, err := api.CallTool("list_active", map[string]any{})
		if err != nil {
			blocks, attachments := internal.ErrorBlocks("Error", err.Error())
			api.SendToChannel(channelID, "Error", blocks, attachments)
			return
		}
		blocks, attachments := internal.InfoBlocks("Active Sessions", result+"\n\nUse `!watch <session_id>` to watch")
		api.SendToChannel(channelID, "Active Sessions", blocks, attachments)
		return
	}

	result, err := api.CallTool("session_status", map[string]any{"session_id": sessionID})
	if err != nil {
		blocks, attachments := internal.ErrorBlocks("Watch Error", err.Error())
		api.SendToChannel(channelID, "Watch Error", blocks, attachments)
		return
	}
	blocks, attachments := internal.InfoBlocks("Session Progress", result)
	api.SendToChannel(channelID, "Session Progress", blocks, attachments)
}
