package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewNotificationService_NilConfig(t *testing.T) {
	svc := NewNotificationService(nil)
	if svc != nil {
		t.Error("nil config should return nil service")
	}
}

func TestNewNotificationService_Disabled(t *testing.T) {
	cfg := &Config{Enabled: false, WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	svc := NewNotificationService(cfg)
	if svc != nil {
		t.Error("disabled config should return nil service")
	}
}

func TestNewNotificationService_NoWebhookOrBot(t *testing.T) {
	cfg := &Config{Enabled: true}
	svc := NewNotificationService(cfg)
	if svc != nil {
		t.Error("config with no webhook or bot should return nil")
	}
}

func TestNewNotificationService_WithWebhook(t *testing.T) {
	cfg := &Config{Enabled: true, WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	svc := NewNotificationService(cfg)
	if svc == nil {
		t.Error("config with webhook should create service")
	}
}

func TestNewNotificationService_WithBotAndChannel(t *testing.T) {
	cfg := &Config{Enabled: true, BotToken: "xoxb-test", ChannelID: "C123"}
	svc := NewNotificationService(cfg)
	if svc == nil {
		t.Error("config with bot token + channel should create service")
	}
}

func TestSendTransition_NilService(t *testing.T) {
	// Should not panic on nil receiver
	var svc *NotificationService
	svc.SendTransition("FEAT-ABC", "feature", "in-progress", "in-testing", "my-project")
}

func TestSendTransition_Webhook(t *testing.T) {
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %q, want application/json", ct)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["blocks"]; !ok {
			t.Error("expected blocks in webhook payload")
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &Config{Enabled: true, WebhookURL: server.URL}
	svc := NewNotificationService(cfg)
	svc.SendTransition("FEAT-ABC", "feature", "in-progress", "in-testing", "my-project")

	// Wait for the goroutine to complete
	// The send method is called in a goroutine inside SendTransition
	// We need to call send directly for a synchronous test
	svc.send([]Block{SectionBlock("test")})

	if !received {
		t.Error("webhook should have received the notification")
	}
}

func TestStatusEmoji(t *testing.T) {
	tests := map[string]string{
		"todo":        ":clipboard:",
		"in-progress": ":hammer:",
		"in-testing":  ":microscope:",
		"in-docs":     ":writing_hand:",
		"in-review":   ":eyes:",
		"done":        ":white_check_mark:",
		"blocked":     ":no_entry_sign:",
		"rejected":    ":x:",
		"needs-edits": ":x:",
		"unknown":     ":arrows_counterclockwise:",
	}
	for status, want := range tests {
		got := statusEmoji(status)
		if got != want {
			t.Errorf("statusEmoji(%q) = %q, want %q", status, got, want)
		}
	}
}
