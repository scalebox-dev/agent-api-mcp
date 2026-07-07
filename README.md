# Agent API MCP Server

Hosted Model Context Protocol server for Agent API.

This service exposes Agent API as MCP tools, resources, and prompts so MCP
clients can create and inspect agent runs, work with durable volumes, search
memory, and use platform catalogs without speaking Agent API directly.

## Runtime

The production runtime is Go. The default transport is hosted Streamable HTTP,
intended to run next to or behind Agent API's public gateway.

```text
MCP client -> agent-api-mcp -> Agent API HTTP gateway -> Agent API services
```

Incoming `Authorization: Bearer ...` headers are forwarded to Agent API. For
local development or service-key deployments, `AGENT_API_KEY` can be set as a
fallback credential.

## Build

```bash
go build ./cmd/agent-api-mcp
go test ./...
```

## Run

```bash
AGENT_API_BASE_URL=https://api.agentsway.dev \
AGENT_API_MCP_ADDR=:8080 \
go run ./cmd/agent-api-mcp
```

Endpoints:

- `POST /mcp` - MCP Streamable HTTP endpoint
- `GET /mcp` - MCP Streamable HTTP event stream endpoint when applicable
- `GET /healthz` - liveness
- `GET /readyz` - readiness/config sanity

## Initial Capability Map

- Agent runtime: create, retrieve, cancel, list events, list children.
- Catalog: list models, presets, and callable Agent API tools.
- Volumes: list/create volumes, list/search entries, read/write files, grep,
  summarize, and patch text lines.
- Memory: search long-term memory with thread or workspace scope.
- Skills: list, retrieve, discover, focus, read/write files, diff, and promote
  dev branches.

The current implementation is an early Go skeleton for the hosted service. The
tool/resource surface is intentionally broad but still thin over Agent API HTTP;
the next step is to harden auth, observability, schema coverage, and deployment.
