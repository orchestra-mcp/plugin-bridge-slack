package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// ToolCaller is a function that invokes MCP tools by name.
type ToolCaller func(name string, args map[string]any) (string, error)

// HandlerRegistrar is a callback that registers handlers on a router.
// This allows external packages to wire up handlers without import cycles.
type HandlerRegistrar func(r *Router)

// Bot manages Slack integration -- Socket Mode gateway, router, handlers.
type Bot struct {
	config    *Config
	gateway   *Gateway
	router    *Router
	rest      *RestClient
	service   *NotificationService
	caller    ToolCaller
	registrar HandlerRegistrar
	mu        sync.RWMutex
	running   bool
}

// NewBot creates a Slack bot from configuration.
func NewBot(cfg *Config, caller ToolCaller) *Bot {
	return &Bot{
		config:  cfg,
		service: NewNotificationService(cfg),
		caller:  caller,
	}
}

// SetHandlerRegistrar sets the callback used to register handlers on Start.
func (b *Bot) SetHandlerRegistrar(fn HandlerRegistrar) {
	b.registrar = fn
}

// Start connects to Slack via Socket Mode, registers handlers, and starts routing events.
func (b *Bot) Start(ctx context.Context) error {
	if b.config == nil || !b.config.Enabled || b.config.AppToken == "" {
		log.Println("[slack] bot disabled or not configured")
		return nil
	}

	gw, err := ConnectGateway(b.config.AppToken)
	if err != nil {
		return fmt.Errorf("connect gateway: %w", err)
	}
	b.gateway = gw
	b.rest = NewRestClient(b.config.BotToken)

	prefix := b.config.CommandPrefix
	if prefix == "" {
		prefix = "!"
	}
	b.router = NewRouter(prefix)

	// Register handlers via external registrar
	if b.registrar != nil {
		b.registrar(b.router)
	}

	gw.SetEventHandler(b.onSocketModeEvent)
	b.running = true

	log.Printf("[slack] bot started (%d handlers)", len(b.router.handlers))

	// Wait for context cancellation
	<-ctx.Done()
	b.Stop()
	return nil
}

// Stop gracefully stops the Slack bot.
func (b *Bot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.gateway != nil {
		b.gateway.Close()
		b.gateway = nil
	}
	b.running = false
	log.Println("[slack] bot stopped")
}

// IsRunning returns whether the bot is active.
func (b *Bot) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

func (b *Bot) onSocketModeEvent(envelope *SocketModeEnvelope) {
	// Always acknowledge the envelope first
	if envelope.EnvelopeID != "" {
		if err := b.gateway.Acknowledge(envelope.EnvelopeID, nil); err != nil {
			log.Printf("[slack] failed to ack envelope: %v", err)
		}
	}

	switch envelope.Type {
	case "events_api":
		var callback EventCallback
		if err := json.Unmarshal(envelope.Payload, &callback); err != nil {
			return
		}
		b.handleEventCallback(callback)
	case "slash_commands":
		var cmd SlashCommandPayload
		if err := json.Unmarshal(envelope.Payload, &cmd); err != nil {
			return
		}
		b.router.RouteSlashCommand(cmd, b)
	case "interactive":
		var payload InteractionPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return
		}
		b.router.RouteInteraction(payload, b)
	}
}

func (b *Bot) handleEventCallback(callback EventCallback) {
	var event struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(callback.Event, &event); err != nil {
		return
	}

	switch event.Type {
	case "app_mention":
		var msg MessageEvent
		if err := json.Unmarshal(callback.Event, &msg); err != nil {
			return
		}
		if msg.BotID != "" || msg.SubType != "" {
			return
		}
		// @mentions go directly to AI chat without prefix
		b.router.RouteDirect(msg, b)
	case "message":
		var msg MessageEvent
		if err := json.Unmarshal(callback.Event, &msg); err != nil {
			return
		}
		if msg.BotID != "" || msg.SubType != "" {
			return
		}
		// DMs (channel starts with D) go directly to AI chat
		if len(msg.Channel) > 0 && msg.Channel[0] == 'D' {
			b.router.RouteDirect(msg, b)
			return
		}
		b.router.RouteMessage(msg, b)
	}
}

// --- HandlerAPI implementation ---

// SendToChannel implements HandlerAPI.
func (b *Bot) SendToChannel(channelID, text string, blocks []Block, attachments []Attachment) error {
	if b.rest == nil {
		return fmt.Errorf("REST client not initialized")
	}
	_, err := b.rest.SendMessage(channelID, text, blocks, attachments)
	return err
}

// SendBlocks implements HandlerAPI.
func (b *Bot) SendBlocks(channelID string, blocks []Block, attachments []Attachment) (string, error) {
	if b.rest == nil {
		return "", fmt.Errorf("REST client not initialized")
	}
	return b.rest.SendMessage(channelID, "", blocks, attachments)
}

// UpdateMessage implements HandlerAPI.
func (b *Bot) UpdateMessage(channelID, ts, text string, blocks []Block) error {
	if b.rest == nil {
		return fmt.Errorf("REST client not initialized")
	}
	return b.rest.UpdateMessage(channelID, ts, text, blocks)
}

// RespondToURL implements HandlerAPI.
func (b *Bot) RespondToURL(responseURL, text string, blocks []Block, replaceOriginal bool) error {
	if b.rest == nil {
		return fmt.Errorf("REST client not initialized")
	}
	return b.rest.RespondToURL(responseURL, text, blocks, replaceOriginal)
}

// ChannelID implements HandlerAPI.
func (b *Bot) ChannelID() string { return b.config.ChannelID }

// Config implements HandlerAPI.
func (b *Bot) Config() *Config { return b.config }

// CallTool implements HandlerAPI -- invokes MCP tools via cross-plugin calls.
func (b *Bot) CallTool(name string, args map[string]any) (string, error) {
	if b.caller == nil {
		return "", fmt.Errorf("tool caller not configured")
	}
	return b.caller(name, args)
}
