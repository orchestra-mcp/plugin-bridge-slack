package internal

import (
	"sync"
	"testing"
	"time"
)

// mockHandler implements Handler for testing.
type mockHandler struct {
	name     string
	prefixes []string
	slashes  []string
	called   bool
	mu       sync.Mutex
}

func (m *mockHandler) Name() string          { return m.name }
func (m *mockHandler) Description() string    { return "test handler" }
func (m *mockHandler) SlashCommand() string   { return "" }
func (m *mockHandler) MatchesPrefix(text string) bool {
	for _, p := range m.prefixes {
		if len(text) >= len(p) && text[:len(p)] == p {
			return true
		}
	}
	return false
}
func (m *mockHandler) MatchesSlash(cmd string) bool {
	for _, s := range m.slashes {
		if cmd == s {
			return true
		}
	}
	return false
}
func (m *mockHandler) HandleMessage(msg *MessageEvent, api HandlerAPI) {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
}
func (m *mockHandler) HandleSlashCommand(cmd *SlashCommandPayload, api HandlerAPI) {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
}
func (m *mockHandler) wasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

// mockAPI implements HandlerAPI for testing.
type mockAPI struct {
	cfg *Config
}

func (m *mockAPI) Config() *Config { return m.cfg }
func (m *mockAPI) SendToChannel(channelID, text string, blocks []Block, attachments []Attachment) error {
	return nil
}
func (m *mockAPI) SendBlocks(channelID string, blocks []Block, attachments []Attachment) (string, error) {
	return "123.456", nil
}
func (m *mockAPI) UpdateMessage(channelID, ts, text string, blocks []Block) error {
	return nil
}
func (m *mockAPI) RespondToURL(responseURL, text string, blocks []Block, replaceOriginal bool) error {
	return nil
}
func (m *mockAPI) CallTool(name string, args map[string]any) (string, error) {
	return "ok", nil
}
func (m *mockAPI) IsRunning() bool   { return true }
func (m *mockAPI) ChannelID() string { return "C123" }

func TestRouter_RouteMessage_MatchesHandler(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.Register(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteMessage(MessageEvent{Text: "!chat hello", User: "U123"}, api)

	time.Sleep(50 * time.Millisecond) // handler runs in goroutine
	if !h.wasCalled() {
		t.Error("expected chat handler to be called")
	}
}

func TestRouter_RouteMessage_IgnoresBot(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.Register(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteMessage(MessageEvent{Text: "!chat hello", BotID: "B999"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("bot messages should be ignored")
	}
}

func TestRouter_RouteMessage_RejectsUnallowed(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.Register(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{"U111"}}}
	r.RouteMessage(MessageEvent{Text: "!chat hello", User: "U999"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("disallowed user should not trigger handler")
	}
}

func TestRouter_RouteMessage_NoPrefix(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.Register(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteMessage(MessageEvent{Text: "hello world", User: "U123"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("messages without prefix should not trigger handler")
	}
}

func TestRouter_RouteMessage_DefaultHandler(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.SetDefault(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteMessage(MessageEvent{Text: "!unknown command", User: "U123"}, api)

	time.Sleep(50 * time.Millisecond)
	if !h.wasCalled() {
		t.Error("default handler should be called for unmatched prefix commands")
	}
}

func TestRouter_RouteSlashCommand(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", slashes: []string{"/chat"}}
	r.Register(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteSlashCommand(SlashCommandPayload{Command: "/chat", UserID: "U123", Text: "hello"}, api)

	time.Sleep(50 * time.Millisecond)
	if !h.wasCalled() {
		t.Error("expected chat handler to be called for /chat slash command")
	}
}

func TestRouter_RouteSlashCommand_RejectsUnallowed(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", slashes: []string{"/chat"}}
	r.Register(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{"U111"}}}
	r.RouteSlashCommand(SlashCommandPayload{Command: "/chat", UserID: "U999"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("disallowed user should not trigger slash handler")
	}
}

func TestRouter_DefaultPrefix(t *testing.T) {
	r := NewRouter("")
	if r.prefix != "!" {
		t.Errorf("empty prefix should default to !, got %q", r.prefix)
	}
}

func TestRouter_CustomPrefix(t *testing.T) {
	r := NewRouter("/")
	if r.prefix != "/" {
		t.Errorf("prefix = %q, want /", r.prefix)
	}
}

// ── RouteDirect tests (app_mention / DM routing) ──

func TestRouter_RouteDirect_CallsDefaultHandler(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.SetDefault(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteDirect(MessageEvent{Text: "hello world", User: "U123", Channel: "D456"}, api)

	time.Sleep(50 * time.Millisecond)
	if !h.wasCalled() {
		t.Error("RouteDirect should call default handler without prefix")
	}
}

func TestRouter_RouteDirect_StripsMention(t *testing.T) {
	r := NewRouter("!")
	var capturedText string
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	h2 := &textCapture{mockHandler: h, captured: &capturedText}
	r.SetDefault(h2)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteDirect(MessageEvent{Text: "<@U00BOT> what is Go?", User: "U123", Channel: "C789"}, api)

	time.Sleep(50 * time.Millisecond)
	if !h2.wasCalled() {
		t.Error("RouteDirect should call handler for @mention messages")
	}
	// The handler receives "chat what is Go?" after mention strip + "chat " prefix
	if capturedText != "chat what is Go?" {
		t.Errorf("expected 'chat what is Go?', got %q", capturedText)
	}
}

func TestRouter_RouteDirect_IgnoresBot(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.SetDefault(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteDirect(MessageEvent{Text: "hello", BotID: "B999", User: "U123"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("RouteDirect should ignore bot messages")
	}
}

func TestRouter_RouteDirect_RejectsUnallowed(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.SetDefault(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{"U111"}}}
	r.RouteDirect(MessageEvent{Text: "hello", User: "U999"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("RouteDirect should reject disallowed users")
	}
}

func TestRouter_RouteDirect_EmptyText(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.SetDefault(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteDirect(MessageEvent{Text: "", User: "U123"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("RouteDirect should not call handler for empty text")
	}
}

func TestRouter_RouteDirect_MentionOnly(t *testing.T) {
	r := NewRouter("!")
	h := &mockHandler{name: "chat", prefixes: []string{"chat"}}
	r.SetDefault(h)

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	r.RouteDirect(MessageEvent{Text: "<@U00BOT>", User: "U123"}, api)

	time.Sleep(50 * time.Millisecond)
	if h.wasCalled() {
		t.Error("RouteDirect should not call handler when message is just a mention with no text")
	}
}

func TestRouter_RouteDirect_NoDefaultHandler(t *testing.T) {
	r := NewRouter("!")
	// No default handler set

	api := &mockAPI{cfg: &Config{AllowedUsers: []string{}}}
	// Should not panic
	r.RouteDirect(MessageEvent{Text: "hello", User: "U123"}, api)
}

// textCapture wraps mockHandler to capture the message text passed to HandleMessage.
type textCapture struct {
	*mockHandler
	captured *string
}

func (tc *textCapture) HandleMessage(msg *MessageEvent, api HandlerAPI) {
	tc.mockHandler.HandleMessage(msg, api)
	tc.mu.Lock()
	*tc.captured = msg.Text
	tc.mu.Unlock()
}
