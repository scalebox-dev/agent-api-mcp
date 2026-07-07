package mcpapp

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
