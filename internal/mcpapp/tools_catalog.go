package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *App) registerCatalogTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_models", Title: "List Models", Description: "List currently available Agent API model ids and metadata.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Models.List(ctx)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_presets", Title: "List Presets", Description: "List managed Agent API presets and public policy metadata.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Presets.List(ctx)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_tools", Title: "List Platform Tools", Description: "List callable platform tools known to Agent API.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Tools.List(ctx)
			return ToolResult(out, err)
		})
}
