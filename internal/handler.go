package internal

// HandlerAPI provides capabilities to handlers for sending messages
// and accessing shared services.
type HandlerAPI interface {
	SendToChannel(channelID, text string, blocks []Block, attachments []Attachment) error
	SendBlocks(channelID string, blocks []Block, attachments []Attachment) (string, error)
	UpdateMessage(channelID, ts, text string, blocks []Block) error
	RespondToURL(responseURL, text string, blocks []Block, replaceOriginal bool) error
	IsRunning() bool
	ChannelID() string
	Config() *Config
	// CallTool invokes an MCP tool by name via cross-plugin call.
	CallTool(name string, args map[string]any) (string, error)
}

// Handler processes Slack commands (both prefix messages and slash commands).
type Handler interface {
	Name() string
	MatchesPrefix(content string) bool
	MatchesSlash(command string) bool
	HandleMessage(msg *MessageEvent, api HandlerAPI)
	HandleSlashCommand(cmd *SlashCommandPayload, api HandlerAPI)
	SlashCommand() string // returns slash command name, e.g. "/chat"
}

// InteractionHandler is an optional interface for handlers that process
// button/component interactions (block_actions).
type InteractionHandler interface {
	MatchesActionID(actionID string) bool
	HandleInteraction(payload *InteractionPayload, api HandlerAPI)
}
