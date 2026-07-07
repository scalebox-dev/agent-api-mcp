import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import * as z from "zod/v4";

export function registerPrompts(server: McpServer): void {
  server.registerPrompt(
    "research_with_agent",
    {
      title: "Research With Agent API",
      description: "Plan and run a research task through Agent API.",
      argsSchema: {
        topic: z.string().trim().min(1),
        depth: z.enum(["fast", "standard", "deep"]).optional(),
      },
    },
    async ({ topic, depth }) => ({
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text:
              `Use Agent API to research this topic: ${topic}\n\n` +
              `Choose an appropriate preset for depth=${depth ?? "standard"}, create a response, then inspect events and artifacts if needed.`,
          },
        },
      ],
    }),
  );

  server.registerPrompt(
    "continue_agent_thread",
    {
      title: "Continue Agent Thread",
      description: "Continue an existing Agent API response thread.",
      argsSchema: {
        previous_response_id: z.string().trim().min(1),
        instruction: z.string().trim().min(1),
      },
    },
    async ({ previous_response_id, instruction }) => ({
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text:
              `Continue Agent API thread ${previous_response_id} with this instruction:\n\n` +
              `${instruction}\n\nUse previous_response_id when creating the next response.`,
          },
        },
      ],
    }),
  );

  server.registerPrompt(
    "debug_agent_response",
    {
      title: "Debug Agent Response",
      description: "Inspect a response, timeline events, child runs, and associated volume.",
      argsSchema: {
        response_id: z.string().trim().min(1),
      },
    },
    async ({ response_id }) => ({
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text:
              `Debug Agent API response ${response_id}. Retrieve the response, list response events, list child responses, ` +
              "and resolve the response volume if present. Summarize status, errors, usage, tool activity, and likely next actions.",
          },
        },
      ],
    }),
  );
}
