package handlers

import (
	"strings"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// PermissionHandler handles tool permission approve/deny via Slack buttons.
type PermissionHandler struct{}

// NewPermissionHandler creates a new permission handler.
func NewPermissionHandler() *PermissionHandler { return &PermissionHandler{} }

func (h *PermissionHandler) Name() string                                                            { return "permission" }
func (h *PermissionHandler) MatchesPrefix(_ string) bool                                             { return false }
func (h *PermissionHandler) MatchesSlash(_ string) bool                                              { return false }
func (h *PermissionHandler) HandleMessage(_ *internal.MessageEvent, _ internal.HandlerAPI)            {}
func (h *PermissionHandler) HandleSlashCommand(_ *internal.SlashCommandPayload, _ internal.HandlerAPI) {}
func (h *PermissionHandler) SlashCommand() string                                                    { return "" }

// MatchesActionID matches permission button interactions.
func (h *PermissionHandler) MatchesActionID(actionID string) bool {
	return strings.HasPrefix(actionID, "perm_approve_") || strings.HasPrefix(actionID, "perm_deny_")
}

// HandleInteraction handles permission button clicks.
func (h *PermissionHandler) HandleInteraction(payload *internal.InteractionPayload, api internal.HandlerAPI) {
	if len(payload.Actions) == 0 {
		return
	}
	actionID := payload.Actions[0].ActionID
	var decision, reqID string

	if strings.HasPrefix(actionID, "perm_approve_") {
		decision = "approve"
		reqID = strings.TrimPrefix(actionID, "perm_approve_")
	} else if strings.HasPrefix(actionID, "perm_deny_") {
		decision = "deny"
		reqID = strings.TrimPrefix(actionID, "perm_deny_")
	}

	_, err := api.CallTool("respond_permission", map[string]any{
		"id":       reqID,
		"decision": decision,
	})
	if err != nil {
		blocks, _ := internal.ErrorBlocks("Permission Error", err.Error())
		api.RespondToURL(payload.ResponseURL, "Permission Error", blocks, true)
		return
	}

	status := "Approved"
	if decision == "deny" {
		status = "Denied"
	}
	blocks, _ := internal.PermissionResultBlocks(status, reqID)
	api.RespondToURL(payload.ResponseURL, "Permission "+status, blocks, true)
}
