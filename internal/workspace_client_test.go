package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkspaceClient_ChatSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Response: "hello"})
	}))
	defer srv.Close()

	client := NewWorkspaceClient(srv.URL, "test-token")
	got, err := client.Chat("ws-123", "hi")
	if err != nil {
		t.Fatalf("Chat() error: %v", err)
	}
	if got != "hello" {
		t.Errorf("Chat() = %q, want %q", got, "hello")
	}
}

func TestWorkspaceClient_ChatError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	client := NewWorkspaceClient(srv.URL, "test-token")
	_, err := client.Chat("ws-123", "hi")
	if err == nil {
		t.Fatal("Chat() expected error for 500 status, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, expected to contain '500'", err.Error())
	}
}

func TestWorkspaceClient_ChatAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Error: "not found"})
	}))
	defer srv.Close()

	client := NewWorkspaceClient(srv.URL, "test-token")
	_, err := client.Chat("ws-123", "hi")
	if err == nil {
		t.Fatal("Chat() expected error for API error response, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, expected to contain 'not found'", err.Error())
	}
}

func TestWorkspaceClient_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Response: "ok"})
	}))
	defer srv.Close()

	client := NewWorkspaceClient(srv.URL, "my-secret-token")
	_, err := client.Chat("ws-123", "hi")
	if err != nil {
		t.Fatalf("Chat() error: %v", err)
	}
	want := "Bearer my-secret-token"
	if gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}

func TestWorkspaceClient_NoToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Response: "ok"})
	}))
	defer srv.Close()

	client := NewWorkspaceClient(srv.URL, "")
	_, err := client.Chat("ws-123", "hi")
	if err != nil {
		t.Fatalf("Chat() error: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization header = %q, want empty (no token)", gotAuth)
	}
}

func TestWorkspaceClient_URLConstruction(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Response: "ok"})
	}))
	defer srv.Close()

	client := NewWorkspaceClient(srv.URL, "token")
	_, err := client.Chat("repo-abc-456", "hi")
	if err != nil {
		t.Fatalf("Chat() error: %v", err)
	}
	want := "/api/repos/repo-abc-456/chat"
	if gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}
}
