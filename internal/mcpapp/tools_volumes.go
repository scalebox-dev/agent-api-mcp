package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/agentapi"
)

type listVolumesInput struct {
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type createVolumeInput struct {
	Name string `json:"name,omitempty"`
}

type volumePathInput struct {
	VolumeID string `json:"volume_id"`
	Path     string `json:"path,omitempty"`
}

type listVolumeEntriesInput struct {
	VolumeID  string `json:"volume_id"`
	Path      string `json:"path,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

type searchVolumeEntriesInput struct {
	VolumeID  string `json:"volume_id"`
	Query     string `json:"query"`
	Path      string `json:"path,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

type readVolumeFileInput struct {
	VolumeID string `json:"volume_id"`
	Path     string `json:"path"`
	Format   string `json:"format,omitempty"`
	MaxBytes int    `json:"max_bytes,omitempty"`
}

type writeVolumeFileInput struct {
	VolumeID string `json:"volume_id"`
	Path     string `json:"path"`
	Content  string `json:"content"`
}

type readVolumeLinesInput struct {
	VolumeID  string `json:"volume_id"`
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line,omitempty"`
}

type patchVolumeLinesInput struct {
	VolumeID    string `json:"volume_id"`
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

type grepVolumeInput struct {
	VolumeID      string `json:"volume_id"`
	Pattern       string `json:"pattern"`
	Path          string `json:"path,omitempty"`
	CaseSensitive bool   `json:"case_sensitive,omitempty"`
	MaxMatches    int    `json:"max_matches,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	PageToken     string `json:"page_token,omitempty"`
}

func (a *App) registerVolumeTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_volumes", Description: "List durable Agent API volumes."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listVolumesInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, "/v1/volumes", agentapi.Query(map[string]any{"limit": in.Limit, "page_token": in.PageToken, "user_id": in.UserID}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_create_volume", Description: "Create a durable Agent API volume."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createVolumeInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.post(ctx, "/v1/volumes", in)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_volume_entries", Description: "List files and directories inside a volume."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listVolumeEntriesInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.get(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "entries"), agentapi.Query(map[string]any{
				"path": in.Path, "recursive": in.Recursive, "limit": in.Limit, "page_token": in.PageToken,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_search_volume_entries", Description: "Search file and directory paths inside a volume."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in searchVolumeEntriesInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.get(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "search"), agentapi.Query(map[string]any{
				"query": in.Query, "path": in.Path, "limit": in.Limit, "page_token": in.PageToken,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_read_volume_file", Description: "Read a volume file."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in readVolumeFileInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.get(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "files", agentapi.Segment(in.Path)), agentapi.Query(map[string]any{
				"format": in.Format, "max_bytes": in.MaxBytes,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_write_volume_file", Description: "Write text content to a volume file."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in writeVolumeFileInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.PutRaw(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "files", agentapi.Segment(in.Path)), in.Content)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_read_volume_lines", Description: "Read a line range from a text file in a volume."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in readVolumeLinesInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "file_lines", agentapi.Segment(in.Path)), agentapi.Query(map[string]any{
				"start_line": in.StartLine, "end_line": in.EndLine,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_patch_volume_lines", Description: "Replace a line range in a volume text file."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in patchVolumeLinesInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.Client.Patch(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "file_lines", agentapi.Segment(in.Path)), map[string]any{
				"start_line": in.StartLine, "end_line": in.EndLine, "replacement": in.Replacement,
			})
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_grep_volume", Description: "Search text content inside volume files."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in grepVolumeInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "grep"), agentapi.Query(map[string]any{
				"pattern": in.Pattern, "path": in.Path, "case_sensitive": in.CaseSensitive, "max_matches": in.MaxMatches, "limit": in.Limit, "page_token": in.PageToken,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_summarize_volume", Description: "Summarize volume contents and text previews."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumePathInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.post(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(in.VolumeID), "summarize"), map[string]any{"path": in.Path})
			return JSONText(out), out, err
		})
}
