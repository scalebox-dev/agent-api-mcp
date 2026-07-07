import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { AgentMcpConfig } from "./config.js";
import { createContext } from "./context.js";
import { registerPrompts } from "./prompts.js";
import { registerResources } from "./resources.js";
import { registerCatalogTools } from "./tools/catalog.js";
import { registerMemoryTools } from "./tools/memories.js";
import { registerResponseTools } from "./tools/responses.js";
import { registerSkillTools } from "./tools/skills.js";
import { registerVolumeTools } from "./tools/volumes.js";

export function createAgentApiMcpServer(config: AgentMcpConfig): McpServer {
  const context = createContext(config);
  const server = new McpServer(
    {
      name: "agent-api-mcp",
      version: "0.1.0",
    },
    {
      capabilities: {
        logging: {},
      },
    },
  );

  registerCatalogTools(server, context);
  registerResponseTools(server, context);
  registerVolumeTools(server, context);
  registerMemoryTools(server, context);
  registerSkillTools(server, context);
  registerResources(server, context);
  registerPrompts(server);

  return server;
}
