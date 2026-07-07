package hosted

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scalebox-dev/agent-api-mcp/internal/config"
)

func TestHealthAndMCPInitialize(t *testing.T) {
	cfg := config.Load([]string{
		"AGENT_API_BASE_URL=http://agent-api.local",
		"AGENT_API_MCP_PATH=/mcp",
	})
	server, err := NewServer(cfg, slog.Default())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /healthz status = %d", resp.StatusCode)
	}
	resp.Body.Close()

	body := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/mcp", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer sk_test")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /mcp initialize: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /mcp status = %d", resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode initialize response: %v", err)
	}
	result, ok := payload["result"].(map[string]any)
	if !ok {
		t.Fatalf("initialize response missing result: %#v", payload)
	}
	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok || serverInfo["name"] != "agent-api-mcp" {
		t.Fatalf("unexpected serverInfo: %#v", result["serverInfo"])
	}
}

func TestUnauthorizedMCPAdvertisesProtectedResourceMetadata(t *testing.T) {
	cfg := config.Load([]string{
		"AGENT_API_BASE_URL=http://agent-api.local",
		"AGENT_API_AUTHORIZATION_SERVER_URL=https://api.example.test",
		"AGENT_API_MCP_PUBLIC_BASE_URL=https://mcp.example.test",
		"AGENT_API_MCP_PATH=/mcp",
	})
	server, err := NewServer(cfg, slog.Default())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	body := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/mcp", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("POST /mcp status = %d", resp.StatusCode)
	}
	if got, want := resp.Header.Get("WWW-Authenticate"), `Bearer resource_metadata="https://mcp.example.test/.well-known/oauth-protected-resource"`; got != want {
		t.Fatalf("WWW-Authenticate = %q, want %q", got, want)
	}
}

func TestProtectedResourceMetadata(t *testing.T) {
	cfg := config.Load([]string{
		"AGENT_API_BASE_URL=http://agent-api.local",
		"AGENT_API_AUTHORIZATION_SERVER_URL=https://api.example.test",
		"AGENT_API_MCP_PUBLIC_BASE_URL=https://mcp.example.test",
		"AGENT_API_MCP_PATH=/mcp",
	})
	server, err := NewServer(cfg, slog.Default())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/.well-known/oauth-protected-resource")
	if err != nil {
		t.Fatalf("GET metadata: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET metadata status = %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if got, want := payload["resource"], "https://mcp.example.test/mcp"; got != want {
		t.Fatalf("resource = %#v, want %q", got, want)
	}
	authServers, ok := payload["authorization_servers"].([]any)
	if !ok || len(authServers) != 1 || authServers[0] != "https://api.example.test" {
		t.Fatalf("authorization_servers = %#v", payload["authorization_servers"])
	}
	if got := payload["bearer_methods_supported"]; got == nil {
		t.Fatalf("missing bearer_methods_supported: %#v", payload)
	}
}

func TestProtectedResourceMetadataCanUseForwardedHost(t *testing.T) {
	cfg := config.Load([]string{
		"AGENT_API_BASE_URL=http://agent-api.local",
		"AGENT_API_MCP_PATH=/mcp",
	})
	server, err := NewServer(cfg, slog.Default())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/.well-known/oauth-protected-resource/mcp", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "mcp.local.test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET metadata: %v", err)
	}
	defer resp.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if got, want := payload["resource"], "https://mcp.local.test/mcp"; got != want {
		t.Fatalf("resource = %#v, want %q", got, want)
	}
}
