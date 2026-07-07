package hosted

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/scalebox-dev/agent-api-mcp/internal/config"
)

type recordedAgentAPIRequest struct {
	Method        string
	Path          string
	RawQuery      string
	Authorization string
	Body          map[string]any
}

type fakeAgentAPI struct {
	server   *httptest.Server
	mu       sync.Mutex
	requests []recordedAgentAPIRequest
}

func newFakeAgentAPI(t *testing.T) *fakeAgentAPI {
	t.Helper()
	api := &fakeAgentAPI{}
	api.server = httptest.NewServer(http.HandlerFunc(api.handle))
	t.Cleanup(api.server.Close)
	return api
}

func (api *fakeAgentAPI) URL() string {
	return api.server.URL
}

func (api *fakeAgentAPI) Requests() []recordedAgentAPIRequest {
	api.mu.Lock()
	defer api.mu.Unlock()
	out := make([]recordedAgentAPIRequest, len(api.requests))
	copy(out, api.requests)
	return out
}

func (api *fakeAgentAPI) handle(w http.ResponseWriter, req *http.Request) {
	var body map[string]any
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		if len(bytes.TrimSpace(raw)) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
	}

	api.mu.Lock()
	api.requests = append(api.requests, recordedAgentAPIRequest{
		Method:        req.Method,
		Path:          req.URL.Path,
		RawQuery:      req.URL.RawQuery,
		Authorization: req.Header.Get("Authorization"),
		Body:          body,
	})
	api.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	switch {
	case req.Method == http.MethodGet && req.URL.Path == "/v1/models":
		_, _ = io.WriteString(w, `{"object":"list","data":[{"id":"test/model","name":"Test Model"}]}`)
	case req.Method == http.MethodPost && req.URL.Path == "/v1/responses":
		_, _ = io.WriteString(w, `{"id":"resp_123","object":"response","created_at":1,"status":"completed","model":"test/model","output":[],"output_text":"done"}`)
	case req.Method == http.MethodGet && req.URL.Path == "/v1/volumes/vol_123/files/notes.txt":
		_, _ = io.WriteString(w, `{"path":"notes.txt","encoding":"utf-8","mime_type":"text/plain","size":5,"truncated":false,"content":"hello"}`)
	default:
		http.Error(w, `{"error":"unexpected fake Agent API request"}`, http.StatusNotFound)
	}
}

func TestMCPToolUsesFallbackAPIKey(t *testing.T) {
	api := newFakeAgentAPI(t)
	mcpServer := newTestMCPServer(t, api.URL(), "sk_fallback")

	result := callMCPTool(t, mcpServer.URL+"/mcp", "", "agent_api_list_models", map[string]any{})
	text := toolText(t, result)
	if !bytes.Contains([]byte(text), []byte("test/model")) {
		t.Fatalf("tool result text = %s", text)
	}

	requests := api.Requests()
	if len(requests) != 1 {
		t.Fatalf("Agent API requests = %d, want 1", len(requests))
	}
	if got, want := requests[0].Authorization, "Bearer sk_fallback"; got != want {
		t.Fatalf("Agent API Authorization = %q, want %q", got, want)
	}
	if got, want := requests[0].Method+" "+requests[0].Path, "GET /v1/models"; got != want {
		t.Fatalf("Agent API request = %q, want %q", got, want)
	}
}

func TestMCPToolsForwardRequestAuthorizationAndCallSDK(t *testing.T) {
	api := newFakeAgentAPI(t)
	mcpServer := newTestMCPServer(t, api.URL(), "sk_fallback")

	create := callMCPTool(t, mcpServer.URL+"/mcp", "Bearer user_token", "agent_api_create_response", map[string]any{
		"input":        "hello",
		"model":        "test/model",
		"metadata":     map[string]any{"source": "mcp-test"},
		"max_steps":    3,
		"volume_id":    "vol_123",
		"instructions": "Be concise.",
	})
	if text := toolText(t, create); !bytes.Contains([]byte(text), []byte("resp_123")) {
		t.Fatalf("create response result text = %s", text)
	}

	read := callMCPTool(t, mcpServer.URL+"/mcp", "Bearer user_token", "agent_api_read_volume_file", map[string]any{
		"volume_id": "vol_123",
		"path":      "notes.txt",
		"max_bytes": 64,
	})
	if text := toolText(t, read); !bytes.Contains([]byte(text), []byte("hello")) {
		t.Fatalf("read file result text = %s", text)
	}

	requests := api.Requests()
	if len(requests) != 2 {
		t.Fatalf("Agent API requests = %d, want 2", len(requests))
	}
	for i, req := range requests {
		if got, want := req.Authorization, "Bearer user_token"; got != want {
			t.Fatalf("request %d Authorization = %q, want %q", i, got, want)
		}
	}
	if got, want := requests[0].Method+" "+requests[0].Path, "POST /v1/responses"; got != want {
		t.Fatalf("create request = %q, want %q", got, want)
	}
	if got := requests[0].Body["stream"]; got == true {
		t.Fatalf("create body stream = %#v, want non-streaming request", got)
	}
	if got := requests[0].Body["input"]; got != "hello" {
		t.Fatalf("create body input = %#v, want hello", got)
	}
	if got, want := requests[1].Method+" "+requests[1].Path, "GET /v1/volumes/vol_123/files/notes.txt"; got != want {
		t.Fatalf("read request = %q, want %q", got, want)
	}
	if got, want := requests[1].RawQuery, "max_bytes=64"; got != want {
		t.Fatalf("read query = %q, want %q", got, want)
	}
}

func newTestMCPServer(t *testing.T, agentAPIBaseURL string, apiKey string) *httptest.Server {
	t.Helper()
	cfg := config.Load([]string{
		"AGENT_API_BASE_URL=" + agentAPIBaseURL,
		"AGENT_API_KEY=" + apiKey,
		"AGENT_API_MCP_PATH=/mcp",
	})
	server, err := NewServer(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(server.Handler)
	t.Cleanup(ts.Close)
	return ts
}

func callMCPTool(t *testing.T, endpoint string, authorization string, name string, arguments map[string]any) map[string]any {
	t.Helper()
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal MCP request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new MCP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST tools/call: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST tools/call status = %d body=%s", resp.StatusCode, raw)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode MCP response: %v", err)
	}
	if payload["error"] != nil {
		t.Fatalf("MCP error response: %#v", payload["error"])
	}
	result, ok := payload["result"].(map[string]any)
	if !ok {
		t.Fatalf("MCP response missing result: %#v", payload)
	}
	return result
}

func toolText(t *testing.T, result map[string]any) string {
	t.Helper()
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("tool result missing content: %#v", result)
	}
	first, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("tool result content[0] = %#v", content[0])
	}
	text, ok := first["text"].(string)
	if !ok {
		t.Fatalf("tool result content[0].text missing: %#v", first)
	}
	return text
}
