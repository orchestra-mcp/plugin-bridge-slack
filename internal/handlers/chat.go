package handlers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

// ChatHandler handles Claude chat via Slack with sticky channel-session mapping.
type ChatHandler struct {
	mu             sync.Mutex
	channelSession map[string]string // channelID -> session ID
}

// NewChatHandler creates a new chat handler.
func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		channelSession: make(map[string]string),
	}
}

func (h *ChatHandler) Name() string                      { return "chat" }
func (h *ChatHandler) MatchesPrefix(content string) bool { return strings.HasPrefix(strings.ToLower(content), "chat ") }
func (h *ChatHandler) MatchesSlash(command string) bool  { return command == "/chat" }
func (h *ChatHandler) SlashCommand() string              { return "/chat" }

// HandleMessage handles a prefix command message.
func (h *ChatHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	prompt := strings.TrimPrefix(msg.Text, "chat ")
	prompt = strings.TrimPrefix(prompt, "Chat ")
	if prompt == "" {
		blocks, attachments := internal.InfoBlocks("Usage", "`!chat <prompt>` or `!chat @workspace <prompt>`")
		api.SendToChannel(msg.Channel, "", blocks, attachments)
		return
	}
	h.doChat(msg.Channel, prompt, api)
}

// HandleSlashCommand handles a slash command.
func (h *ChatHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	prompt := cmd.Text
	if prompt == "" {
		api.RespondToURL(cmd.ResponseURL, "Please provide a prompt", nil, false)
		return
	}
	// Acknowledge immediately
	api.RespondToURL(cmd.ResponseURL, "Processing...", nil, false)
	h.doChat(cmd.ChannelID, prompt, api)
}

// parseWorkspace extracts @workspace-id from the beginning of a prompt.
// Returns (workspaceID, remainingPrompt). If no @workspace prefix, returns ("", original).
func parseWorkspace(prompt string) (string, string) {
	if !strings.HasPrefix(prompt, "@") {
		return "", prompt
	}
	parts := strings.SplitN(prompt, " ", 2)
	wsID := strings.TrimPrefix(parts[0], "@")
	if len(parts) < 2 {
		return wsID, ""
	}
	return wsID, parts[1]
}

func (h *ChatHandler) doChat(channelID, prompt string, api internal.HandlerAPI) {
	cfg := api.Config()

	// Check for workspace routing: "!chat @workspace-id what is the status?"
	wsID, remainingPrompt := parseWorkspace(prompt)

	// If no explicit workspace but default is configured, use it when API is available
	if wsID == "" && cfg.DefaultWorkspace != "" && cfg.APIURL != "" {
		wsID = cfg.DefaultWorkspace
		remainingPrompt = prompt
	}

	// Route through web server API if workspace is specified and API is configured
	if wsID != "" && cfg.APIURL != "" && cfg.APIToken != "" {
		if remainingPrompt == "" {
			blocks, attachments := internal.ErrorBlocks("Error", "Please provide a prompt after the workspace name")
			api.SendToChannel(channelID, "Error", blocks, attachments)
			return
		}

		blocks, attachments := internal.InfoBlocks("Processing", fmt.Sprintf("Workspace `%s`\n```\n%s\n```", wsID, internal.Truncate(remainingPrompt, 200)))
		api.SendToChannel(channelID, "Processing...", blocks, attachments)

		client := internal.NewWorkspaceClient(cfg.APIURL, cfg.APIToken)
		result, err := client.Chat(wsID, remainingPrompt)
		if err != nil {
			blocks, attachments = internal.ErrorBlocks("Error", err.Error())
			api.SendToChannel(channelID, "Error", blocks, attachments)
			return
		}

		h.sendResponse(channelID, result, api)
		return
	}

	// Fallback: local CallTool("ai_prompt", ...)
	blocks, attachments := internal.InfoBlocks("Processing", fmt.Sprintf("```\n%s\n```", internal.Truncate(prompt, 200)))
	api.SendToChannel(channelID, "Processing...", blocks, attachments)

	result, err := api.CallTool("ai_prompt", map[string]any{
		"prompt": prompt,
		"wait":   true,
	})
	if err != nil {
		blocks, attachments = internal.ErrorBlocks("Error", err.Error())
		api.SendToChannel(channelID, "Error", blocks, attachments)
		return
	}

	h.sendResponse(channelID, result, api)
}

func (h *ChatHandler) sendResponse(channelID, result string, api internal.HandlerAPI) {
	if len(result) <= internal.SafeBlockText {
		blocks, attachments := internal.SuccessBlocks("Response", result)
		api.SendToChannel(channelID, "Response", blocks, attachments)
		return
	}
	chunks := splitMessage(result, internal.SafeBlockText)
	for i, chunk := range chunks {
		title := fmt.Sprintf("Response (%d/%d)", i+1, len(chunks))
		blocks, attachments := internal.SuccessBlocks(title, chunk)
		api.SendToChannel(channelID, title, blocks, attachments)
	}
}

func splitMessage(s string, maxLen int) []string {
	var chunks []string
	for len(s) > maxLen {
		chunks = append(chunks, s[:maxLen])
		s = s[maxLen:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}
