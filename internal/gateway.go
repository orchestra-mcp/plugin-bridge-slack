package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Gateway maintains a Socket Mode WebSocket connection to Slack.
type Gateway struct {
	appToken  string
	conn      *websocket.Conn
	done      chan struct{}
	mu        sync.Mutex
	onEvent   SocketModeCallback
}

// SocketModeCallback is called by the gateway for Socket Mode events.
type SocketModeCallback func(envelope *SocketModeEnvelope)

// SetEventHandler sets the callback for Socket Mode events.
func (g *Gateway) SetEventHandler(cb SocketModeCallback) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onEvent = cb
}

// ConnectGateway connects to Slack via Socket Mode and keeps the connection alive.
func ConnectGateway(appToken string) (*Gateway, error) {
	wsURL, err := getSocketModeURL(appToken)
	if err != nil {
		return nil, fmt.Errorf("get socket mode URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dial socket mode: %w", err)
	}

	g := &Gateway{
		appToken: appToken,
		conn:     conn,
		done:     make(chan struct{}),
	}

	// Read the hello message
	if err := g.readHello(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("read hello: %w", err)
	}

	go g.pingLoop()
	go g.readLoop()

	return g, nil
}

// Close disconnects from the Socket Mode gateway.
func (g *Gateway) Close() {
	select {
	case <-g.done:
		return
	default:
		close(g.done)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn != nil {
		g.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		g.conn.Close()
	}
}

// Acknowledge sends an acknowledgement for a Socket Mode envelope.
func (g *Gateway) Acknowledge(envelopeID string, payload any) error {
	ack := map[string]any{
		"envelope_id": envelopeID,
	}
	if payload != nil {
		ack["payload"] = payload
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn == nil {
		return fmt.Errorf("connection closed")
	}
	return g.conn.WriteJSON(ack)
}

func getSocketModeURL(appToken string) (string, error) {
	req, err := http.NewRequest("POST", "https://slack.com/api/apps.connections.open", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct {
		OK    bool   `json:"ok"`
		URL   string `json:"url"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if !data.OK {
		return "", fmt.Errorf("apps.connections.open: %s", data.Error)
	}
	return data.URL, nil
}

func (g *Gateway) readHello() error {
	var envelope struct {
		Type string `json:"type"`
	}
	if err := g.conn.ReadJSON(&envelope); err != nil {
		return err
	}
	if envelope.Type != "hello" {
		return fmt.Errorf("expected hello, got %s", envelope.Type)
	}
	return nil
}

func (g *Gateway) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-g.done:
			return
		case <-ticker.C:
			g.mu.Lock()
			err := g.conn.WriteMessage(websocket.PingMessage, nil)
			g.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (g *Gateway) readLoop() {
	for {
		select {
		case <-g.done:
			return
		default:
		}
		var envelope SocketModeEnvelope
		if err := g.conn.ReadJSON(&envelope); err != nil {
			return
		}
		if g.onEvent != nil {
			g.onEvent(&envelope)
		}
	}
}
