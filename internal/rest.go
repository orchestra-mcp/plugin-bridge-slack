package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const slackAPI = "https://slack.com/api"

// RestClient communicates with the Slack Web API.
type RestClient struct {
	token  string
	client *http.Client
}

// NewRestClient creates a REST client for the Slack Web API.
func NewRestClient(token string) *RestClient {
	return &RestClient{
		token:  token,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendMessage sends a message to a Slack channel. Returns the message timestamp.
func (r *RestClient) SendMessage(channelID, text string, blocks []Block, attachments []Attachment) (string, error) {
	body := map[string]any{
		"channel": channelID,
	}
	if text != "" {
		body["text"] = text
	}
	if len(blocks) > 0 {
		body["blocks"] = blocks
	}
	if len(attachments) > 0 {
		body["attachments"] = attachments
	}
	respBody, err := r.do("chat.postMessage", body)
	if err != nil {
		return "", err
	}
	var msg struct {
		OK    bool   `json:"ok"`
		TS    string `json:"ts"`
		Error string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(respBody, &msg)
	if !msg.OK {
		return "", fmt.Errorf("chat.postMessage: %s", msg.Error)
	}
	return msg.TS, nil
}

// UpdateMessage updates an existing Slack message.
func (r *RestClient) UpdateMessage(channelID, ts, text string, blocks []Block) error {
	body := map[string]any{
		"channel": channelID,
		"ts":      ts,
	}
	if text != "" {
		body["text"] = text
	}
	if len(blocks) > 0 {
		body["blocks"] = blocks
	}
	respBody, err := r.do("chat.update", body)
	if err != nil {
		return err
	}
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(respBody, &resp)
	if !resp.OK {
		return fmt.Errorf("chat.update: %s", resp.Error)
	}
	return nil
}

// RespondToURL sends an ephemeral or in-channel response via response_url.
func (r *RestClient) RespondToURL(responseURL, text string, blocks []Block, replaceOriginal bool) error {
	body := map[string]any{}
	if text != "" {
		body["text"] = text
	}
	if len(blocks) > 0 {
		body["blocks"] = blocks
	}
	if replaceOriginal {
		body["replace_original"] = "true"
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", responseURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("response_url: %d %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (r *RestClient) do(method string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", slackAPI+"/"+method, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("slack API %s: %d %s", method, resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
