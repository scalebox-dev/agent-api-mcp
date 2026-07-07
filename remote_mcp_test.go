package agentapimcp_test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRemoteMCPDeployedServer(t *testing.T) {
	env := loadTestEnv(t, ".test.env")
	apiKey := firstNonEmpty(os.Getenv("AGENT_API_KEY"), env["AGENT_API_KEY"])
	if apiKey == "" {
		t.Skip("set AGENT_API_KEY in .test.env or process env to run remote MCP integration test")
	}

	endpoint, err := remoteMCPEndpoint(env)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "agent-api-mcp-remote-test",
		Version: "0.1.0",
	}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             endpoint,
		HTTPClient:           bearerHTTPClient(apiKey),
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect remote MCP endpoint %s: %v", endpoint, err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	if len(tools.Tools) == 0 {
		t.Fatal("tools/list returned no tools")
	}
	if !hasTool(tools, "agent_api_list_models") {
		t.Fatalf("tools/list missing agent_api_list_models; got %d tools", len(tools.Tools))
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "agent_api_list_models"})
	if err != nil {
		t.Fatalf("tools/call agent_api_list_models: %v", err)
	}
	if result.IsError {
		t.Fatalf("agent_api_list_models returned MCP tool error: %s", textContent(result.Content))
	}
	if got := textContent(result.Content); !strings.Contains(got, `"object"`) && !strings.Contains(got, `"data"`) {
		t.Fatalf("agent_api_list_models returned unexpected content: %s", got)
	}
}

func loadTestEnv(t *testing.T, path string) map[string]string {
	t.Helper()

	values := map[string]string{}
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return values
	}
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = cleanEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return values
}

func remoteMCPEndpoint(env map[string]string) (string, error) {
	endpoint := firstNonEmpty(
		os.Getenv("AGENT_API_MCP_TEST_URL"),
		env["AGENT_API_MCP_TEST_URL"],
	)
	if endpoint == "" {
		return "", fmt.Errorf("set AGENT_API_MCP_TEST_URL in .test.env or process env")
	}
	if !strings.HasSuffix(strings.TrimRight(endpoint, "/"), "/mcp") {
		return "", fmt.Errorf("AGENT_API_MCP_TEST_URL must point to the MCP endpoint, e.g. https://mcp.example.com/mcp")
	}
	return strings.TrimRight(endpoint, "/"), nil
}

func bearerHTTPClient(token string) *http.Client {
	return &http.Client{
		Timeout:   65 * time.Second,
		Transport: bearerRoundTripper{token: token, base: http.DefaultTransport},
	}
}

type bearerRoundTripper struct {
	token string
	base  http.RoundTripper
}

func (rt bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	if clone.Header.Get("Authorization") == "" {
		clone.Header.Set("Authorization", "Bearer "+rt.token)
	}
	return rt.base.RoundTrip(clone)
}

func hasTool(result *mcp.ListToolsResult, name string) bool {
	for _, tool := range result.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func textContent(content []mcp.Content) string {
	var out strings.Builder
	for _, item := range content {
		text, ok := item.(*mcp.TextContent)
		if !ok {
			continue
		}
		if out.Len() > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(text.Text)
	}
	return out.String()
}

func cleanEnvValue(value string) string {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
