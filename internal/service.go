package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NotificationService sends Slack notifications for workflow transitions.
type NotificationService struct {
	config *Config
	client *http.Client
}

// NewNotificationService creates a Slack notification service.
func NewNotificationService(cfg *Config) *NotificationService {
	if cfg == nil || !cfg.Enabled {
		return nil
	}
	if cfg.WebhookURL == "" && (cfg.BotToken == "" || cfg.ChannelID == "") {
		return nil
	}
	return &NotificationService{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendTransition sends a workflow transition notification to Slack.
func (s *NotificationService) SendTransition(taskID, taskType, from, to, project string) {
	if s == nil {
		return
	}
	emoji := statusEmoji(to)
	title := fmt.Sprintf("%s %s -> %s", emoji, taskID, to)
	desc := fmt.Sprintf("%s *%s* `%s` -> `%s`", emoji, taskID, from, to)

	blocks := []Block{
		SectionBlock(fmt.Sprintf("*%s*", title)),
		SectionBlock(desc),
		ContextBlock(
			fmt.Sprintf("*Task:* `%s` %s", taskID, taskType),
			fmt.Sprintf("*Project:* %s", project),
		),
	}

	go s.send(blocks)
}

func (s *NotificationService) send(blocks []Block) {
	body := map[string]any{
		"blocks": blocks,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return
	}

	if s.config.WebhookURL != "" {
		req, _ := http.NewRequest("POST", s.config.WebhookURL, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := s.client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	// Use chat.postMessage with bot token
	body["channel"] = s.config.ChannelID
	data, _ = json.Marshal(body)
	req, _ := http.NewRequest("POST", slackAPI+"/chat.postMessage", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+s.config.BotToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, _ := s.client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
}

func statusEmoji(status string) string {
	switch status {
	case "todo":
		return ":clipboard:"
	case "in-progress":
		return ":hammer:"
	case "in-testing":
		return ":microscope:"
	case "in-docs":
		return ":writing_hand:"
	case "in-review":
		return ":eyes:"
	case "done":
		return ":white_check_mark:"
	case "blocked":
		return ":no_entry_sign:"
	case "rejected", "needs-edits":
		return ":x:"
	default:
		return ":arrows_counterclockwise:"
	}
}
