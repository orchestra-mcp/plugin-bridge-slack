package handlers

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
)

var startTime = time.Now()

// PingHandler provides a health check command.
type PingHandler struct{}

// NewPingHandler creates a new ping handler.
func NewPingHandler() *PingHandler { return &PingHandler{} }

func (h *PingHandler) Name() string                      { return "ping" }
func (h *PingHandler) MatchesPrefix(content string) bool { return strings.ToLower(strings.TrimSpace(content)) == "ping" }
func (h *PingHandler) MatchesSlash(command string) bool  { return command == "/ping" }
func (h *PingHandler) SlashCommand() string              { return "/ping" }

// HandleMessage handles a prefix command message.
func (h *PingHandler) HandleMessage(msg *internal.MessageEvent, api internal.HandlerAPI) {
	h.doPing(msg.Channel, api)
}

// HandleSlashCommand handles a slash command.
func (h *PingHandler) HandleSlashCommand(cmd *internal.SlashCommandPayload, api internal.HandlerAPI) {
	api.RespondToURL(cmd.ResponseURL, "Pong!", nil, false)
	h.doPing(cmd.ChannelID, api)
}

func (h *PingHandler) doPing(channelID string, api internal.HandlerAPI) {
	uptime := time.Since(startTime).Round(time.Second)
	desc := fmt.Sprintf("*Uptime:* %s\n*Go:* %s\n*Platform:* %s/%s", uptime, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	blocks, attachments := internal.SuccessBlocks("Pong!", desc)
	api.SendToChannel(channelID, "Pong!", blocks, attachments)
}
