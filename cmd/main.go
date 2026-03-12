package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal/handlers"
	"github.com/orchestra-mcp/plugin-bridge-slack/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
	"google.golang.org/protobuf/types/known/structpb"
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

	// Wire lazy caller — uses OrchestratorClient once Run() connects.
	bot.SetCaller(func(name string, args map[string]any) (string, error) {
		client := p.OrchestratorClient()
		if client == nil {
			return "", fmt.Errorf("not connected to orchestrator")
		}
		argsStruct, _ := structpb.NewStruct(args)
		resp, err := client.Send(ctx, &pluginv1.PluginRequest{
			Request: &pluginv1.PluginRequest_ToolCall{
				ToolCall: &pluginv1.ToolRequest{
					ToolName:  name,
					Arguments: argsStruct,
				},
			},
		})
		if err != nil {
			return "", err
		}
		tc := resp.GetToolCall()
		if tc == nil {
			return "", fmt.Errorf("unexpected response for tool %s", name)
		}
		if !tc.Success {
			return "", fmt.Errorf("%s: %s", tc.ErrorCode, tc.ErrorMessage)
		}
		return internal.ExtractText(tc.Result), nil
	})

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
