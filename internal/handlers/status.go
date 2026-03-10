package handlers

import (
	"strings"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// StatusHandler shows project workflow status.
type StatusHandler struct{}

// NewStatusHandler creates a new status handler.
func NewStatusHandler() *StatusHandler { return &StatusHandler{} }

func (h *StatusHandler) Name() string                      { return "status" }
func (h *StatusHandler) MatchesPrefix(content string) bool { return strings.HasPrefix(strings.ToLower(content), "status") }
func (h *StatusHandler) MatchesSlash(command string) bool  { return command == "/status" }
func (h *StatusHandler) SlashCommand() string              { return "/status" }

// HandleMessage handles a prefix command message.
func (h *StatusHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	parts := strings.Fields(msg.Text)
	projectID := ""
	if len(parts) > 1 {
		projectID = parts[1]
	}
	h.doStatus(msg.Channel, projectID, api)
}

// HandleSlashCommand handles a slash command.
func (h *StatusHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	api.RespondToURL(cmd.ResponseURL, "Checking status...", nil, false)
	h.doStatus(cmd.ChannelID, strings.TrimSpace(cmd.Text), api)
}

func (h *StatusHandler) doStatus(channelID, projectID string, api internal.HandlerAPI) {
	args := map[string]any{}
	if projectID != "" {
		args["project_id"] = projectID
	}

	result, err := api.CallTool("get_project_status", args)
	if err != nil {
		// Try get_progress as fallback
		result, err = api.CallTool("get_progress", args)
		if err != nil {
			blocks, attachments := internal.ErrorBlocks("Status Error", err.Error())
			api.SendToChannel(channelID, "Status Error", blocks, attachments)
			return
		}
	}

	blocks, attachments := internal.InfoBlocks("Project Status", result)
	api.SendToChannel(channelID, "Project Status", blocks, attachments)
}
