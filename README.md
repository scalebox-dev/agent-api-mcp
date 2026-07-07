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

Incoming `Authorization: Bearer ...` headers are required and forwarded to
Agent API. The MCP server does not provide or fall back to a managed Agent API
key.

## Build

```bash
go build ./cmd/agent-api-mcp
go test ./...
```

## Run

```bash
cp .env.example .env
# edit .env if needed
go run ./cmd/agent-api-mcp
```

Process environment values override `.env` values, so hosted deployments should
inject real environment variables rather than mounting a `.env` file.

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

## Environment

The supported service environment variables are tracked in `.env.example`.

- `AGENT_API_BASE_URL` - Agent API upstream base URL.
- `AGENT_API_MCP_ADDR` - hosted HTTP bind address.
- `AGENT_API_MCP_PATH` - Streamable HTTP MCP endpoint path.
- `AGENT_API_HTTP_TIMEOUT_MS` - Agent API SDK request timeout.
- `AGENT_API_MCP_SESSION_TIMEOUT_MS` - MCP session timeout.

## Initial Capability Map

- Agent runtime: create, retrieve, cancel, list events, list children.
- Catalog: list models, presets, and callable Agent API tools.
- Volumes: list/create volumes, list/search entries, read/write files, grep,
  summarize, and patch text lines.
- Memory: search long-term memory with thread or workspace scope.
- Skills: list, retrieve, discover, focus, read/write files, diff, and promote
  dev branches.

The current implementation is a hosted Go foundation backed by the official
Agent API Go SDK. The tool/resource surface is intentionally broad and now
includes initial MCP safety annotations; the next step is to keep hardening
error handling, observability, schema coverage, and deployment packaging.
