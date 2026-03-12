package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WorkspaceClient calls the Orchestra web server's repo workspace chat API.
type WorkspaceClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewWorkspaceClient creates a client for the given API URL and token.
func NewWorkspaceClient(apiURL, apiToken string) *WorkspaceClient {
	return &WorkspaceClient{
		baseURL: strings.TrimRight(apiURL, "/"),
		token:   apiToken,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// ChatRequest is the payload for POST /api/repos/:id/chat.
type ChatRequest struct {
	Prompt string `json:"prompt"`
}

// ChatResponse is the response from POST /api/repos/:id/chat.
type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Chat sends a prompt to a workspace and returns the response.
func (c *WorkspaceClient) Chat(workspaceID, prompt string) (string, error) {
	body, _ := json.Marshal(ChatRequest{Prompt: prompt})

	url := fmt.Sprintf("%s/api/repos/%s/chat", c.baseURL, workspaceID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("workspace chat: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("workspace chat %d: %s", resp.StatusCode, string(data))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if chatResp.Error != "" {
		return "", fmt.Errorf("workspace: %s", chatResp.Error)
	}
	return chatResp.Response, nil
}
