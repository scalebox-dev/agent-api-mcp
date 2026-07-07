package config

import (
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
