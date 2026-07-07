package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
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
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_search_memories", Title: "Search Memories", Description: "Search Agent API long-term memory with thread or workspace scope.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in searchMemoriesInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.Query, "query"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Memories.Search(ctx, agentapi.MemorySearchParams(in))
			return ToolResult(out, err)
		})
}
