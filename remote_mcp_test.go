package agentapimcp_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRemoteMCPDeployedServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	session := connectRemoteSession(t, ctx)
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	if len(tools.Tools) == 0 {
		t.Fatal("tools/list returned no tools")
	}
	for _, name := range []string{
		"agent_api_list_models",
		"agent_api_get_volume",
		"agent_api_delete_volume_path",
		"agent_api_create_skill",
		"agent_api_import_skill_archive",
	} {
		if !hasTool(tools, name) {
			t.Fatalf("tools/list missing %s; got %d tools", name, len(tools.Tools))
		}
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

func TestRemoteMCPAgentResponseE2E(t *testing.T) {
	requireRemoteTestLevel(t, "agent")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	session := connectRemoteSession(t, ctx)
	defer session.Close()

	created := callToolJSON(t, ctx, session, "agent_api_create_response", map[string]any{
		"input":             "What is the latest FIFA 2026 World Cup news?",
		"preset":            "pro-search",
		"max_output_tokens": 512,
		"max_steps":         4,
		"store":             true,
	})
	id, _ := created["id"].(string)
	if !strings.HasPrefix(id, "resp_") {
		t.Fatalf("agent_api_create_response returned id %q", id)
	}
	if object, _ := created["object"].(string); object != "" && object != "response" {
		t.Fatalf("agent_api_create_response object = %q", object)
	}

	retrieved := callToolJSON(t, ctx, session, "agent_api_get_response", map[string]any{
		"response_id": id,
	})
	if got, _ := retrieved["id"].(string); got != id {
		t.Fatalf("agent_api_get_response id = %q, want %q", got, id)
	}
}

func requireRemoteTestLevel(t *testing.T, minimum string) {
	t.Helper()

	env := loadTestEnv(t, ".test.env")
	level := strings.ToLower(firstNonEmpty(os.Getenv("AGENT_API_MCP_TEST_LEVEL"), env["AGENT_API_MCP_TEST_LEVEL"], "smoke"))
	if !remoteTestLevelEnabled(level, minimum) {
		t.Skipf("set AGENT_API_MCP_TEST_LEVEL=%s or full to run this remote MCP workload; current level is %s", minimum, level)
	}
}

func remoteTestLevelEnabled(level string, minimum string) bool {
	rank := map[string]int{
		"smoke": 1,
		"agent": 2,
		"full":  3,
	}
	return rank[level] >= rank[minimum]
}

func connectRemoteSession(t *testing.T, ctx context.Context) *mcp.ClientSession {
	t.Helper()

	env := loadTestEnv(t, ".test.env")
	apiKey := firstNonEmpty(os.Getenv("AGENT_API_KEY"), env["AGENT_API_KEY"])
	if apiKey == "" {
		t.Skip("set AGENT_API_KEY in .test.env or process env to run remote MCP integration test")
	}

	endpoint, err := remoteMCPEndpoint(env)
	if err != nil {
		t.Fatal(err)
	}

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
	return session
}

func callToolJSON(t *testing.T, ctx context.Context, session *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("tools/call %s: %v", name, err)
	}
	body := textContent(result.Content)
	if result.IsError {
		t.Fatalf("%s returned MCP tool error: %s", name, body)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("%s returned non-JSON content: %v: %s", name, err, body)
	}
	return out
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
		Timeout:   5 * time.Minute,
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
