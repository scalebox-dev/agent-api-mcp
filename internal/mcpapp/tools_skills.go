package mcpapp

import (
	"context"
	"encoding/base64"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
)

type listSkillsInput struct {
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	Archived  bool   `json:"archived,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type skillIDInput struct {
	SkillID string `json:"skill_id"`
}

type createSkillInput struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type updateSkillInput struct {
	SkillID     string         `json:"skill_id"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type createSkillDevInput struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description,omitempty"`
	Metadata    map[string]any               `json:"metadata,omitempty"`
	Files       []agentapi.SkillFileMutation `json:"files,omitempty"`
}

type updateSkillFilesInput struct {
	Updates []agentapi.SkillFileUpdateMutation `json:"updates"`
}

type discoverSkillsInput struct {
	Query              string                          `json:"query,omitempty"`
	Branch             string                          `json:"branch,omitempty"`
	IncludeDev         bool                            `json:"include_dev,omitempty"`
	Limit              int                             `json:"limit,omitempty"`
	PreviousResponseID string                          `json:"previous_response_id,omitempty"`
	TenantSearch       bool                            `json:"tenant_search,omitempty"`
	LocalSkills        []agentapi.LocalSkillDescriptor `json:"local_skills,omitempty"`
}

type focusSkillsInput struct {
	Skills           []agentapi.SkillFocusItem `json:"skills"`
	FallbackToMain   bool                      `json:"fallback_to_main,omitempty"`
	MaxManifestChars int                       `json:"max_manifest_chars,omitempty"`
	MaxFileChars     int                       `json:"max_file_chars,omitempty"`
}

type skillFileInput struct {
	SkillID        string `json:"skill_id"`
	Path           string `json:"path"`
	Branch         string `json:"branch,omitempty"`
	FallbackToMain bool   `json:"fallback_to_main,omitempty"`
	MaxBytes       int    `json:"max_bytes,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	PageToken      string `json:"page_token,omitempty"`
}

type writeSkillFileInput struct {
	SkillID string `json:"skill_id"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Branch  string `json:"branch,omitempty"`
}

type diffSkillInput struct {
	SkillID          string `json:"skill_id"`
	Path             string `json:"path,omitempty"`
	MaxFileChars     int    `json:"max_file_chars,omitempty"`
	IncludeUnchanged bool   `json:"include_unchanged,omitempty"`
}

type acceptSkillInput struct {
	SkillID  string `json:"skill_id"`
	Strategy string `json:"strategy,omitempty" jsonschema:"patch or mirror."`
}

type deleteSkillFileInput struct {
	SkillID string `json:"skill_id"`
	Path    string `json:"path"`
	Branch  string `json:"branch,omitempty"`
}

type skillArchiveInput struct {
	SkillID        string `json:"skill_id"`
	Path           string `json:"path,omitempty"`
	Branch         string `json:"branch,omitempty"`
	FallbackToMain bool   `json:"fallback_to_main,omitempty"`
}

type importSkillArchiveInput struct {
	SkillID          string `json:"skill_id"`
	ContentBase64    string `json:"content_base64"`
	Path             string `json:"path,omitempty"`
	Branch           string `json:"branch,omitempty"`
	Replace          bool   `json:"replace,omitempty"`
	StripTopLevelDir bool   `json:"strip_top_level_dir,omitempty"`
}

func (a *App) registerSkillTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_skills", Title: "List Skills", Description: "List workspace skills.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listSkillsInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Skills.List(ctx, agentapi.ListSkillsParams{
				IncludeArchived: in.Archived,
				Limit:           in.Limit,
				PageToken:       in.PageToken,
				UserID:          in.UserID,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_create_skill", Title: "Create Skill", Description: "Create a workspace skill metadata record.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createSkillInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Skills.Create(ctx, agentapi.CreateSkillParams{
				Name:        in.Name,
				Description: in.Description,
				Metadata:    in.Metadata,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_skill", Title: "Get Skill", Description: "Retrieve skill metadata.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.Retrieve(ctx, in.SkillID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_update_skill", Title: "Update Skill", Description: "Update workspace skill metadata.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in updateSkillInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.Update(ctx, in.SkillID, agentapi.UpdateSkillParams{
				Name:        in.Name,
				Description: in.Description,
				Metadata:    in.Metadata,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_archive_skill", Title: "Archive Skill", Description: "Archive a workspace skill.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.Archive(ctx, in.SkillID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_delete_skill", Title: "Delete Skill", Description: "Delete a workspace skill.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.Delete(ctx, in.SkillID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_discover_skills", Title: "Discover Skills", Description: "Discover relevant skills for a task or query.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in discoverSkillsInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Skills.Discover(ctx, agentapi.DiscoverSkillsParams{
				Query:              in.Query,
				Branch:             in.Branch,
				IncludeDev:         in.IncludeDev,
				Limit:              in.Limit,
				PreviousResponseID: in.PreviousResponseID,
				TenantSearch:       in.TenantSearch,
				LocalSkills:        in.LocalSkills,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_focus_skills", Title: "Focus Skills", Description: "Load selected skill manifests and files for model context.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in focusSkillsInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Skills.Focus(ctx, agentapi.FocusSkillParams{
				Skills:           in.Skills,
				FallbackToMain:   boolPtr(in.FallbackToMain),
				MaxManifestChars: in.MaxManifestChars,
				MaxFileChars:     in.MaxFileChars,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_create_skill_dev", Title: "Create Skill Dev", Description: "Create or update a skill development branch from initial metadata and files.", Annotations: mutatingTool(false, false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createSkillDevInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.Name, "name"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.CreateDev(ctx, agentapi.CreateSkillDevParams{
				Name:        in.Name,
				Description: in.Description,
				Metadata:    in.Metadata,
				Files:       in.Files,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_update_skill_files", Title: "Update Skill Files", Description: "Apply primitive file updates to one or more skill dev branches.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in updateSkillFilesInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Skills.UpdateFile(ctx, agentapi.UpdateSkillFilePrimitiveParams{Updates: in.Updates})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_skill_files", Title: "List Skill Files", Description: "List files in a skill branch.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillFileInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.ListFiles(ctx, in.SkillID, agentapi.ListSkillFilesParams{
				Path:           in.Path,
				Branch:         in.Branch,
				FallbackToMain: boolPtr(in.FallbackToMain),
				Limit:          in.Limit,
				PageToken:      in.PageToken,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_read_skill_file", Title: "Read Skill File", Description: "Read a file from a skill branch.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillFileInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.ReadFile(ctx, in.SkillID, in.Path, agentapi.ReadSkillFileParams{
				Branch:         in.Branch,
				FallbackToMain: boolPtr(in.FallbackToMain),
				MaxBytes:       in.MaxBytes,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_write_skill_file", Title: "Write Skill File", Description: "Write text content to a skill branch file, replacing existing content at the target path.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in writeSkillFileInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.WriteFile(ctx, in.SkillID, in.Path, []byte(in.Content), in.Branch)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_delete_skill_file", Title: "Delete Skill File", Description: "Delete a file from a skill branch.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in deleteSkillFileInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.Path, "path"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.DeleteFile(ctx, in.SkillID, in.Path, in.Branch)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_export_skill_archive", Title: "Export Skill Archive", Description: "Export skill files as an archive encoded as base64 content.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillArchiveInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.ExportArchive(ctx, in.SkillID, agentapi.SkillArchiveParams{
				Path:           in.Path,
				Branch:         in.Branch,
				FallbackToMain: boolPtr(in.FallbackToMain),
			})
			if err != nil {
				return ToolResult(nil, err)
			}
			return ToolResult(binaryArchivePayload(out.Path, out.ContentType, out.Content), nil)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_import_skill_archive", Title: "Import Skill Archive", Description: "Import a base64-encoded archive into a skill branch.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in importSkillArchiveInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			if err := require(in.ContentBase64, "content_base64"); err != nil {
				return nil, nil, err
			}
			content, err := base64.StdEncoding.DecodeString(in.ContentBase64)
			if err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.ImportArchive(ctx, in.SkillID, content, agentapi.ImportSkillArchiveParams{
				Path:             in.Path,
				Branch:           in.Branch,
				Replace:          in.Replace,
				StripTopLevelDir: boolPtr(in.StripTopLevelDir),
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_diff_skill", Title: "Diff Skill", Description: "Diff skill main and dev branches.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in diffSkillInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.Diff(ctx, in.SkillID, agentapi.SkillBranchDiffParams{
				Path:             in.Path,
				MaxFileChars:     in.MaxFileChars,
				IncludeUnchanged: in.IncludeUnchanged,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_accept_skill_dev", Title: "Accept Skill Dev", Description: "Promote a skill dev branch to main.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in acceptSkillInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.AcceptDev(ctx, in.SkillID, in.Strategy)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_discard_skill_dev", Title: "Discard Skill Dev", Description: "Discard a skill development branch.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.SkillID, "skill_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Skills.DiscardDev(ctx, in.SkillID)
			return ToolResult(out, err)
		})
}
