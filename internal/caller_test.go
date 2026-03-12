package internal

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestExtractText_Nil(t *testing.T) {
	got := ExtractText(nil)
	if got != "" {
		t.Errorf("ExtractText(nil) = %q, want empty string", got)
	}
}

func TestExtractText_WithText(t *testing.T) {
	s, err := structpb.NewStruct(map[string]any{
		"text": "hello world",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := ExtractText(s)
	if got != "hello world" {
		t.Errorf("ExtractText = %q, want %q", got, "hello world")
	}
}

func TestExtractText_WithoutText(t *testing.T) {
	s, err := structpb.NewStruct(map[string]any{
		"status": "ok",
		"count":  float64(42),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := ExtractText(s)

	// Should be valid JSON containing both fields.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("ExtractText returned invalid JSON: %q, err: %v", got, err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("parsed[status] = %v, want ok", parsed["status"])
	}
	if parsed["count"] != float64(42) {
		t.Errorf("parsed[count] = %v, want 42", parsed["count"])
	}
}

func TestExtractText_EmptyStruct(t *testing.T) {
	s, err := structpb.NewStruct(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	got := ExtractText(s)
	if got != "{}" {
		t.Errorf("ExtractText(empty struct) = %q, want %q", got, "{}")
	}
}

func TestExtractText_NestedStruct(t *testing.T) {
	s, err := structpb.NewStruct(map[string]any{
		"metadata": map[string]any{
			"version": "1.0",
			"tags":    []any{"alpha", "beta"},
		},
		"active": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := ExtractText(s)

	// No "text" field, so should return JSON of the whole struct.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("ExtractText returned invalid JSON: %q, err: %v", got, err)
	}
	meta, ok := parsed["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata not a map, got %T", parsed["metadata"])
	}
	if meta["version"] != "1.0" {
		t.Errorf("metadata.version = %v, want 1.0", meta["version"])
	}
	tags, ok := meta["tags"].([]any)
	if !ok {
		t.Fatalf("metadata.tags not a slice, got %T", meta["tags"])
	}
	if len(tags) != 2 {
		t.Errorf("metadata.tags len = %d, want 2", len(tags))
	}
	if parsed["active"] != true {
		t.Errorf("active = %v, want true", parsed["active"])
	}
}
