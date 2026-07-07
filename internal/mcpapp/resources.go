package mcpapp

import (
	"context"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/agentapi"
)

func (a *App) registerResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{URI: "agentapi://models", Name: "Agent API Models", MIMEType: "application/json"}, a.readStaticResource)
	server.AddResource(&mcp.Resource{URI: "agentapi://presets", Name: "Agent API Presets", MIMEType: "application/json"}, a.readStaticResource)
	server.AddResource(&mcp.Resource{URI: "agentapi://tools", Name: "Agent API Tools", MIMEType: "application/json"}, a.readStaticResource)
	server.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "agentapi://responses/{response_id}"}, a.readDynamicResource)
	server.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "agentapi://responses/{response_id}/events"}, a.readDynamicResource)
	server.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "agentapi://responses/{response_id}/children"}, a.readDynamicResource)
	server.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "agentapi://volumes/{volume_id}/files/{path}"}, a.readDynamicResource)
	server.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "agentapi://skills/{skill_id}/files/{path}"}, a.readDynamicResource)
}

func (a *App) readStaticResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	switch req.Params.URI {
	case "agentapi://models":
		out, err := a.get(ctx, "/v1/models", nil)
		if err != nil {
			return nil, err
		}
		return ResourceJSON(req.Params.URI, out)
	case "agentapi://presets":
		out, err := a.get(ctx, "/v1/presets", nil)
		if err != nil {
			return nil, err
		}
		return ResourceJSON(req.Params.URI, out)
	case "agentapi://tools":
		out, err := a.get(ctx, "/v1/tools", nil)
		if err != nil {
			return nil, err
		}
		return ResourceJSON(req.Params.URI, out)
	default:
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
}

func (a *App) readDynamicResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	u, err := url.Parse(req.Params.URI)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "agentapi" {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	parts := splitResourcePath(u)
	if len(parts) == 0 {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	var out map[string]any
	switch {
	case len(parts) == 2 && parts[0] == "responses":
		out, err = a.get(ctx, agentapi.Join("v1", "responses", agentapi.Segment(parts[1])), nil)
	case len(parts) == 3 && parts[0] == "responses" && parts[2] == "events":
		out, err = a.get(ctx, agentapi.Join("v1", "responses", agentapi.Segment(parts[1]), "events"), nil)
	case len(parts) == 3 && parts[0] == "responses" && parts[2] == "children":
		out, err = a.get(ctx, agentapi.Join("v1", "responses", agentapi.Segment(parts[1]), "children"), nil)
	case len(parts) >= 4 && parts[0] == "volumes" && parts[2] == "files":
		out, err = a.get(ctx, agentapi.Join("v1", "volumes", agentapi.Segment(parts[1]), "files", agentapi.Segment(strings.Join(parts[3:], "/"))), nil)
	case len(parts) >= 4 && parts[0] == "skills" && parts[2] == "files":
		out, err = a.get(ctx, agentapi.Join("v1", "skills", agentapi.Segment(parts[1]), "files", agentapi.Segment(strings.Join(parts[3:], "/"))), nil)
	default:
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	if err != nil {
		return nil, err
	}
	return ResourceJSON(req.Params.URI, out)
}

func splitResourcePath(u *url.URL) []string {
	raw := strings.Trim(strings.TrimPrefix(u.Host+u.Path, "/"), "/")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if unescaped, err := url.PathUnescape(part); err == nil && unescaped != "" {
			out = append(out, unescaped)
		}
	}
	return out
}
