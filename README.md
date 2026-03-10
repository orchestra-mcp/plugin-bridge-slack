# Orchestra Plugin: bridge-slack

A tools plugin for the [Orchestra MCP](https://github.com/orchestra-mcp/framework) framework.

## Install

```bash
go install github.com/orchestra-mcp/plugin-bridge-slack/cmd@latest
```

## Usage

Add to your `plugins.yaml`:

```yaml
- id: tools.bridge-slack
  binary: ./bin/bridge-slack
  enabled: true
```

## Tools

| Tool | Description |
|------|-------------|
| `start_slack_bot` | Start the Slack bot |
| `stop_slack_bot` | Stop the Slack bot |
| `slack_bot_status` | Get bot status |
| `slack_send_message` | Send a message to Slack |
| `slack_set_config` | Update Slack configuration |

## Related Packages

- [sdk-go](https://github.com/orchestra-mcp/sdk-go) — Plugin SDK
- [gen-go](https://github.com/orchestra-mcp/gen-go) — Generated Protobuf types
