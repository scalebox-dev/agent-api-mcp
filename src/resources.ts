import { McpServer, ResourceTemplate } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { AgentMcpContext } from "./context.js";
import { jsonResource } from "./mcp-result.js";

export function registerResources(server: McpServer, context: AgentMcpContext): void {
  server.registerResource(
    "agent-api-models",
    "agentapi://models",
    {
      title: "Agent API Models",
      description: "Current Agent API model catalog.",
      mimeType: "application/json",
    },
    async (uri) => jsonResource(uri.href, await context.client.models.list()),
  );

  server.registerResource(
    "agent-api-presets",
    "agentapi://presets",
    {
      title: "Agent API Presets",
      description: "Current Agent API preset catalog.",
      mimeType: "application/json",
    },
    async (uri) => jsonResource(uri.href, await context.client.presets.list()),
  );

  server.registerResource(
    "agent-api-tools",
    "agentapi://tools",
    {
      title: "Agent API Tools",
      description: "Current Agent API callable tool catalog.",
      mimeType: "application/json",
    },
    async (uri) => jsonResource(uri.href, await context.client.tools.list()),
  );

  server.registerResource(
    "agent-api-response",
    new ResourceTemplate("agentapi://responses/{response_id}", { list: undefined }),
    {
      title: "Agent API Response",
      description: "A persisted Agent API response.",
      mimeType: "application/json",
    },
    async (uri, vars) => jsonResource(uri.href, await context.client.responses.retrieve(requiredVar(vars, "response_id"))),
  );

  server.registerResource(
    "agent-api-response-events",
    new ResourceTemplate("agentapi://responses/{response_id}/events", { list: undefined }),
    {
      title: "Agent API Response Events",
      description: "Audit/timeline events for an Agent API response.",
      mimeType: "application/json",
    },
    async (uri, vars) => jsonResource(uri.href, await context.client.responses.listEvents(requiredVar(vars, "response_id"))),
  );

  server.registerResource(
    "agent-api-response-children",
    new ResourceTemplate("agentapi://responses/{response_id}/children", { list: undefined }),
    {
      title: "Agent API Child Responses",
      description: "Delegated sub-agent runs for an Agent API response.",
      mimeType: "application/json",
    },
    async (uri, vars) => jsonResource(uri.href, await context.client.responses.listChildren(requiredVar(vars, "response_id"))),
  );

  server.registerResource(
    "agent-api-volume-file",
    new ResourceTemplate("agentapi://volumes/{volume_id}/files/{path}", { list: undefined }),
    {
      title: "Agent API Volume File",
      description: "A file delivered from an Agent API durable volume.",
      mimeType: "application/json",
    },
    async (uri, vars) => {
      const volumeID = requiredVar(vars, "volume_id");
      const path = requiredVar(vars, "path");
      return jsonResource(uri.href, await context.client.volumes.readFile(volumeID, path));
    },
  );

  server.registerResource(
    "agent-api-skill-file",
    new ResourceTemplate("agentapi://skills/{skill_id}/files/{path}", { list: undefined }),
    {
      title: "Agent API Skill File",
      description: "A file from an Agent API skill.",
      mimeType: "application/json",
    },
    async (uri, vars) => {
      const skillID = requiredVar(vars, "skill_id");
      const path = requiredVar(vars, "path");
      return jsonResource(uri.href, await context.client.skills.readFile(skillID, path));
    },
  );
}

function requiredVar(vars: Record<string, string | string[]>, name: string): string {
  const value = vars[name];
  const first = Array.isArray(value) ? value[0] : value;
  if (!first) {
    throw new Error(`Missing resource variable: ${name}`);
  }
  return first;
}
