package mcpapp

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
)

func JSONText(value any) *mcp.CallToolResult {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		raw = []byte(`{"error":"failed to encode result"}`)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}},
	}
}

func ToolResult(value any, err error) (*mcp.CallToolResult, any, error) {
	if err != nil {
		return ToolErrorResult(err), nil, nil
	}
	return JSONText(value), value, nil
}

func ToolErrorResult(err error) *mcp.CallToolResult {
	payload := normalizeToolError(err)
	raw, marshalErr := json.MarshalIndent(payload, "", "  ")
	if marshalErr != nil {
		raw = []byte(`{"ok":false,"error":{"message":"tool call failed"}}`)
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(raw)}},
		StructuredContent: payload,
		IsError:           true,
	}
}

func normalizeToolError(err error) map[string]any {
	payload := map[string]any{
		"ok": false,
		"error": map[string]any{
			"message": "tool call failed",
		},
	}
	if err == nil {
		return payload
	}

	detail := payload["error"].(map[string]any)
	detail["message"] = err.Error()

	var rateLimit *agentapi.RateLimitError
	if errors.As(err, &rateLimit) && rateLimit != nil {
		detail["kind"] = "rate_limit"
		if rateLimit.RetryAfter > 0 {
			detail["retry_after_ms"] = int64(rateLimit.RetryAfter / time.Millisecond)
		}
		addAPIErrorFields(detail, rateLimit.APIError)
		return payload
	}

	var apiErr *agentapi.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		detail["kind"] = "agent_api_error"
		addAPIErrorFields(detail, apiErr)
		return payload
	}

	var connErr *agentapi.APIConnectionError
	if errors.As(err, &connErr) {
		detail["kind"] = "agent_api_connection_error"
		return payload
	}

	detail["kind"] = "tool_error"
	return payload
}

func addAPIErrorFields(detail map[string]any, apiErr *agentapi.APIError) {
	if apiErr == nil {
		return
	}
	if apiErr.Status > 0 {
		detail["status"] = apiErr.Status
	}
	if strings.TrimSpace(apiErr.Code) != "" {
		detail["code"] = apiErr.Code
	}
	if strings.TrimSpace(apiErr.Type) != "" {
		detail["type"] = apiErr.Type
	}
	if strings.TrimSpace(apiErr.RequestID) != "" {
		detail["request_id"] = apiErr.RequestID
	}
}

func ResourceJSON(uri string, value any) (*mcp.ReadResourceResult, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(raw),
		}},
	}, nil
}
