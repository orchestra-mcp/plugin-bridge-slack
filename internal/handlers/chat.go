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
		blocks, attachments := internal.InfoBlocks("Usage", "`!chat <prompt>`")
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

func (h *ChatHandler) doChat(channelID, prompt string, api internal.HandlerAPI) {
	blocks, attachments := internal.InfoBlocks("Processing", fmt.Sprintf("```\n%s\n```", internal.Truncate(prompt, 200)))
	api.SendToChannel(channelID, "Processing...", blocks, attachments)

	// Call ai_prompt tool via cross-plugin bridge
	result, err := api.CallTool("ai_prompt", map[string]any{
		"prompt": prompt,
		"wait":   true,
	})
	if err != nil {
		blocks, attachments = internal.ErrorBlocks("Error", err.Error())
		api.SendToChannel(channelID, "Error", blocks, attachments)
		return
	}

	// Split long responses for Slack's block text limit
	if len(result) <= internal.SafeBlockText {
		blocks, attachments = internal.SuccessBlocks("Response", result)
		api.SendToChannel(channelID, "Response", blocks, attachments)
		return
	}
	chunks := splitMessage(result, internal.SafeBlockText)
	for i, chunk := range chunks {
		title := fmt.Sprintf("Response (%d/%d)", i+1, len(chunks))
		blocks, attachments = internal.SuccessBlocks(title, chunk)
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
