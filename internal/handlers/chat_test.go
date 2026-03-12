package handlers

import "testing"

func TestParseWorkspace_NoPrefix(t *testing.T) {
	wsID, prompt := parseWorkspace("hello world")
	if wsID != "" {
		t.Errorf("expected empty wsID, got %q", wsID)
	}
	if prompt != "hello world" {
		t.Errorf("expected prompt %q, got %q", "hello world", prompt)
	}
}

func TestParseWorkspace_WithWorkspace(t *testing.T) {
	wsID, prompt := parseWorkspace("@myrepo what is the status?")
	if wsID != "myrepo" {
		t.Errorf("expected wsID %q, got %q", "myrepo", wsID)
	}
	if prompt != "what is the status?" {
		t.Errorf("expected prompt %q, got %q", "what is the status?", prompt)
	}
}

func TestParseWorkspace_WorkspaceOnly(t *testing.T) {
	wsID, prompt := parseWorkspace("@myrepo")
	if wsID != "myrepo" {
		t.Errorf("expected wsID %q, got %q", "myrepo", wsID)
	}
	if prompt != "" {
		t.Errorf("expected empty prompt, got %q", prompt)
	}
}

func TestParseWorkspace_WorkspaceWithDashes(t *testing.T) {
	wsID, prompt := parseWorkspace("@my-cool-repo list features")
	if wsID != "my-cool-repo" {
		t.Errorf("expected wsID %q, got %q", "my-cool-repo", wsID)
	}
	if prompt != "list features" {
		t.Errorf("expected prompt %q, got %q", "list features", prompt)
	}
}

func TestParseWorkspace_EmptyString(t *testing.T) {
	wsID, prompt := parseWorkspace("")
	if wsID != "" {
		t.Errorf("expected empty wsID, got %q", wsID)
	}
	if prompt != "" {
		t.Errorf("expected empty prompt, got %q", prompt)
	}
}

func TestParseWorkspace_AtOnly(t *testing.T) {
	wsID, prompt := parseWorkspace("@")
	if wsID != "" {
		t.Errorf("expected empty wsID, got %q", wsID)
	}
	if prompt != "" {
		t.Errorf("expected empty prompt, got %q", prompt)
	}
}

func TestParseWorkspace_WorkspaceUUID(t *testing.T) {
	wsID, prompt := parseWorkspace("@abc123-def456 show status")
	if wsID != "abc123-def456" {
		t.Errorf("expected wsID %q, got %q", "abc123-def456", wsID)
	}
	if prompt != "show status" {
		t.Errorf("expected prompt %q, got %q", "show status", prompt)
	}
}
