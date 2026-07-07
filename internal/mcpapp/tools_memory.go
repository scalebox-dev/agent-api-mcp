package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type searchMemoriesInput struct {
	Query              string  `json:"query"`
	Limit              int     `json:"limit,omitempty"`
	PreviousResponseID string  `json:"previous_response_id,omitempty"`
	TenantSearch       bool    `json:"tenant_search,omitempty"`
	Lang               string  `json:"lang,omitempty"`
	SemanticWeight     float64 `json:"semantic_weight,omitempty"`
}

func (a *App) registerMemoryTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_search_memories", Description: "Search Agent API long-term memory with thread or workspace scope."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in searchMemoriesInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.Query, "query"); err != nil {
				return nil, nil, err
			}
			out, err := a.post(ctx, "/v1/memories/search", in)
			return JSONText(out), out, err
		})
}
