package bridgeslack

import (
	"context"

	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal/handlers"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Cleanup is a function that should be called during shutdown.
type Cleanup func()

// Register adds all Slack bridge tools to the builder and starts the bot
// in the background if configured. Returns a cleanup function.
func Register(builder *plugin.PluginBuilder) Cleanup {
	return RegisterWithContext(context.Background(), builder)
}

// RegisterWithContext is like Register but accepts a context for the bot lifecycle.
func RegisterWithContext(ctx context.Context, builder *plugin.PluginBuilder) Cleanup {
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

	if cfg.Enabled && cfg.IsValid() {
		go bot.Start(ctx)
	}

	return bot.Stop
}
