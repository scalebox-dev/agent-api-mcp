package hosted

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/config"
	"github.com/scalebox-dev/agent-api-mcp/internal/mcpapp"
)

func NewServer(cfg config.Config, logger *slog.Logger) (*http.Server, error) {
	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return mcpapp.NewServer(cfg, logger, mcpapp.AuthHeaderFromRequest(req))
	}, &mcp.StreamableHTTPOptions{
		Stateless:      true,
		JSONResponse:   true,
		Logger:         logger,
		SessionTimeout: cfg.SessionTimeout,
	})

	mux.Handle(cfg.MCPPath, requireBearerAuth(mcpHandler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !cfg.Ready() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":                 true,
			"agent_api_base_url": cfg.AgentAPIBaseURL,
			"mcp_path":           cfg.MCPPath,
		})
	})

	return &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}, nil
}

func requireBearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if mcpapp.AuthHeaderFromRequest(req) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": map[string]any{
					"code":    "unauthorized",
					"message": "Authorization bearer token is required",
				},
			})
			return
		}
		next.ServeHTTP(w, req)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
