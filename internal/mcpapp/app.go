package mcpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/agentapi"
	"github.com/scalebox-dev/agent-api-mcp/internal/config"
)

type App struct {
	Config config.Config
	Logger *slog.Logger
	Client *agentapi.Client
}

func NewServer(cfg config.Config, logger *slog.Logger, requestAuth string) *mcp.Server {
	auth := agentapi.PreferRequestAuth(requestAuth, cfg.AgentAPIKey)
	app := &App{
		Config: cfg,
		Logger: logger,
		Client: agentapi.New(cfg.AgentAPIBaseURL, auth, cfg.HTTPTimeout),
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

func (a *App) get(ctx context.Context, endpoint string, query map[string]string) (map[string]any, error) {
	return a.Client.Get(ctx, endpoint, query)
}

func (a *App) post(ctx context.Context, endpoint string, body any) (map[string]any, error) {
	return a.Client.Post(ctx, endpoint, body)
}

func require(value string, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
