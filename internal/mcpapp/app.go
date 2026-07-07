package mcpapp

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/config"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
)

type App struct {
	Config config.Config
	Logger *slog.Logger
	Client *agentapi.Client
}

func NewServer(cfg config.Config, logger *slog.Logger, requestAuth string) *mcp.Server {
	defaultHeaders := map[string]string{}
	if auth := strings.TrimSpace(requestAuth); auth != "" {
		defaultHeaders["Authorization"] = auth
	}
	app := &App{
		Config: cfg,
		Logger: logger,
		Client: agentapi.NewClient(&agentapi.ClientOptions{
			APIKey:         cfg.AgentAPIKey,
			BaseURL:        cfg.AgentAPIBaseURL,
			Timeout:        cfg.HTTPTimeout,
			DefaultHeaders: defaultHeaders,
		}),
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-api-mcp",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: "Hosted MCP server for Agent API. Use the exposed tools and resources to create, inspect, and operate Agent API responses, volumes, memory, and skills.",
		Logger:       logger,
	})

	app.registerCatalogTools(server)
	app.registerResponseTools(server)
	app.registerVolumeTools(server)
	app.registerMemoryTools(server)
	app.registerSkillTools(server)
	app.registerResources(server)
	app.registerPrompts(server)
	return server
}

func AuthHeaderFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	auth := strings.TrimSpace(req.Header.Get("Authorization"))
	if auth == "" {
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return ""
	}
	return auth
}

func require(value string, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func boolPtr(value bool) *bool {
	if !value {
		return nil
	}
	return &value
}

func readOnlyTool(openWorld bool) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		ReadOnlyHint:  true,
		OpenWorldHint: &openWorld,
	}
}

func mutatingTool(destructive bool, openWorld bool) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		DestructiveHint: &destructive,
		OpenWorldHint:   &openWorld,
	}
}

func destructiveTool(openWorld bool) *mcp.ToolAnnotations {
	destructive := true
	return &mcp.ToolAnnotations{
		DestructiveHint: &destructive,
		OpenWorldHint:   &openWorld,
	}
}
