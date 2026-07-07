package mcpapp

import (
	"context"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
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
		out, err := a.Client.Models.List(ctx)
		if err != nil {
			return nil, err
		}
		return ResourceJSON(req.Params.URI, out)
	case "agentapi://presets":
		out, err := a.Client.Presets.List(ctx)
		if err != nil {
			return nil, err
		}
		return ResourceJSON(req.Params.URI, out)
	case "agentapi://tools":
		out, err := a.Client.Tools.List(ctx)
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

	var out any
	switch {
	case len(parts) == 2 && parts[0] == "responses":
		out, err = a.Client.Responses.Retrieve(ctx, parts[1])
	case len(parts) == 3 && parts[0] == "responses" && parts[2] == "events":
		out, err = a.Client.Responses.ListEvents(ctx, parts[1], agentapi.ListEventsParams{})
	case len(parts) == 3 && parts[0] == "responses" && parts[2] == "children":
		out, err = a.Client.Responses.ListChildren(ctx, parts[1])
	case len(parts) >= 4 && parts[0] == "volumes" && parts[2] == "files":
		out, err = a.Client.Volumes.ReadFile(ctx, parts[1], strings.Join(parts[3:], "/"), agentapi.ReadFileParams{})
	case len(parts) >= 4 && parts[0] == "skills" && parts[2] == "files":
		out, err = a.Client.Skills.ReadFile(ctx, parts[1], strings.Join(parts[3:], "/"), agentapi.ReadSkillFileParams{})
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
