package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal/handlers"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("bridge.slack").
		Version("0.1.0").
		Description("Slack bridge plugin for Orchestra MCP").
		Author("Orchestra").
		Binary("bridge-slack")

	cfg := internal.LoadConfig()
	bot := internal.NewBot(cfg, nil)

	// Wire up handler registration (avoids import cycle in internal)
	bot.SetHandlerRegistrar(func(r *internal.Router) {
		chat := handlers.NewChatHandler()
		r.Register(chat)
		r.SetDefault(chat)
		r.Register(handlers.NewMcpHandler())
		r.Register(handlers.NewStatusHandler())
		r.Register(handlers.NewToolsHandler())
		r.Register(handlers.NewStopHandler())
		r.Register(handlers.NewProgressHandler())
		r.Register(handlers.NewPingHandler())
		r.Register(handlers.NewPermissionHandler())
	})

	// Register MCP tools
	bp := &internal.BridgePlugin{Bot: bot}
	tools.RegisterAll(builder, bp)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Start bot in background if configured
	if cfg.Enabled && cfg.IsValid() {
		go func() {
			if err := bot.Start(ctx); err != nil {
				log.Printf("[slack] bot error: %v", err)
			}
		}()
	}

	if err := p.Run(ctx); err != nil {
		log.Fatalf("bridge.slack: %v", err)
	}
}
