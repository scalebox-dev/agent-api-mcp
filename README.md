# Agent API MCP Server

First-class Model Context Protocol server for Agent API.

This server exposes Agent API as MCP tools, resources, and prompts so MCP hosts
can create and inspect agent runs, work with durable volumes, search memory, and
use platform catalogs without speaking Agent API directly.

## Status

Early implementation scaffold. The initial transport is stdio for local MCP
hosts. Remote Streamable HTTP support is planned as the service surface settles.

## Install

```bash
npm install
npm run build
```

## Run

```bash
AGENT_API_KEY=sk-... \
AGENT_API_BASE_URL=https://api.agentsway.dev \
npm run dev
```

MCP host command:

```bash
node /path/to/agent-api-mcp/dist/index.js
```

## Initial Capability Map

- Agent runtime: create, retrieve, cancel, list events, list children.
- Catalog: list models, presets, and callable Agent API tools.
- Volumes: list/create volumes, list/search entries, read/write files, grep,
  summarize, and patch text lines.
- Memory: search long-term memory with thread or workspace scope.
- Skills: list, retrieve, discover, focus, read/write files, diff, and promote
  dev branches.

Credentials are read from environment variables and are never exposed through MCP
tool results.
