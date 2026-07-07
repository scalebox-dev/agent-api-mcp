package config

import (
	"bufio"
	"net/url"
	"os"
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
	return load(environMap(environ))
}

func LoadDotEnv(environ []string, paths ...string) Config {
	env := map[string]string{}
	for _, path := range paths {
		values, err := readDotEnv(path)
		if err != nil {
			continue
		}
		for key, value := range values {
			env[key] = value
		}
	}
	for key, value := range environMap(environ) {
		env[key] = value
	}
	return load(env)
}

func load(env map[string]string) Config {
	return Config{
		AgentAPIBaseURL: cleanBaseURL(firstNonEmpty(env["AGENT_API_BASE_URL"], defaultAgentAPIBaseURL)),
		AgentAPIKey:     strings.TrimSpace(env["AGENT_API_KEY"]),
		ListenAddr:      firstNonEmpty(env["AGENT_API_MCP_ADDR"], defaultListenAddr),
		MCPPath:         cleanPath(firstNonEmpty(env["AGENT_API_MCP_PATH"], defaultMCPPath)),
		HTTPTimeout:     durationMillis(env["AGENT_API_HTTP_TIMEOUT_MS"], defaultHTTPTimeout),
		SessionTimeout:  durationMillis(env["AGENT_API_MCP_SESSION_TIMEOUT_MS"], defaultSessionTimeout),
	}
}

func environMap(environ []string) map[string]string {
	env := map[string]string{}
	for _, item := range environ {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			env[key] = value
		}
	}
	return env
}

func readDotEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || strings.ContainsAny(key, " \t") {
			continue
		}
		values[key] = cleanDotEnvValue(value)
	}
	return values, scanner.Err()
}

func cleanDotEnvValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		quote := trimmed[0]
		if (quote == '"' || quote == '\'') && trimmed[len(trimmed)-1] == quote {
			return trimmed[1 : len(trimmed)-1]
		}
	}
	if before, _, ok := strings.Cut(trimmed, " #"); ok {
		return strings.TrimSpace(before)
	}
	return trimmed
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
