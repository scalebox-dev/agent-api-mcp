package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/agentapi"
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

type discoverSkillsInput struct {
	Query      string           `json:"query,omitempty"`
	Skills     []map[string]any `json:"skills,omitempty"`
	MaxResults int              `json:"max_results,omitempty"`
}

type focusSkillsInput struct {
	Skills           []map[string]any `json:"skills"`
	FallbackToMain   bool             `json:"fallback_to_main,omitempty"`
	MaxManifestChars int              `json:"max_manifest_chars,omitempty"`
	MaxFileChars     int              `json:"max_file_chars,omitempty"`
}

type skillFileInput struct {
	SkillID        string `json:"skill_id"`
	Path           string `json:"path"`
	Branch         string `json:"branch,omitempty"`
	FallbackToMain bool   `json:"fallback_to_main,omitempty"`
	MaxBytes       int    `json:"max_bytes,omitempty"`
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

func (a *App) registerSkillTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_skills", Description: "List workspace skills."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listSkillsInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, "/v1/skills", agentapi.Query(map[string]any{"limit": in.Limit, "page_token": in.PageToken, "archived": in.Archived, "user_id": in.UserID}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_skill", Description: "Retrieve skill metadata."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillIDInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, agentapi.Join("v1", "skills", agentapi.Segment(in.SkillID)), nil)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_discover_skills", Description: "Discover relevant skills for a task or query."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in discoverSkillsInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.post(ctx, "/v1/skills/discover", in)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_focus_skills", Description: "Load selected skill manifests and files for model context."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in focusSkillsInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.post(ctx, "/v1/skills/focus", in)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_skill_files", Description: "List files in a skill branch."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillFileInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, agentapi.Join("v1", "skills", agentapi.Segment(in.SkillID), "files"), agentapi.Query(map[string]any{
				"path": in.Path, "branch": in.Branch, "fallback_to_main": in.FallbackToMain,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_read_skill_file", Description: "Read a file from a skill branch."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in skillFileInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, agentapi.Join("v1", "skills", agentapi.Segment(in.SkillID), "files", agentapi.Segment(in.Path)), agentapi.Query(map[string]any{
				"branch": in.Branch, "fallback_to_main": in.FallbackToMain, "max_bytes": in.MaxBytes,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_write_skill_file", Description: "Write text content to a skill branch file."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in writeSkillFileInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.Client.PutRaw(ctx, agentapi.Join("v1", "skills", agentapi.Segment(in.SkillID), "files", agentapi.Segment(in.Path))+"?branch="+agentapi.Segment(in.Branch), in.Content)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_diff_skill", Description: "Diff skill main and dev branches."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in diffSkillInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, agentapi.Join("v1", "skills", agentapi.Segment(in.SkillID), "diff"), agentapi.Query(map[string]any{
				"path": in.Path, "max_file_chars": in.MaxFileChars, "include_unchanged": in.IncludeUnchanged,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_accept_skill_dev", Description: "Promote a skill dev branch to main."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in acceptSkillInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.post(ctx, agentapi.Join("v1", "skills", agentapi.Segment(in.SkillID), "accept_dev")+"?strategy="+agentapi.Segment(in.Strategy), map[string]any{})
			return JSONText(out), out, err
		})
}
