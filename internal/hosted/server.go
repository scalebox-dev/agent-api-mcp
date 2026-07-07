package hosted

import (
	"encoding/json"
	"html"
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
	mux.HandleFunc("/", docsRedirectHandler)
	mux.HandleFunc("/docs", docsHandler(cfg))
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

func docsRedirectHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	http.Redirect(w, req, "/docs", http.StatusFound)
}

func docsHandler(cfg config.Config) http.HandlerFunc {
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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if req.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(docsHTML(cfg, req)))
	}
}

func docsHTML(cfg config.Config, req *http.Request) string {
	origin := publicOrigin(cfg, req)
	mcpURL := origin + cfg.MCPPath
	metadataURL := origin + protectedResourceWellKnownPath
	authServer := strings.TrimRight(strings.TrimSpace(cfg.AuthorizationServerURL), "/")
	if authServer == "" {
		authServer = cfg.AgentAPIBaseURL
	}
	mcpURL = html.EscapeString(mcpURL)
	metadataURL = html.EscapeString(metadataURL)
	authServer = html.EscapeString(authServer)
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Agent API MCP Server</title>
  <style>
    :root { color-scheme: light dark; --border: #d0d7de; --muted: #57606a; --accent: #0969da; }
    body { margin: 0; font: 16px/1.5 system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    main { max-width: 860px; margin: 0 auto; padding: 48px 20px 64px; }
    h1 { margin: 0 0 8px; font-size: 34px; line-height: 1.15; }
    h2 { margin: 32px 0 8px; font-size: 20px; }
    p { margin: 8px 0; color: var(--muted); }
    code, pre { font: 14px/1.45 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
    pre { overflow-x: auto; padding: 14px 16px; border: 1px solid var(--border); border-radius: 8px; }
    a { color: var(--accent); }
    .endpoint { margin-top: 20px; padding: 14px 16px; border: 1px solid var(--border); border-radius: 8px; }
  </style>
</head>
<body>
  <main>
    <h1>Agent API MCP Server</h1>
    <p>Use Agent API capabilities from MCP-compatible clients with your Agent API bearer token.</p>

    <div class="endpoint">
      <strong>MCP endpoint</strong>
      <pre>` + mcpURL + `</pre>
    </div>

    <h2>Authentication</h2>
    <p>Send your Agent API credential as an HTTP bearer token. The MCP server forwards it to Agent API and does not use a managed backend key.</p>
    <pre>Authorization: Bearer &lt;agent-api-token&gt;</pre>

    <h2>Client Configuration</h2>
    <p>Use Streamable HTTP transport and configure the endpoint above in your MCP client.</p>
    <pre>{
  "mcpServers": {
    "agent-api": {
      "url": "` + mcpURL + `",
      "headers": {
        "Authorization": "Bearer &lt;agent-api-token&gt;"
      }
    }
  }
}</pre>

    <h2>Discovery</h2>
    <p>Protected-resource metadata is available at <a href="` + metadataURL + `">` + metadataURL + `</a>.</p>
    <p>Authorization server: <code>` + authServer + `</code></p>

    <h2>Health</h2>
    <p><a href="/healthz">/healthz</a> and <a href="/readyz">/readyz</a> are available for operational checks.</p>
  </main>
</body>
</html>`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
