package internal

import (
	"encoding/json"
	"testing"
)

func TestSocketModeEnvelope_Unmarshal(t *testing.T) {
	raw := `{"envelope_id":"abc123","type":"events_api","payload":{"event":{"type":"message"}},"accepts_response_payload":false}`
	var env SocketModeEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatal(err)
	}
	if env.EnvelopeID != "abc123" {
		t.Errorf("envelope_id = %q, want abc123", env.EnvelopeID)
	}
	if env.Type != "events_api" {
		t.Errorf("type = %q, want events_api", env.Type)
	}
	if env.Payload == nil {
		t.Error("payload should not be nil")
	}
}

func TestSocketModeEnvelope_SlashCommand(t *testing.T) {
	raw := `{"envelope_id":"def456","type":"slash_commands","payload":{"command":"/chat","text":"hello"}}`
	var env SocketModeEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatal(err)
	}
	if env.Type != "slash_commands" {
		t.Errorf("type = %q, want slash_commands", env.Type)
	}
}

func TestSocketModeEnvelope_Interactive(t *testing.T) {
	raw := `{"envelope_id":"ghi789","type":"interactive","payload":{"type":"block_actions","actions":[{"action_id":"perm_approve_123"}]}}`
	var env SocketModeEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatal(err)
	}
	if env.Type != "interactive" {
		t.Errorf("type = %q, want interactive", env.Type)
	}
}

func TestMessageEvent_Unmarshal(t *testing.T) {
	raw := `{"type":"message","channel":"C12345","user":"U67890","text":"!chat hello","ts":"1234567890.123456","bot_id":""}`
	var msg MessageEvent
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Channel != "C12345" {
		t.Errorf("channel = %q, want C12345", msg.Channel)
	}
	if msg.User != "U67890" {
		t.Errorf("user = %q, want U67890", msg.User)
	}
	if msg.Text != "!chat hello" {
		t.Errorf("text = %q, want '!chat hello'", msg.Text)
	}
	if msg.BotID != "" {
		t.Error("bot_id should be empty")
	}
}

func TestMessageEvent_BotMessage(t *testing.T) {
	raw := `{"type":"message","channel":"C123","user":"","text":"bot message","ts":"111.222","bot_id":"B999"}`
	var msg MessageEvent
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.BotID != "B999" {
		t.Errorf("bot_id = %q, want B999", msg.BotID)
	}
}

func TestSlashCommandPayload_Unmarshal(t *testing.T) {
	raw := `{"command":"/chat","text":"tell me a joke","response_url":"https://hooks.slack.com/commands/T123/456/abc","trigger_id":"111.222","user_id":"U999","user_name":"testuser","channel_id":"C555","channel_name":"general","team_id":"T123"}`
	var cmd SlashCommandPayload
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.Command != "/chat" {
		t.Errorf("command = %q, want /chat", cmd.Command)
	}
	if cmd.Text != "tell me a joke" {
		t.Errorf("text = %q, want 'tell me a joke'", cmd.Text)
	}
	if cmd.UserID != "U999" {
		t.Errorf("user_id = %q, want U999", cmd.UserID)
	}
	if cmd.ChannelID != "C555" {
		t.Errorf("channel_id = %q, want C555", cmd.ChannelID)
	}
}

func TestInteractionPayload_Unmarshal(t *testing.T) {
	raw := `{"type":"block_actions","trigger_id":"111.222","response_url":"https://hooks.slack.com/actions/T/B/X","user":{"id":"U999","username":"testuser","name":"Test User","team_id":"T123"},"channel":{"id":"C555","name":"general"},"actions":[{"action_id":"perm_approve_REQ123","block_id":"perm_actions_REQ123","type":"button","value":"REQ123"}]}`
	var payload InteractionPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Type != "block_actions" {
		t.Errorf("type = %q, want block_actions", payload.Type)
	}
	if payload.User.ID != "U999" {
		t.Errorf("user.id = %q, want U999", payload.User.ID)
	}
	if payload.User.Username != "testuser" {
		t.Errorf("user.username = %q, want testuser", payload.User.Username)
	}
	if len(payload.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(payload.Actions))
	}
	if payload.Actions[0].ActionID != "perm_approve_REQ123" {
		t.Errorf("action_id = %q, want perm_approve_REQ123", payload.Actions[0].ActionID)
	}
	if payload.Actions[0].Value != "REQ123" {
		t.Errorf("value = %q, want REQ123", payload.Actions[0].Value)
	}
}

func TestBlockText_Mrkdwn(t *testing.T) {
	bt := Mrkdwn("*bold* text")
	if bt.Type != "mrkdwn" {
		t.Errorf("type = %q, want mrkdwn", bt.Type)
	}
	if bt.Text != "*bold* text" {
		t.Errorf("text = %q, want '*bold* text'", bt.Text)
	}
}

func TestBlockText_PlainText(t *testing.T) {
	bt := PlainText("hello world")
	if bt.Type != "plain_text" {
		t.Errorf("type = %q, want plain_text", bt.Type)
	}
	if !bt.Emoji {
		t.Error("plain_text should have emoji=true")
	}
}

func TestSectionBlock(t *testing.T) {
	b := SectionBlock("some text")
	if b.Type != "section" {
		t.Errorf("type = %q, want section", b.Type)
	}
	if b.Text == nil || b.Text.Text != "some text" {
		t.Error("section text not set correctly")
	}
}

func TestHeaderBlock(t *testing.T) {
	b := HeaderBlock("My Header")
	if b.Type != "header" {
		t.Errorf("type = %q, want header", b.Type)
	}
	if b.Text == nil || b.Text.Text != "My Header" {
		t.Error("header text not set correctly")
	}
	if b.Text.Type != "plain_text" {
		t.Errorf("header text type = %q, want plain_text", b.Text.Type)
	}
}

func TestDividerBlock(t *testing.T) {
	b := DividerBlock()
	if b.Type != "divider" {
		t.Errorf("type = %q, want divider", b.Type)
	}
}

func TestButtonElement(t *testing.T) {
	btn := ButtonElement("Click Me", "my_action", "my_value", "primary")
	if btn.Type != "button" {
		t.Errorf("type = %q, want button", btn.Type)
	}
	if btn.ActionID != "my_action" {
		t.Errorf("action_id = %q, want my_action", btn.ActionID)
	}
	if btn.Value != "my_value" {
		t.Errorf("value = %q, want my_value", btn.Value)
	}
	if btn.Style != "primary" {
		t.Errorf("style = %q, want primary", btn.Style)
	}
	if btn.Text == nil || btn.Text.Text != "Click Me" {
		t.Error("button text not set correctly")
	}
}

func TestActionsBlock(t *testing.T) {
	btn := ButtonElement("OK", "ok_btn", "ok", "")
	b := ActionsBlock("actions_1", btn)
	if b.Type != "actions" {
		t.Errorf("type = %q, want actions", b.Type)
	}
	if b.BlockID != "actions_1" {
		t.Errorf("block_id = %q, want actions_1", b.BlockID)
	}
	if len(b.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(b.Elements))
	}
}

func TestContextBlock(t *testing.T) {
	b := ContextBlock("item 1", "item 2")
	if b.Type != "context" {
		t.Errorf("type = %q, want context", b.Type)
	}
	if len(b.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(b.Elements))
	}
}
