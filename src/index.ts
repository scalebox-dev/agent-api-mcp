#!/usr/bin/env node
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { loadConfig } from "./config.js";
import { createAgentApiMcpServer } from "./server.js";

async function main(): Promise<void> {
  const server = createAgentApiMcpServer(loadConfig());
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("agent-api-mcp running on stdio");
}

main().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`agent-api-mcp failed: ${message}`);
  process.exit(1);
});
