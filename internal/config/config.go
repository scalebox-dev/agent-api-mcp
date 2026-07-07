package config

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAgentAPIBaseURL = "https://api.agentsway.dev"
	defaultListenAddr      = ":8080"
	defaultMCPPath         = "/mcp"
	defaultHTTPTimeout     = 10 * time.Minute
	defaultSessionTimeout  = 30 * time.Minute
)

type Config struct {
	AgentAPIBaseURL string
	AgentAPIKey     string
	ListenAddr      string
	MCPPath         string
	HTTPTimeout     time.Duration
	SessionTimeout  time.Duration
}

func Load(environ []string) Config {
	env := map[string]string{}
	for _, item := range environ {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			env[key] = value
		}
	}

	return Config{
		AgentAPIBaseURL: cleanBaseURL(firstNonEmpty(env["AGENT_API_BASE_URL"], defaultAgentAPIBaseURL)),
		AgentAPIKey:     strings.TrimSpace(env["AGENT_API_KEY"]),
		ListenAddr:      firstNonEmpty(env["AGENT_API_MCP_ADDR"], defaultListenAddr),
		MCPPath:         cleanPath(firstNonEmpty(env["AGENT_API_MCP_PATH"], defaultMCPPath)),
		HTTPTimeout:     durationMillis(env["AGENT_API_HTTP_TIMEOUT_MS"], defaultHTTPTimeout),
		SessionTimeout:  durationMillis(env["AGENT_API_MCP_SESSION_TIMEOUT_MS"], defaultSessionTimeout),
	}
}

func (c Config) Ready() bool {
	return c.AgentAPIBaseURL != "" && c.ListenAddr != "" && c.MCPPath != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func cleanBaseURL(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return defaultAgentAPIBaseURL
	}
	if _, err := url.ParseRequestURI(trimmed); err != nil {
		return defaultAgentAPIBaseURL
	}
	return trimmed
}

func cleanPath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultMCPPath
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return trimmed
}

func durationMillis(raw string, fallback time.Duration) time.Duration {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return time.Duration(value) * time.Millisecond
}
