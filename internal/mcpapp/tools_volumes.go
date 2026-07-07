package mcpapp

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
)

type listVolumesInput struct {
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type createVolumeInput struct {
	Name string `json:"name,omitempty"`
}

type volumeIDInput struct {
	VolumeID string `json:"volume_id"`
}

type updateVolumeInput struct {
	VolumeID string `json:"volume_id"`
	Name     string `json:"name,omitempty"`
}

type volumePathInput struct {
	VolumeID string `json:"volume_id"`
	Path     string `json:"path,omitempty"`
}

type listVolumeEntriesInput struct {
	VolumeID  string `json:"volume_id"`
	Path      string `json:"path,omitempty"`
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
	VolumeID  string `json:"volume_id"`
	Pattern   string `json:"pattern"`
	Path      string `json:"path,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

func (a *App) registerVolumeTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_volumes", Title: "List Volumes", Description: "List durable Agent API volumes.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listVolumesInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Volumes.List(ctx, agentapi.ListParams{
				Limit:     in.Limit,
				PageToken: in.PageToken,
				UserID:    in.UserID,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_volume", Title: "Get Volume", Description: "Retrieve durable Agent API volume metadata.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumeIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.Retrieve(ctx, in.VolumeID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_create_volume", Title: "Create Volume", Description: "Create a durable Agent API volume.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createVolumeInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Volumes.Create(ctx, in.Name)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_update_volume", Title: "Update Volume", Description: "Update durable Agent API volume metadata such as the display name.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in updateVolumeInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.UpdateWithParams(ctx, in.VolumeID, agentapi.UpdateVolumeParams{Name: in.Name})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_delete_volume", Title: "Delete Volume", Description: "Delete a durable Agent API volume.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumeIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			err := a.Client.Volumes.Delete(ctx, in.VolumeID)
			return ToolResult(map[string]any{"deleted": err == nil, "volume_id": in.VolumeID}, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_reconcile_volume_usage", Title: "Reconcile Volume Usage", Description: "Recompute volume usage accounting for a durable Agent API volume.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumeIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.ReconcileUsage(ctx, in.VolumeID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_volume_entries", Title: "List Volume Entries", Description: "List files and directories inside a volume.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listVolumeEntriesInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.ListEntries(ctx, in.VolumeID, agentapi.VolumeEntriesParams{
				Path:      in.Path,
				Limit:     in.Limit,
				PageToken: in.PageToken,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_search_volume_entries", Title: "Search Volume Entries", Description: "Search file and directory paths inside a volume.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in searchVolumeEntriesInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.SearchEntries(ctx, in.VolumeID, agentapi.VolumeEntriesParams{
				Query:     in.Query,
				Path:      in.Path,
				Limit:     in.Limit,
				PageToken: in.PageToken,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_create_volume_directory", Title: "Create Volume Directory", Description: "Create a directory inside a durable Agent API volume.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumePathInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.CreateDirectory(ctx, in.VolumeID, in.Path)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_delete_volume_path", Title: "Delete Volume Path", Description: "Delete a file or directory path inside a durable Agent API volume.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumePathInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.DeletePath(ctx, in.VolumeID, in.Path)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_read_volume_file", Title: "Read Volume File", Description: "Read a volume file.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in readVolumeFileInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			params := agentapi.ReadFileParams{MaxBytes: in.MaxBytes}
			if strings.EqualFold(in.Format, "raw") {
				out, err := a.Client.Volumes.ReadFileRaw(ctx, in.VolumeID, in.Path, params)
				return ToolResult(out, err)
			}
			out, err := a.Client.Volumes.ReadFile(ctx, in.VolumeID, in.Path, params)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_write_volume_file", Title: "Write Volume File", Description: "Write text content to a volume file, replacing existing content at the target path.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in writeVolumeFileInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.WriteFile(ctx, in.VolumeID, in.Path, []byte(in.Content))
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_read_volume_lines", Title: "Read Volume Lines", Description: "Read a line range from a text file in a volume.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in readVolumeLinesInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.ReadLines(ctx, in.VolumeID, in.Path, agentapi.ReadLinesParams{
				StartLine: in.StartLine,
				EndLine:   in.EndLine,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_patch_volume_lines", Title: "Patch Volume Lines", Description: "Replace a line range in a volume text file.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in patchVolumeLinesInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.PatchLines(ctx, in.VolumeID, in.Path, agentapi.PatchLinesParams{
				StartLine:   in.StartLine,
				EndLine:     in.EndLine,
				Replacement: in.Replacement,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_grep_volume", Title: "Grep Volume", Description: "Search text content inside volume files.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in grepVolumeInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Pattern, "pattern"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.Grep(ctx, in.VolumeID, agentapi.VolumeEntriesParams{
				Query:     in.Pattern,
				Path:      in.Path,
				Limit:     in.Limit,
				PageToken: in.PageToken,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_download_volume_archive", Title: "Download Volume Archive", Description: "Download a volume subtree as an archive encoded as base64 content.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumePathInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.DownloadArchive(ctx, in.VolumeID, in.Path)
			if err != nil {
				return ToolResult(nil, err)
			}
			return ToolResult(binaryArchivePayload(out.Path, out.ContentType, out.Content), nil)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_summarize_volume", Title: "Summarize Volume", Description: "Summarize volume contents and text previews.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in volumePathInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.VolumeID, "volume_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Volumes.Summarize(ctx, in.VolumeID, in.Path)
			return ToolResult(out, err)
		})
}
