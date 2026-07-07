package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load(nil)
	if cfg.AgentAPIBaseURL != defaultAgentAPIBaseURL {
		t.Fatalf("AgentAPIBaseURL = %q, want %q", cfg.AgentAPIBaseURL, defaultAgentAPIBaseURL)
	}
	if cfg.ListenAddr != defaultListenAddr {
		t.Fatalf("ListenAddr = %q, want %q", cfg.ListenAddr, defaultListenAddr)
	}
	if cfg.MCPPath != defaultMCPPath {
		t.Fatalf("MCPPath = %q, want %q", cfg.MCPPath, defaultMCPPath)
	}
	if cfg.HTTPTimeout != defaultHTTPTimeout {
		t.Fatalf("HTTPTimeout = %v, want %v", cfg.HTTPTimeout, defaultHTTPTimeout)
	}
}

func TestLoadExplicitValues(t *testing.T) {
	cfg := Load([]string{
		"AGENT_API_BASE_URL=http://localhost:18000/",
		"AGENT_API_KEY=sk_test",
		"AGENT_API_MCP_ADDR=:9090",
		"AGENT_API_MCP_PATH=mcp",
		"AGENT_API_HTTP_TIMEOUT_MS=1234",
	})
	if cfg.AgentAPIBaseURL != "http://localhost:18000" {
		t.Fatalf("AgentAPIBaseURL = %q", cfg.AgentAPIBaseURL)
	}
	if cfg.AgentAPIKey != "sk_test" || cfg.ListenAddr != ":9090" || cfg.MCPPath != "/mcp" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.HTTPTimeout != 1234*time.Millisecond {
		t.Fatalf("HTTPTimeout = %v", cfg.HTTPTimeout)
	}
}

func TestLoadDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(`
# local development settings
AGENT_API_BASE_URL=http://localhost:18000/
AGENT_API_KEY="sk_from_file"
AGENT_API_MCP_ADDR=:9090
AGENT_API_MCP_PATH=mcp
AGENT_API_HTTP_TIMEOUT_MS=1234
AGENT_API_MCP_SESSION_TIMEOUT_MS=5678
IGNORED_LINE
`), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := LoadDotEnv(nil, path)
	if cfg.AgentAPIBaseURL != "http://localhost:18000" {
		t.Fatalf("AgentAPIBaseURL = %q", cfg.AgentAPIBaseURL)
	}
	if cfg.AgentAPIKey != "sk_from_file" || cfg.ListenAddr != ":9090" || cfg.MCPPath != "/mcp" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.HTTPTimeout != 1234*time.Millisecond {
		t.Fatalf("HTTPTimeout = %v", cfg.HTTPTimeout)
	}
	if cfg.SessionTimeout != 5678*time.Millisecond {
		t.Fatalf("SessionTimeout = %v", cfg.SessionTimeout)
	}
}

func TestLoadDotEnvProcessEnvWins(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("AGENT_API_KEY=sk_from_file\nAGENT_API_MCP_ADDR=:9090\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := LoadDotEnv([]string{"AGENT_API_KEY=sk_from_env"}, path)
	if cfg.AgentAPIKey != "sk_from_env" {
		t.Fatalf("AgentAPIKey = %q, want process env override", cfg.AgentAPIKey)
	}
	if cfg.ListenAddr != ":9090" {
		t.Fatalf("ListenAddr = %q, want .env fallback", cfg.ListenAddr)
	}
}
