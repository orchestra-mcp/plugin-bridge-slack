package internal

import "encoding/json"

// SocketModeEnvelope represents a Slack Socket Mode envelope.
// All events arrive wrapped in this envelope and must be acknowledged.
type SocketModeEnvelope struct {
	EnvelopeID string          `json:"envelope_id"`
	Type       string          `json:"type"`       // events_api, slash_commands, interactive
	Payload    json.RawMessage `json:"payload"`
	AcceptsResponsePayload bool `json:"accepts_response_payload,omitempty"`
	RetryAttempt int          `json:"retry_attempt,omitempty"`
	RetryReason  string       `json:"retry_reason,omitempty"`
}

// EventCallback represents a Slack Events API callback.
type EventCallback struct {
	Type      string          `json:"type"` // event_callback
	Token     string          `json:"token"`
	TeamID    string          `json:"team_id"`
	Event     json.RawMessage `json:"event"`
	EventID   string          `json:"event_id"`
	EventTime int64           `json:"event_time"`
}

// MessageEvent represents a Slack message event.
type MessageEvent struct {
	Type      string `json:"type"`
	SubType   string `json:"subtype,omitempty"`
	Channel   string `json:"channel"`
	User      string `json:"user"`
	Text      string `json:"text"`
	TS        string `json:"ts"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	BotID     string `json:"bot_id,omitempty"`
}

// SlashCommandPayload represents a Slack slash command.
type SlashCommandPayload struct {
	Command     string `json:"command"`
	Text        string `json:"text"`
	ResponseURL string `json:"response_url"`
	TriggerID   string `json:"trigger_id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	TeamID      string `json:"team_id"`
}

// InteractionPayload represents a Slack interactive component payload.
type InteractionPayload struct {
	Type        string          `json:"type"` // block_actions, message_action
	TriggerID   string          `json:"trigger_id"`
	ResponseURL string          `json:"response_url"`
	User        SlackUser       `json:"user"`
	Channel     SlackChannel    `json:"channel"`
	Actions     []BlockAction   `json:"actions"`
	Message     *SlackMessage   `json:"message,omitempty"`
	Container   json.RawMessage `json:"container,omitempty"`
}

// SlackUser represents a Slack user in an interaction.
type SlackUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	TeamID   string `json:"team_id"`
}

// SlackChannel represents a Slack channel reference.
type SlackChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BlockAction represents a single action from an interactive component.
type BlockAction struct {
	ActionID string `json:"action_id"`
	BlockID  string `json:"block_id"`
	Type     string `json:"type"` // button, static_select, etc.
	Value    string `json:"value,omitempty"`
	Text     *BlockText `json:"text,omitempty"`
}

// Block represents a Slack Block Kit block.
type Block struct {
	Type     string      `json:"type"`               // section, header, divider, context, actions
	Text     *BlockText  `json:"text,omitempty"`
	Fields   []BlockText `json:"fields,omitempty"`
	Elements []BlockElement `json:"elements,omitempty"`
	BlockID  string      `json:"block_id,omitempty"`
}

// BlockText represents text content in Block Kit.
type BlockText struct {
	Type  string `json:"type"`  // mrkdwn, plain_text
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// BlockElement represents an interactive element (button, etc.) in Block Kit.
type BlockElement struct {
	Type     string     `json:"type"`               // button, static_select
	Text     *BlockText `json:"text,omitempty"`
	ActionID string     `json:"action_id,omitempty"`
	Value    string     `json:"value,omitempty"`
	Style    string     `json:"style,omitempty"` // primary, danger
}

// Attachment represents a Slack message attachment (used for colored sidebars).
type Attachment struct {
	Color    string  `json:"color,omitempty"`
	Blocks   []Block `json:"blocks,omitempty"`
	Fallback string  `json:"fallback,omitempty"`
}

// SlackMessage represents a Slack message with blocks.
type SlackMessage struct {
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`       // Fallback text
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	TS          string       `json:"ts,omitempty"`
	ThreadTS    string       `json:"thread_ts,omitempty"`
}

// Mrkdwn creates a mrkdwn BlockText.
func Mrkdwn(text string) BlockText {
	return BlockText{Type: "mrkdwn", Text: text}
}

// PlainText creates a plain_text BlockText.
func PlainText(text string) BlockText {
	return BlockText{Type: "plain_text", Text: text, Emoji: true}
}

// SectionBlock creates a section block with mrkdwn text.
func SectionBlock(text string) Block {
	t := Mrkdwn(text)
	return Block{Type: "section", Text: &t}
}

// HeaderBlock creates a header block.
func HeaderBlock(text string) Block {
	t := PlainText(text)
	return Block{Type: "header", Text: &t}
}

// DividerBlock creates a divider block.
func DividerBlock() Block {
	return Block{Type: "divider"}
}

// ContextBlock creates a context block with mrkdwn elements.
func ContextBlock(texts ...string) Block {
	elements := make([]BlockElement, len(texts))
	for i, t := range texts {
		elements[i] = BlockElement{Type: "mrkdwn", Text: &BlockText{Type: "mrkdwn", Text: t}}
	}
	return Block{Type: "context", Elements: elements}
}

// ButtonElement creates a button element.
func ButtonElement(text, actionID, value, style string) BlockElement {
	t := PlainText(text)
	return BlockElement{
		Type:     "button",
		Text:     &t,
		ActionID: actionID,
		Value:    value,
		Style:    style,
	}
}

// ActionsBlock creates an actions block with button elements.
func ActionsBlock(blockID string, elements ...BlockElement) Block {
	return Block{Type: "actions", BlockID: blockID, Elements: elements}
}
