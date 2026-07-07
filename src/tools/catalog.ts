import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { AgentMcpContext } from "../context.js";
import { jsonToolResult } from "../mcp-result.js";

export function registerCatalogTools(server: McpServer, context: AgentMcpContext): void {
  server.registerTool(
    "agent_api_list_models",
    {
      title: "List Agent API Models",
      description: "List currently available Agent API model ids and metadata.",
    },
    async () => jsonToolResult(await context.client.models.list()),
  );

  server.registerTool(
    "agent_api_list_presets",
    {
      title: "List Agent API Presets",
      description: "List managed Agent API presets and their public policy metadata.",
    },
    async () => jsonToolResult(await context.client.presets.list()),
  );

  server.registerTool(
    "agent_api_list_tools",
    {
      title: "List Agent API Tools",
      description: "List callable platform tools known to Agent API.",
    },
    async () => jsonToolResult(await context.client.tools.list()),
  );
}
