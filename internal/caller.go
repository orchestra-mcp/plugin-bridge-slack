package internal

import (
	"context"
	"encoding/json"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// Sender is satisfied by both *OrchestratorClient (QUIC) and
// *inprocess.Router (direct in-process calls).
type Sender interface {
	Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error)
}

// MakeToolCaller returns a ToolCaller that routes tool calls through the
// given Sender (orchestrator client or in-process router).
func MakeToolCaller(ctx context.Context, sender Sender) ToolCaller {
	return func(name string, args map[string]any) (string, error) {
		argsStruct, err := structpb.NewStruct(args)
		if err != nil {
			return "", fmt.Errorf("build args: %w", err)
		}

		resp, err := sender.Send(ctx, &pluginv1.PluginRequest{
			Request: &pluginv1.PluginRequest_ToolCall{
				ToolCall: &pluginv1.ToolRequest{
					ToolName:  name,
					Arguments: argsStruct,
				},
			},
		})
		if err != nil {
			return "", fmt.Errorf("call tool %s: %w", name, err)
		}

		tc := resp.GetToolCall()
		if tc == nil {
			return "", fmt.Errorf("unexpected response for tool %s", name)
		}
		if !tc.Success {
			return "", fmt.Errorf("%s: %s", tc.ErrorCode, tc.ErrorMessage)
		}
		return ExtractText(tc.Result), nil
	}
}

// ExtractText pulls a string from a ToolResponse Result struct.
// Convention: tools return {"text": "..."} in the Result struct.
func ExtractText(s *structpb.Struct) string {
	if s == nil {
		return ""
	}
	fields := s.GetFields()
	if v, ok := fields["text"]; ok {
		return v.GetStringValue()
	}
	// Fallback: JSON-encode the whole struct.
	b, _ := json.Marshal(s.AsMap())
	return string(b)
}
