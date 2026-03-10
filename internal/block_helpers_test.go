package internal

import (
	"strings"
	"testing"
)

func TestSuccessBlocks(t *testing.T) {
	blocks, attachments := SuccessBlocks("Test Title", "Test description")
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Type != "header" {
		t.Errorf("block[0].type = %q, want header", blocks[0].Type)
	}
	if blocks[1].Type != "section" {
		t.Errorf("block[1].type = %q, want section", blocks[1].Type)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].Color != ColorSuccess {
		t.Errorf("color = %q, want %q", attachments[0].Color, ColorSuccess)
	}
}

func TestErrorBlocks(t *testing.T) {
	blocks, attachments := ErrorBlocks("Error Title", "Error desc")
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if attachments[0].Color != ColorError {
		t.Errorf("color = %q, want %q", attachments[0].Color, ColorError)
	}
}

func TestInfoBlocks(t *testing.T) {
	_, attachments := InfoBlocks("Info Title", "Info desc")
	if attachments[0].Color != ColorInfo {
		t.Errorf("color = %q, want %q", attachments[0].Color, ColorInfo)
	}
}

func TestWarningBlocks(t *testing.T) {
	_, attachments := WarningBlocks("Warning Title", "Warning desc")
	if attachments[0].Color != ColorWarning {
		t.Errorf("color = %q, want %q", attachments[0].Color, ColorWarning)
	}
}

func TestToolBlocks_Running(t *testing.T) {
	blocks, attachments := ToolBlocks("Read", "running", "reading file")
	if len(blocks) < 1 {
		t.Fatal("expected at least 1 block")
	}
	if attachments[0].Color != ColorInfo {
		t.Errorf("running tool color = %q, want %q", attachments[0].Color, ColorInfo)
	}
}

func TestToolBlocks_Done(t *testing.T) {
	_, attachments := ToolBlocks("Write", "done", "wrote file")
	if attachments[0].Color != ColorSuccess {
		t.Errorf("done tool color = %q, want %q", attachments[0].Color, ColorSuccess)
	}
}

func TestToolBlocks_Error(t *testing.T) {
	_, attachments := ToolBlocks("Bash", "error", "command failed")
	if attachments[0].Color != ColorError {
		t.Errorf("error tool color = %q, want %q", attachments[0].Color, ColorError)
	}
}

func TestActionBlocks_ValidJSON(t *testing.T) {
	raw := `{"tool":"Read","input":"{\"file_path\":\"/tmp/test.go\"}","status":"done","result":"file contents"}`
	blocks, _ := ActionBlocks(raw)
	if len(blocks) < 1 {
		t.Fatal("expected at least 1 block")
	}
}

func TestActionBlocks_InvalidJSON(t *testing.T) {
	blocks, _ := ActionBlocks("not json at all")
	if len(blocks) < 1 {
		t.Fatal("expected at least 1 block for invalid JSON")
	}
}

func TestPermissionBlocks(t *testing.T) {
	blocks, attachments := PermissionBlocks("Bash", "needs approval", "rm -rf /tmp", "REQ-123")
	if len(blocks) < 4 {
		t.Fatalf("expected at least 4 blocks (header + section + section + actions), got %d", len(blocks))
	}
	if blocks[0].Type != "header" {
		t.Errorf("block[0].type = %q, want header", blocks[0].Type)
	}
	// Check that approve/deny buttons exist
	actionsBlock := blocks[3]
	if actionsBlock.Type != "actions" {
		t.Errorf("block[3].type = %q, want actions", actionsBlock.Type)
	}
	if len(actionsBlock.Elements) != 2 {
		t.Fatalf("expected 2 button elements, got %d", len(actionsBlock.Elements))
	}
	if actionsBlock.Elements[0].ActionID != "perm_approve_REQ-123" {
		t.Errorf("approve action_id = %q, want perm_approve_REQ-123", actionsBlock.Elements[0].ActionID)
	}
	if actionsBlock.Elements[1].ActionID != "perm_deny_REQ-123" {
		t.Errorf("deny action_id = %q, want perm_deny_REQ-123", actionsBlock.Elements[1].ActionID)
	}
	if attachments[0].Color != ColorWarning {
		t.Errorf("permission color = %q, want %q", attachments[0].Color, ColorWarning)
	}
}

func TestPermissionResultBlocks_Approved(t *testing.T) {
	blocks, attachments := PermissionResultBlocks("Approved", "REQ-123")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if attachments[0].Color != ColorSuccess {
		t.Errorf("approved color = %q, want %q", attachments[0].Color, ColorSuccess)
	}
}

func TestPermissionResultBlocks_Denied(t *testing.T) {
	_, attachments := PermissionResultBlocks("Denied", "REQ-456")
	if attachments[0].Color != ColorError {
		t.Errorf("denied color = %q, want %q", attachments[0].Color, ColorError)
	}
}

func TestTruncate_Short(t *testing.T) {
	s := "hello"
	if got := Truncate(s, 100); got != "hello" {
		t.Errorf("Truncate(%q, 100) = %q", s, got)
	}
}

func TestTruncate_Exact(t *testing.T) {
	s := "hello"
	if got := Truncate(s, 5); got != "hello" {
		t.Errorf("Truncate(%q, 5) = %q", s, got)
	}
}

func TestTruncate_Long(t *testing.T) {
	s := "hello world this is a long string"
	got := Truncate(s, 10)
	if len(got) > 10 {
		t.Errorf("Truncate result len = %d, want <= 10", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncated string should end with ..., got %q", got)
	}
}

func TestTruncate_VerySmallMax(t *testing.T) {
	s := "hello"
	got := Truncate(s, 2)
	if len(got) > 2 {
		t.Errorf("Truncate with max=2 len = %d, want <= 2", len(got))
	}
}

func TestToolEmoji(t *testing.T) {
	tests := map[string]string{
		"Read":       ":book:",
		"Write":      ":memo:",
		"Edit":       ":pencil2:",
		"Bash":       ":computer:",
		"Grep":       ":mag:",
		"Glob":       ":file_folder:",
		"Task":       ":robot_face:",
		"TodoWrite":  ":clipboard:",
		"WebFetch":   ":globe_with_meridians:",
		"WebSearch":  ":mag_right:",
		"unknown":    ":wrench:",
		"mcp__foo__bar": ":electric_plug:",
	}
	for tool, want := range tests {
		got := toolEmoji(tool)
		if got != want {
			t.Errorf("toolEmoji(%q) = %q, want %q", tool, got, want)
		}
	}
}

func TestHumanToolName(t *testing.T) {
	tests := map[string]string{
		"Read":  "Reading file",
		"Write": "Writing file",
		"Bash":  "Running command",
		"Grep":  "Searching code",
		"mcp__orchestra__list_features": "orchestra/list features",
	}
	for tool, want := range tests {
		got := humanToolName(tool)
		if got != want {
			t.Errorf("humanToolName(%q) = %q, want %q", tool, got, want)
		}
	}
}

func TestShortPath(t *testing.T) {
	tests := map[string]string{
		"/Users/fady/Sites/project/main.go":        "project/main.go",
		"/tmp/test.go":                              "tmp/test.go",
		"file.go":                                   "file.go",
	}
	for input, want := range tests {
		got := shortPath(input)
		if got != want {
			t.Errorf("shortPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFormatToolInput_Read(t *testing.T) {
	got := formatToolInput("Read", `{"file_path":"/tmp/test.go"}`)
	if !strings.Contains(got, "test.go") {
		t.Errorf("formatToolInput(Read) = %q, expected to contain test.go", got)
	}
}

func TestFormatToolInput_Bash(t *testing.T) {
	got := formatToolInput("Bash", `{"command":"go test ./..."}`)
	if !strings.Contains(got, "go test") {
		t.Errorf("formatToolInput(Bash) = %q, expected to contain 'go test'", got)
	}
}

func TestFormatToolInput_InvalidJSON(t *testing.T) {
	got := formatToolInput("Read", "not json")
	if got != "not json" {
		t.Errorf("formatToolInput with invalid JSON = %q, expected 'not json'", got)
	}
}
