package internal

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// Slack message limits
const (
	MaxBlockText   = 3000
	MaxSectionText = 3000
	MaxMessageText = 4000
	// Safe limits (leave room for formatting)
	SafeBlockText  = 2900
	SafeMessageLen = 3900
)

// Color constants for attachments (hex strings)
const (
	ColorSuccess = "#2ecc71"
	ColorError   = "#e74c3c"
	ColorInfo    = "#3498db"
	ColorWarning = "#f39c12"
)

// SuccessBlocks creates green-colored blocks for successful operations.
func SuccessBlocks(title, desc string) ([]Block, []Attachment) {
	blocks := []Block{
		HeaderBlock(Truncate(title, 150)),
		SectionBlock(Truncate(desc, SafeBlockText)),
	}
	attachments := []Attachment{
		{Color: ColorSuccess, Fallback: title},
	}
	return blocks, attachments
}

// ErrorBlocks creates red-colored blocks for errors.
func ErrorBlocks(title, desc string) ([]Block, []Attachment) {
	blocks := []Block{
		HeaderBlock(Truncate(title, 150)),
		SectionBlock(Truncate(desc, SafeBlockText)),
	}
	attachments := []Attachment{
		{Color: ColorError, Fallback: title},
	}
	return blocks, attachments
}

// InfoBlocks creates blue-colored blocks for informational messages.
func InfoBlocks(title, desc string) ([]Block, []Attachment) {
	blocks := []Block{
		HeaderBlock(Truncate(title, 150)),
		SectionBlock(Truncate(desc, SafeBlockText)),
	}
	attachments := []Attachment{
		{Color: ColorInfo, Fallback: title},
	}
	return blocks, attachments
}

// WarningBlocks creates orange-colored blocks for warnings.
func WarningBlocks(title, desc string) ([]Block, []Attachment) {
	blocks := []Block{
		HeaderBlock(Truncate(title, 150)),
		SectionBlock(Truncate(desc, SafeBlockText)),
	}
	attachments := []Attachment{
		{Color: ColorWarning, Fallback: title},
	}
	return blocks, attachments
}

// ToolBlocks creates colored blocks for tool execution status.
func ToolBlocks(tool, status, detail string) ([]Block, []Attachment) {
	color := ColorInfo
	emoji := ":wrench:"
	switch status {
	case "done":
		color = ColorSuccess
		emoji = ":white_check_mark:"
	case "error":
		color = ColorError
		emoji = ":x:"
	}
	blocks := []Block{
		SectionBlock(fmt.Sprintf("*%s %s*", emoji, Truncate(tool, 200))),
	}
	if detail != "" {
		blocks = append(blocks, SectionBlock(Truncate(detail, SafeBlockText)))
	}
	attachments := []Attachment{
		{Color: color, Fallback: tool},
	}
	return blocks, attachments
}

// ActionBlocks parses a raw action JSON payload and returns human-readable blocks.
func ActionBlocks(rawJSON string) ([]Block, []Attachment) {
	var action struct {
		Tool      string `json:"tool"`
		Input     string `json:"input"`
		Status    string `json:"status"`
		ToolUseID string `json:"toolUseId"`
		Result    string `json:"result"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &action); err != nil {
		return ToolBlocks("Tool", "running", Truncate(rawJSON, SafeBlockText))
	}

	emoji := toolEmoji(action.Tool)
	color := ColorInfo
	if action.Status == "done" {
		color = ColorSuccess
	}

	title := fmt.Sprintf("%s %s", emoji, humanToolName(action.Tool))
	desc := formatToolInput(action.Tool, action.Input)

	blocks := []Block{
		SectionBlock(fmt.Sprintf("*%s*", Truncate(title, 200))),
	}
	if desc != "" {
		blocks = append(blocks, SectionBlock(Truncate(desc, SafeBlockText)))
	}
	if action.Result != "" && action.Status == "done" {
		blocks = append(blocks, SectionBlock(fmt.Sprintf("*Result:*\n```\n%s\n```", Truncate(action.Result, SafeBlockText-20))))
	}

	attachments := []Attachment{
		{Color: color, Fallback: title},
	}
	return blocks, attachments
}

// PermissionBlocks creates blocks for permission requests with approve/deny buttons.
func PermissionBlocks(toolName, reason, input, requestID string) ([]Block, []Attachment) {
	blocks := []Block{
		HeaderBlock(fmt.Sprintf(":lock: Permission: %s", Truncate(toolName, 120))),
		SectionBlock(Truncate(reason, SafeBlockText)),
		SectionBlock(fmt.Sprintf("*Tool:* `%s`\n*Input:*\n```\n%s\n```", toolName, Truncate(input, SafeBlockText-50))),
		ActionsBlock("perm_actions_"+requestID,
			ButtonElement("Approve", "perm_approve_"+requestID, requestID, "primary"),
			ButtonElement("Deny", "perm_deny_"+requestID, requestID, "danger"),
		),
	}
	attachments := []Attachment{
		{Color: ColorWarning, Fallback: "Permission request: " + toolName},
	}
	return blocks, attachments
}

// PermissionResultBlocks creates blocks showing the result of a permission decision.
func PermissionResultBlocks(status, requestID string) ([]Block, []Attachment) {
	color := ColorSuccess
	emoji := ":white_check_mark:"
	if status == "Denied" {
		color = ColorError
		emoji = ":x:"
	}
	blocks := []Block{
		SectionBlock(fmt.Sprintf("*%s Permission %s*", emoji, status)),
	}
	attachments := []Attachment{
		{Color: color, Fallback: "Permission " + status},
	}
	return blocks, attachments
}

// Truncate limits a string to maxLen characters with ellipsis.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// --- Internal helpers ---

func toolEmoji(tool string) string {
	switch tool {
	case "Read":
		return ":book:"
	case "Write":
		return ":memo:"
	case "Edit":
		return ":pencil2:"
	case "Bash":
		return ":computer:"
	case "Grep":
		return ":mag:"
	case "Glob":
		return ":file_folder:"
	case "Task":
		return ":robot_face:"
	case "TodoWrite":
		return ":clipboard:"
	case "WebFetch":
		return ":globe_with_meridians:"
	case "WebSearch":
		return ":mag_right:"
	default:
		if strings.HasPrefix(tool, "mcp__") {
			return ":electric_plug:"
		}
		return ":wrench:"
	}
}

func humanToolName(tool string) string {
	// MCP tools: mcp__server__tool_name -> server/tool_name
	if strings.HasPrefix(tool, "mcp__") {
		parts := strings.SplitN(tool, "__", 3)
		if len(parts) == 3 {
			return parts[1] + "/" + strings.ReplaceAll(parts[2], "_", " ")
		}
	}
	switch tool {
	case "Read":
		return "Reading file"
	case "Write":
		return "Writing file"
	case "Edit":
		return "Editing file"
	case "Bash":
		return "Running command"
	case "Grep":
		return "Searching code"
	case "Glob":
		return "Finding files"
	case "Task":
		return "Sub-agent task"
	case "TodoWrite":
		return "Updating todo list"
	case "WebFetch":
		return "Fetching URL"
	case "WebSearch":
		return "Web search"
	default:
		return tool
	}
}

func formatToolInput(tool, rawInput string) string {
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(rawInput), &input); err != nil {
		return Truncate(rawInput, 200)
	}

	switch tool {
	case "Read":
		fp, _ := input["file_path"].(string)
		if fp != "" {
			return fmt.Sprintf("`%s`", shortPath(fp))
		}
	case "Write":
		fp, _ := input["file_path"].(string)
		if fp != "" {
			return fmt.Sprintf("`%s`", shortPath(fp))
		}
	case "Edit":
		fp, _ := input["file_path"].(string)
		old, _ := input["old_string"].(string)
		if fp != "" {
			desc := fmt.Sprintf("`%s`", shortPath(fp))
			if old != "" {
				desc += fmt.Sprintf("\n```\n%s\n```", Truncate(old, 100))
			}
			return desc
		}
	case "Bash":
		cmd, _ := input["command"].(string)
		if cmd != "" {
			return fmt.Sprintf("```\n%s\n```", Truncate(cmd, 300))
		}
	case "Grep":
		pat, _ := input["pattern"].(string)
		path, _ := input["path"].(string)
		desc := fmt.Sprintf("Pattern: `%s`", pat)
		if path != "" {
			desc += fmt.Sprintf(" in `%s`", shortPath(path))
		}
		return desc
	case "Glob":
		pat, _ := input["pattern"].(string)
		return fmt.Sprintf("Pattern: `%s`", pat)
	case "Task":
		desc, _ := input["description"].(string)
		agentType, _ := input["subagent_type"].(string)
		if desc != "" {
			s := desc
			if agentType != "" {
				s = fmt.Sprintf("[%s] %s", agentType, desc)
			}
			return Truncate(s, 300)
		}
	case "TodoWrite":
		return "Updating task list"
	case "WebSearch":
		q, _ := input["query"].(string)
		if q != "" {
			return fmt.Sprintf("Query: `%s`", Truncate(q, 200))
		}
	case "WebFetch":
		u, _ := input["url"].(string)
		if u != "" {
			return Truncate(u, 300)
		}
	default:
		// MCP tools -- show key args compactly
		if strings.HasPrefix(tool, "mcp__") {
			return formatMCPInput(input)
		}
	}

	return ""
}

func formatMCPInput(input map[string]interface{}) string {
	var parts []string
	for k, v := range input {
		s := fmt.Sprintf("%v", v)
		if len(s) > 80 {
			s = s[:77] + "..."
		}
		parts = append(parts, fmt.Sprintf("*%s:* %s", k, s))
	}
	return Truncate(strings.Join(parts, "\n"), 500)
}

func shortPath(fp string) string {
	// Show just filename or last 2 path components
	base := filepath.Base(fp)
	dir := filepath.Base(filepath.Dir(fp))
	if dir == "." || dir == "/" {
		return base
	}
	return dir + "/" + base
}
