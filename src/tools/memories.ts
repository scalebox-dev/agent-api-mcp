import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import * as z from "zod/v4";
import type { AgentMcpContext } from "../context.js";
import { jsonToolResult } from "../mcp-result.js";
import { optionalString } from "../schema.js";

export function registerMemoryTools(server: McpServer, context: AgentMcpContext): void {
  server.registerTool(
    "agent_api_search_memories",
    {
      title: "Search Agent Memories",
      description: "Search Agent API long-term memory with thread or explicit workspace scope.",
      inputSchema: {
        query: z.string().trim().min(1),
        limit: z.number().int().nonnegative().max(100).optional(),
        previous_response_id: optionalString,
        tenant_search: z.boolean().optional(),
        lang: optionalString,
        semantic_weight: z.number().min(0).max(1).optional(),
      },
    },
    async (args) => jsonToolResult(await context.client.memories.search(args)),
  );
}
