import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { ResponseCreateParamsNonStreaming } from "@agent-api/sdk";
import * as z from "zod/v4";
import type { AgentMcpContext } from "../context.js";
import { jsonToolResult } from "../mcp-result.js";
import { paginationSchema, responseIdSchema, optionalString } from "../schema.js";

const createResponseSchema = {
  input: z.union([z.string(), z.array(z.unknown()), z.record(z.string(), z.unknown())]).describe("Agent API input payload."),
  instructions: optionalString,
  preset: optionalString.describe("Managed preset name such as pro-search or deep-research."),
  model: optionalString.describe("Single model id in vendor/model form."),
  models: z.array(z.string().trim().min(1)).min(1).max(5).optional(),
  model_routing: z.enum(["chain", "auto", "openmark"]).optional(),
  routing_strategy: z.enum(["quality", "cost-effective", "low-latency", "balanced"]).optional(),
  previous_response_id: optionalString,
  volume_id: optionalString,
  max_output_tokens: z.number().int().positive().optional(),
  max_steps: z.number().int().positive().max(10).optional(),
  plan_mode_preference: z.enum(["off", "auto", "preferred", "required"]).optional(),
  sub_agent_preference: z.enum(["off", "auto", "preferred", "required"]).optional(),
  memory: z
    .object({
      enabled: z.boolean().optional(),
      read: z.boolean().optional(),
      write: z.boolean().optional(),
      tenant_search: z.boolean().optional(),
    })
    .optional(),
  metadata: z.record(z.string(), z.union([z.string(), z.number(), z.boolean()])).optional(),
  preferred_sites: z.array(z.string().trim().min(1)).max(3).optional(),
};

export function registerResponseTools(server: McpServer, context: AgentMcpContext): void {
  server.registerTool(
    "agent_api_create_response",
    {
      title: "Create Agent API Response",
      description:
        "Create a non-streaming Agent API response. Supports presets, model routing, memory, volumes, planning, and sub-agent preferences.",
      inputSchema: createResponseSchema,
    },
    async (args) =>
      jsonToolResult(
        await context.client.responses.create({
          ...(args as unknown as Omit<ResponseCreateParamsNonStreaming, "stream">),
          stream: false,
        }),
      ),
  );

  server.registerTool(
    "agent_api_list_responses",
    {
      title: "List Agent API Responses",
      description: "List stored responses visible to the configured Agent API credential.",
      inputSchema: {
        ...paginationSchema,
        status: optionalString,
        safety_identifier: optionalString,
      },
    },
    async (args) => jsonToolResult(await context.client.responses.list(args)),
  );

  server.registerTool(
    "agent_api_get_response",
    {
      title: "Get Agent API Response",
      description: "Retrieve a persisted Agent API response by id.",
      inputSchema: {
        ...responseIdSchema,
        safety_identifier: optionalString,
      },
    },
    async ({ response_id, safety_identifier }) =>
      jsonToolResult(await context.client.responses.retrieve(response_id, { safety_identifier })),
  );

  server.registerTool(
    "agent_api_cancel_response",
    {
      title: "Cancel Agent API Response",
      description: "Best-effort cancellation for an in-flight Agent API response.",
      inputSchema: responseIdSchema,
    },
    async ({ response_id }) => jsonToolResult(await context.client.responses.cancel(response_id)),
  );

  server.registerTool(
    "agent_api_list_response_events",
    {
      title: "List Agent API Response Events",
      description: "List response audit/timeline events for debugging and replay.",
      inputSchema: {
        ...responseIdSchema,
        ...paginationSchema,
        after_sequence: z.number().int().nonnegative().optional(),
        view: z.enum(["timeline", "full"]).optional(),
      },
    },
    async ({ response_id, ...params }) => jsonToolResult(await context.client.responses.listEvents(response_id, params)),
  );

  server.registerTool(
    "agent_api_list_child_responses",
    {
      title: "List Child Responses",
      description: "List delegated sub-agent runs for a parent response.",
      inputSchema: responseIdSchema,
    },
    async ({ response_id }) => jsonToolResult(await context.client.responses.listChildren(response_id)),
  );

  server.registerTool(
    "agent_api_get_response_volume",
    {
      title: "Get Response Volume",
      description: "Resolve the durable agent volume associated with a response.",
      inputSchema: responseIdSchema,
    },
    async ({ response_id }) => jsonToolResult(await context.client.responses.retrieveVolume(response_id)),
  );
}
