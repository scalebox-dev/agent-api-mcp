package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *App) registerCatalogTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_models", Description: "List currently available Agent API model ids and metadata."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Models.List(ctx)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_presets", Description: "List managed Agent API presets and public policy metadata."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Presets.List(ctx)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_tools", Description: "List callable platform tools known to Agent API."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Tools.List(ctx)
			return JSONText(out), out, err
		})
}
