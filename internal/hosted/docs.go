package hosted

import (
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/scalebox-dev/agent-api-mcp/internal/config"
)

type docsTool struct {
	Category string
	Name     string
	Use      string
	Mode     string
}

var docsTools = []docsTool{
	{"Agent responses", "agent_api_create_response", "Run an Agent API task, continue a thread, attach volumes, choose presets/models, and enable memory or skills.", "Write"},
	{"Agent responses", "agent_api_list_responses", "List stored responses visible to the credential.", "Read"},
	{"Agent responses", "agent_api_get_response", "Retrieve a persisted response by id.", "Read"},
	{"Agent responses", "agent_api_cancel_response", "Cancel an in-flight response.", "Write"},
	{"Agent responses", "agent_api_list_response_events", "Inspect response timeline and audit events.", "Read"},
	{"Agent responses", "agent_api_list_child_responses", "List delegated sub-agent runs for a parent response.", "Read"},
	{"Agent responses", "agent_api_get_response_volume", "Resolve the durable volume associated with a response.", "Read"},
	{"Catalog", "agent_api_list_models", "List available model ids and metadata.", "Read"},
	{"Catalog", "agent_api_list_presets", "List managed presets such as search and research workflows.", "Read"},
	{"Catalog", "agent_api_list_tools", "List callable platform tools known to Agent API.", "Read"},
	{"Volumes", "agent_api_list_volumes", "List durable Agent API volumes.", "Read"},
	{"Volumes", "agent_api_get_volume", "Retrieve durable volume metadata.", "Read"},
	{"Volumes", "agent_api_create_volume", "Create a durable volume.", "Write"},
	{"Volumes", "agent_api_update_volume", "Update durable volume metadata.", "Write"},
	{"Volumes", "agent_api_delete_volume", "Delete a durable volume.", "Write"},
	{"Volumes", "agent_api_reconcile_volume_usage", "Recompute volume usage accounting.", "Write"},
	{"Volumes", "agent_api_list_volume_entries", "List files and directories inside a volume.", "Read"},
	{"Volumes", "agent_api_search_volume_entries", "Search file and directory paths inside a volume.", "Read"},
	{"Volumes", "agent_api_create_volume_directory", "Create a directory inside a volume.", "Write"},
	{"Volumes", "agent_api_delete_volume_path", "Delete a file or directory path inside a volume.", "Write"},
	{"Volumes", "agent_api_read_volume_file", "Read a volume file.", "Read"},
	{"Volumes", "agent_api_write_volume_file", "Write text content to a volume file.", "Write"},
	{"Volumes", "agent_api_read_volume_lines", "Read a line range from a text file.", "Read"},
	{"Volumes", "agent_api_patch_volume_lines", "Replace a line range in a text file.", "Write"},
	{"Volumes", "agent_api_grep_volume", "Search text content inside volume files.", "Read"},
	{"Volumes", "agent_api_download_volume_archive", "Download a volume subtree as a base64 archive.", "Read"},
	{"Volumes", "agent_api_summarize_volume", "Summarize volume contents and text previews.", "Read"},
	{"Memory", "agent_api_search_memories", "Search long-term memory with thread or workspace scope.", "Read"},
	{"Skills", "agent_api_list_skills", "List workspace skills.", "Read"},
	{"Skills", "agent_api_create_skill", "Create skill metadata.", "Write"},
	{"Skills", "agent_api_get_skill", "Retrieve skill metadata.", "Read"},
	{"Skills", "agent_api_update_skill", "Update skill metadata.", "Write"},
	{"Skills", "agent_api_archive_skill", "Archive a skill.", "Write"},
	{"Skills", "agent_api_delete_skill", "Delete a skill.", "Write"},
	{"Skills", "agent_api_discover_skills", "Discover relevant skills for a task or query.", "Read"},
	{"Skills", "agent_api_focus_skills", "Load selected skill manifests and files for model context.", "Read"},
	{"Skills", "agent_api_create_skill_dev", "Create or update a skill dev branch.", "Write"},
	{"Skills", "agent_api_update_skill_files", "Apply primitive file updates to skill dev branches.", "Write"},
	{"Skills", "agent_api_list_skill_files", "List files in a skill branch.", "Read"},
	{"Skills", "agent_api_read_skill_file", "Read a skill branch file.", "Read"},
	{"Skills", "agent_api_write_skill_file", "Write a skill branch file.", "Write"},
	{"Skills", "agent_api_delete_skill_file", "Delete a file from a skill branch.", "Write"},
	{"Skills", "agent_api_export_skill_archive", "Export skill files as a base64 archive.", "Read"},
	{"Skills", "agent_api_import_skill_archive", "Import a base64 archive into a skill branch.", "Write"},
	{"Skills", "agent_api_diff_skill", "Diff skill main and dev branches.", "Read"},
	{"Skills", "agent_api_accept_skill_dev", "Promote a skill dev branch to main.", "Write"},
	{"Skills", "agent_api_discard_skill_dev", "Discard a skill dev branch.", "Write"},
}

var docsResources = []string{
	"agentapi://models",
	"agentapi://presets",
	"agentapi://tools",
	"agentapi://responses/{response_id}",
	"agentapi://responses/{response_id}/events",
	"agentapi://responses/{response_id}/children",
	"agentapi://volumes/{volume_id}/files/{path}",
	"agentapi://skills/{skill_id}/files/{path}",
}

var docsPrompts = []string{
	"research_with_agent",
	"continue_agent_thread",
	"debug_agent_response",
}

func renderDocsHTML(cfg config.Config, req *http.Request) string {
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
	agentAPIURL := html.EscapeString(cfg.AgentAPIBaseURL)

	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description" content="Connect MCP clients to Agent API responses, search, volumes, memory, and skills.">
  <title>Agent API MCP Server Documentation</title>
  <script>
    (function () {
      try {
        var key = "agent-api-theme";
        var params = new URLSearchParams(window.location.search);
        var requested = params.get("theme");
        var cookie = document.cookie.split("; ").find(function (part) {
          return part.indexOf(encodeURIComponent(key) + "=") === 0;
        });
        var stored = requested || window.localStorage.getItem(key) || (cookie ? decodeURIComponent(cookie.split("=").slice(1).join("=")) : "") || "system";
        var effective = stored === "dark" || (stored === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches) ? "dark" : "light";
        document.documentElement.classList.add(effective);
        document.documentElement.dataset.themeMode = stored === "light" || stored === "dark" || stored === "system" ? stored : "system";
      } catch (_) {
        document.documentElement.classList.add("light");
      }
    })();
  </script>
  <style>
    :root {
      color-scheme: light;
      --bg: #f8fafc;
      --panel: #ffffff;
      --panel-muted: #f8fafc;
      --text: #020617;
      --muted: #475569;
      --subtle: #64748b;
      --border: #e2e8f0;
      --soft: #f1f5f9;
      --accent: #2563eb;
      --accent-soft: #dbeafe;
      --code-bg: #020617;
      --code-text: #f8fafc;
      --ok: #047857;
      --ok-bg: #d1fae5;
      --warn: #b45309;
      --warn-bg: #ffedd5;
      --shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
    }
    .dark {
      color-scheme: dark;
      --bg: #020617;
      --panel: #0f172a;
      --panel-muted: #111827;
      --text: #f8fafc;
      --muted: #cbd5e1;
      --subtle: #94a3b8;
      --border: #1e293b;
      --soft: #1e293b;
      --accent: #60a5fa;
      --accent-soft: rgba(37, 99, 235, 0.18);
      --code-bg: #020617;
      --code-text: #f8fafc;
      --ok: #6ee7b7;
      --ok-bg: rgba(16, 185, 129, 0.16);
      --warn: #fbbf24;
      --warn-bg: rgba(245, 158, 11, 0.16);
      --shadow: none;
    }
    * { box-sizing: border-box; }
    html { scroll-behavior: smooth; }
    body {
      margin: 0;
      background: var(--bg);
      color: var(--text);
      font: 15px/1.55 Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    a { color: var(--accent); text-decoration-thickness: 1px; text-underline-offset: 3px; }
    .site-header {
      background: var(--panel);
      border-bottom: 1px solid var(--border);
    }
    .wrap { max-width: 1280px; margin: 0 auto; padding: 0 24px; }
    .topbar {
      display: flex;
      min-height: 64px;
      align-items: center;
      justify-content: space-between;
      gap: 18px;
    }
    .brand {
      color: var(--text);
      font-weight: 700;
      text-decoration: none;
    }
    .brand span { color: var(--subtle); font-weight: 500; }
    .toplinks {
      display: flex;
      align-items: center;
      gap: 14px;
      color: var(--subtle);
      font-size: 14px;
    }
    .toplinks a { color: var(--subtle); text-decoration: none; }
    .toplinks a:hover { color: var(--text); }
    .theme-toggle {
      display: inline-flex;
      align-items: center;
      gap: 2px;
      padding: 3px;
      border: 1px solid var(--border);
      border-radius: 12px;
      background: var(--panel);
    }
    .theme-toggle button {
      border: 0;
      border-radius: 9px;
      background: transparent;
      color: var(--subtle);
      cursor: pointer;
      font: inherit;
      font-size: 12px;
      padding: 5px 8px;
    }
    .theme-toggle button[aria-pressed="true"] {
      background: var(--soft);
      color: var(--text);
    }
    .docs-shell {
      display: flex;
      gap: 40px;
      padding-top: 56px;
      padding-bottom: 64px;
    }
    .toc {
      width: 210px;
      flex: 0 0 210px;
      position: sticky;
      top: 24px;
      max-height: calc(100vh - 48px);
      overflow-y: auto;
      align-self: flex-start;
      color: var(--subtle);
      font-size: 14px;
    }
    .toc strong {
      display: block;
      margin-bottom: 12px;
      color: var(--subtle);
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }
    .toc a {
      display: block;
      border-left: 1px solid var(--border);
      color: var(--subtle);
      padding: 5px 0 5px 13px;
      text-decoration: none;
      transition: border-color 120ms ease, color 120ms ease;
    }
    .toc a:hover { border-color: var(--accent); color: var(--text); }
    article { min-width: 0; flex: 1; }
    .hero {
      border-bottom: 1px solid var(--border);
      padding-bottom: 40px;
    }
    .eyebrow {
      color: var(--accent);
      font-size: 14px;
      font-weight: 700;
      letter-spacing: 0.16em;
      margin: 0 0 12px;
      text-transform: uppercase;
    }
    h1 { margin: 0; font-size: clamp(32px, 5vw, 48px); font-weight: 650; line-height: 1.08; letter-spacing: 0; max-width: 900px; }
    .lead { margin: 16px 0 0; max-width: 780px; color: var(--muted); font-size: 17px; line-height: 1.7; }
    .quick {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 14px;
      margin-top: 24px;
    }
    .quick div, .panel {
      background: var(--panel);
      border: 1px solid var(--border);
      border-radius: 14px;
      box-shadow: var(--shadow);
    }
    .quick div { padding: 16px; }
    .quick strong { display: block; margin-bottom: 4px; }
    .quick span { color: var(--muted); font-size: 14px; }
    section {
      border-bottom: 1px solid var(--border);
      padding: 48px 0;
      scroll-margin-top: 24px;
    }
    section:last-child { border-bottom: 0; }
    .panel { padding: 0; background: transparent; border: 0; box-shadow: none; }
    h2 { margin: 0 0 10px; font-size: 22px; font-weight: 650; letter-spacing: 0; }
    h3 { margin: 26px 0 8px; font-size: 16px; font-weight: 650; letter-spacing: 0; }
    p { margin: 8px 0; color: var(--muted); }
    ul, ol { margin: 10px 0 0 22px; padding: 0; color: var(--muted); }
    li { margin: 4px 0; }
    code, pre { font: 13px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
    code { color: var(--text); background: var(--soft); border-radius: 5px; padding: 1px 4px; }
    pre {
      margin: 12px 0 0;
      overflow-x: auto;
      color: var(--code-text);
      background: var(--code-bg);
      border: 1px solid #1e293b;
      border-radius: 10px;
      padding: 16px;
    }
    .endpoint {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
      margin-top: 10px;
      padding: 13px 14px;
      border: 1px solid var(--border);
      border-radius: 12px;
      background: var(--panel);
      box-shadow: var(--shadow);
    }
    .endpoint span {
      min-width: 140px;
      color: var(--subtle);
      font-size: 13px;
      font-weight: 600;
    }
    .pill {
      display: inline-flex;
      align-items: center;
      min-height: 24px;
      padding: 2px 8px;
      border-radius: 5px;
      background: var(--accent-soft);
      color: var(--muted);
      font-size: 12px;
      font-weight: 700;
    }
    .pill.write { color: var(--warn); background: var(--warn-bg); }
    .pill.read { color: var(--ok); background: var(--ok-bg); }
    table {
      width: 100%;
      border-collapse: collapse;
      margin-top: 12px;
      border: 1px solid var(--border);
      background: var(--panel);
      border-radius: 12px;
      overflow: hidden;
      box-shadow: var(--shadow);
    }
    th, td { text-align: left; vertical-align: top; padding: 10px 12px; border-bottom: 1px solid var(--border); }
    th { background: var(--panel-muted); color: var(--text); font-size: 13px; font-weight: 650; }
    td { color: var(--muted); }
    tr:last-child td { border-bottom: 0; }
    td code { white-space: nowrap; }
    .cards { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; margin-top: 12px; }
    .mini { padding: 16px; background: var(--panel); border: 1px solid var(--border); border-radius: 12px; box-shadow: var(--shadow); }
    .mini strong { display: block; margin-bottom: 4px; }
    .mini p { margin: 0; }
    footer { color: var(--subtle); padding: 0 24px 32px; max-width: 1280px; margin: 0 auto; }
    @media (max-width: 900px) {
      .topbar { align-items: flex-start; flex-direction: column; padding: 16px 0; }
      .toplinks { width: 100%; justify-content: space-between; }
      .docs-shell { display: block; padding-top: 32px; }
      .toc { display: none; }
      .quick, .cards { grid-template-columns: 1fr; }
      h1 { font-size: 32px; }
    }
  </style>
</head>
<body>
  <header class="site-header">
    <div class="wrap topbar">
      <a class="brand" href="/docs">Agent API <span>/ MCP</span></a>
      <div class="toplinks">
        <a href="` + agentAPIURL + `">Agent API</a>
        <a href="` + metadataURL + `">OAuth metadata</a>
        <div class="theme-toggle" aria-label="Theme">
          <button type="button" data-theme-option="light">Light</button>
          <button type="button" data-theme-option="dark">Dark</button>
          <button type="button" data-theme-option="system">System</button>
        </div>
      </div>
    </div>
  </header>
  <main class="wrap docs-shell">
    <nav class="toc" aria-label="Documentation navigation">
      <strong>On this page</strong>
      <a href="#start">Quick Start</a>
      <a href="#auth">Authentication</a>
      <a href="#workflow">Typical Workflow</a>
      <a href="#capabilities">Capabilities</a>
      <a href="#tools">Tool Reference</a>
      <a href="#resources">Resources and Prompts</a>
      <a href="#operations">Operations</a>
    </nav>
    <article>
        <header class="hero">
          <p class="eyebrow">Managed Agent API</p>
          <h1>Agent API MCP Server Documentation</h1>
          <p class="lead">Use Agent API from any MCP-compatible client. This hosted MCP server exposes responses, search presets, model catalogs, durable volumes, memory, and skills through Streamable HTTP.</p>
          <div class="quick">
            <div><strong>Transport</strong><span>Streamable HTTP at <code>` + mcpURL + `</code></span></div>
            <div><strong>Auth</strong><span>Caller-supplied Agent API bearer token forwarded upstream.</span></div>
            <div><strong>Best for</strong><span>Research agents, coding assistants, workspace automation, and MCP-native integrations.</span></div>
          </div>
        </header>
        <section id="start" class="panel">
          <h2>Quick Start</h2>
          <p>Configure your MCP client with the endpoint below and attach your Agent API key as an HTTP bearer token.</p>
          <div class="endpoint"><span>MCP endpoint</span><code>` + mcpURL + `</code></div>
          <div class="endpoint"><span>Auth header</span><code>Authorization: Bearer &lt;agent-api-token&gt;</code></div>
          <div class="endpoint"><span>Agent API upstream</span><code>` + agentAPIURL + `</code></div>
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
          <p>Use the exact header mechanism supported by your MCP client. If your client manages secrets separately, store the token there instead of hardcoding it in a config file.</p>
        </section>

        <section id="auth" class="panel">
          <h2>Authentication Model</h2>
          <p>The MCP server requires a bearer token on MCP requests and forwards that token to Agent API. Workspace access, quotas, billing, rate limits, and authorization are enforced by Agent API.</p>
          <ul>
            <li>No server-side managed Agent API key is used for public callers.</li>
            <li>Unauthenticated MCP requests return <code>401</code> with a protected-resource metadata challenge.</li>
            <li>Protected-resource metadata is available at <a href="` + metadataURL + `">` + metadataURL + `</a>.</li>
            <li>Authorization server: <code>` + authServer + `</code>.</li>
          </ul>
        </section>

        <section id="workflow" class="panel">
          <h2>Typical Workflow</h2>
          <ol>
            <li>List presets or models to understand the available execution modes.</li>
            <li>Create a response with <code>agent_api_create_response</code>, usually with a preset such as <code>pro-search</code> for search-backed research.</li>
            <li>Retrieve the response, inspect events, and resolve the response volume if artifacts were produced.</li>
            <li>Use volume, memory, and skill tools to continue work across longer tasks.</li>
          </ol>
          <pre>{
  "tool": "agent_api_create_response",
  "arguments": {
    "input": "What is the latest FIFA 2026 World Cup news?",
    "preset": "pro-search",
    "max_output_tokens": 512,
    "max_steps": 4,
    "store": true
  }
}</pre>
        </section>

        <section id="capabilities" class="panel">
          <h2>Capabilities</h2>
          <div class="cards">
            <div class="mini"><strong>Run agents</strong><p>Create and continue Agent API responses with presets, models, reasoning, memory, skills, and volumes.</p></div>
            <div class="mini"><strong>Inspect execution</strong><p>Retrieve responses, timeline events, child runs, usage, status, errors, and associated volumes.</p></div>
            <div class="mini"><strong>Work with files</strong><p>Create volumes, read/write files, patch line ranges, grep content, and summarize artifacts.</p></div>
            <div class="mini"><strong>Use workspace context</strong><p>Search memories, discover/focus skills, inspect skill files, and promote skill dev branches.</p></div>
          </div>
        </section>

        <section id="tools" class="panel">
          <h2>Tool Reference</h2>
          <p>The server exposes ` + toolCountHTML() + ` MCP tools. Read tools are safe discovery and inspection operations. Write tools create, mutate, cancel, or promote state in Agent API.</p>
          ` + toolTableHTML() + `
        </section>

        <section id="resources" class="panel">
          <h2>Resources and Prompts</h2>
          <p>Resources expose common Agent API objects as JSON URIs. Prompts help MCP clients orchestrate repeatable research, continuation, and debugging workflows.</p>
          <h3>Resources</h3>
          <pre>` + escapedLines(docsResources) + `</pre>
          <h3>Prompts</h3>
          <pre>` + escapedLines(docsPrompts) + `</pre>
        </section>

        <section id="operations" class="panel">
          <h2>Operations</h2>
          <p>Use these endpoints for monitoring and MCP auth discovery.</p>
          <div class="endpoint"><span>Liveness</span><a href="/healthz">/healthz</a></div>
          <div class="endpoint"><span>Readiness</span><a href="/readyz">/readyz</a></div>
          <div class="endpoint"><span>Protected resource</span><a href="` + metadataURL + `">` + metadataURL + `</a></div>
          <p>The canonical MCP protocol endpoint is <code>` + mcpURL + `</code>. The documentation page remains available at <code>/docs</code>, and the service root redirects here for human visitors.</p>
        </section>
    </article>
  </main>
  <footer class="wrap">Agent API MCP Server. Use Agent API credentials only with clients and environments you trust.</footer>
  <script>
    (function () {
      var key = "agent-api-theme";
      var root = document.documentElement;
      var media = window.matchMedia("(prefers-color-scheme: dark)");
      function readMode() {
        return root.dataset.themeMode || "system";
      }
      function effective(mode) {
        return mode === "dark" || (mode === "system" && media.matches) ? "dark" : "light";
      }
      function writeCookie(mode) {
        document.cookie = encodeURIComponent(key) + "=" + encodeURIComponent(mode) + "; Path=/; Max-Age=34560000; SameSite=Lax" + (location.protocol === "https:" ? "; Secure" : "");
      }
      function apply(mode, persist) {
        if (mode !== "light" && mode !== "dark" && mode !== "system") mode = "system";
        root.classList.remove("light", "dark");
        root.classList.add(effective(mode));
        root.dataset.themeMode = mode;
        if (persist) {
          window.localStorage.setItem(key, mode);
          writeCookie(mode);
        }
        document.querySelectorAll("[data-theme-option]").forEach(function (button) {
          button.setAttribute("aria-pressed", button.getAttribute("data-theme-option") === mode ? "true" : "false");
        });
      }
      document.querySelectorAll("[data-theme-option]").forEach(function (button) {
        button.addEventListener("click", function () {
          apply(button.getAttribute("data-theme-option") || "system", true);
        });
      });
      media.addEventListener("change", function () {
        if (readMode() === "system") apply("system", false);
      });
      window.addEventListener("message", function (event) {
        var theme = event.data && (event.data.theme || event.data.agentApiTheme);
        if (event.data && event.data.type === "agent-api-theme" && theme) apply(theme, true);
      });
      apply(readMode(), false);
    })();
  </script>
</body>
</html>`
}

func toolCountHTML() string {
	return html.EscapeString(strconv.Itoa(len(docsTools)))
}

func toolTableHTML() string {
	var b strings.Builder
	b.WriteString(`<table><thead><tr><th>Category</th><th>Tool</th><th>Mode</th><th>Use</th></tr></thead><tbody>`)
	for _, tool := range docsTools {
		modeClass := "read"
		if tool.Mode == "Write" {
			modeClass = "write"
		}
		b.WriteString(`<tr><td>`)
		b.WriteString(html.EscapeString(tool.Category))
		b.WriteString(`</td><td><code>`)
		b.WriteString(html.EscapeString(tool.Name))
		b.WriteString(`</code></td><td><span class="pill `)
		b.WriteString(modeClass)
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(tool.Mode))
		b.WriteString(`</span></td><td>`)
		b.WriteString(html.EscapeString(tool.Use))
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func escapedLines(values []string) string {
	return html.EscapeString(strings.Join(values, "\n"))
}
