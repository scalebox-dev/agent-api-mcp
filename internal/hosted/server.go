package hosted

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/config"
	"github.com/scalebox-dev/agent-api-mcp/internal/mcpapp"
)

const protectedResourceWellKnownPath = "/.well-known/oauth-protected-resource"

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

	mux.Handle(cfg.MCPPath, requireBearerAuth(cfg, mcpHandler))
	mux.HandleFunc(protectedResourceWellKnownPath, protectedResourceMetadataHandler(cfg))
	if pathSpecific := protectedResourceWellKnownPath + cfg.MCPPath; pathSpecific != protectedResourceWellKnownPath {
		mux.HandleFunc(pathSpecific, protectedResourceMetadataHandler(cfg))
	}
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

func requireBearerAuth(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if mcpapp.AuthHeaderFromRequest(req) == "" {
			w.Header().Set("WWW-Authenticate", bearerChallenge(cfg, req))
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

func protectedResourceMetadataHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
				"error": map[string]any{
					"code":    "method_not_allowed",
					"message": "method must be GET or HEAD",
				},
			})
			return
		}

		value := protectedResourceMetadata(cfg, req)
		if req.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, value)
	}
}

func protectedResourceMetadata(cfg config.Config, req *http.Request) map[string]any {
	origin := publicOrigin(cfg, req)
	resource := origin + cfg.MCPPath
	authServer := strings.TrimRight(strings.TrimSpace(cfg.AuthorizationServerURL), "/")

	metadata := map[string]any{
		"resource":                 resource,
		"resource_name":            "Agent API MCP Server",
		"bearer_methods_supported": []string{"header"},
	}
	if authServer != "" {
		metadata["authorization_servers"] = []string{authServer}
	}
	return metadata
}

func bearerChallenge(cfg config.Config, req *http.Request) string {
	return `Bearer resource_metadata="` + publicOrigin(cfg, req) + protectedResourceWellKnownPath + `"`
}

func publicOrigin(cfg config.Config, req *http.Request) string {
	if cfg.MCPPublicBaseURL != "" {
		return cfg.MCPPublicBaseURL
	}
	if req == nil {
		return ""
	}
	host := firstForwarded(req.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = req.Host
	}
	proto := firstForwarded(req.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		proto = "http"
		if req.TLS != nil {
			proto = "https"
		}
	}
	return strings.TrimRight(proto+"://"+host, "/")
}

func firstForwarded(value string) string {
	if before, _, ok := strings.Cut(value, ","); ok {
		value = before
	}
	return strings.TrimSpace(value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
