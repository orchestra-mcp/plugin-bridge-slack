# Tools Reference

## start_slack_bot

Start the Slack bot (Socket Mode connection).

### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| (none) | | | Starts the bot with current config |

## stop_slack_bot

Stop the Slack bot.

### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| (none) | | | Stops the running bot |

## slack_bot_status

Get the current status of the Slack bot.

### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| (none) | | | Returns status info |

## slack_send_message

Send a message to a Slack channel.

### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | No | Channel ID (defaults to configured) |
| `content` | string | Yes | Message text |
| `title` | string | No | Optional title for the message |
| `color` | string | No | Color: success, error, info, warning |

## slack_set_config

Update Slack bot configuration.

### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `enabled` | boolean | No | Enable/disable bot |
| `bot_token` | string | No | Bot token (xoxb-...) |
| `app_token` | string | No | App-level token (xapp-...) |
| `signing_secret` | string | No | Signing secret |
| `app_id` | string | No | Slack app ID |
| `channel_id` | string | No | Default channel ID |
| `command_prefix` | string | No | Command prefix (default: !) |
| `webhook_url` | string | No | Incoming webhook URL |
| `allowed_users` | string | No | Comma-separated user IDs |
| `team_id` | string | No | Slack team/workspace ID |
