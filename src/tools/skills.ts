import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import * as z from "zod/v4";
import type { AgentMcpContext } from "../context.js";
import { jsonToolResult } from "../mcp-result.js";
import { branchSchema, optionalString, paginationSchema, skillIdSchema } from "../schema.js";

export function registerSkillTools(server: McpServer, context: AgentMcpContext): void {
  server.registerTool(
    "agent_api_list_skills",
    {
      title: "List Skills",
      description: "List workspace skills.",
      inputSchema: {
        ...paginationSchema,
        archived: z.boolean().optional(),
        user_id: optionalString,
      },
    },
    async (args) => jsonToolResult(await context.client.skills.list(args)),
  );

  server.registerTool(
    "agent_api_get_skill",
    {
      title: "Get Skill",
      description: "Retrieve skill metadata.",
      inputSchema: skillIdSchema,
    },
    async ({ skill_id }) => jsonToolResult(await context.client.skills.retrieve(skill_id)),
  );

  server.registerTool(
    "agent_api_discover_skills",
    {
      title: "Discover Skills",
      description: "Discover relevant skills for a task or query.",
      inputSchema: {
        query: optionalString,
        skills: z
          .array(
            z.object({
              skill_id: z.string().trim().min(1),
              branch: branchSchema,
            }),
          )
          .optional(),
        max_results: z.number().int().positive().max(50).optional(),
      },
    },
    async (args) => jsonToolResult(await context.client.skills.discover(args)),
  );

  server.registerTool(
    "agent_api_focus_skills",
    {
      title: "Focus Skills",
      description: "Load selected skill manifests and files for model context.",
      inputSchema: {
        skills: z.array(
          z.object({
            skill_id: z.string().trim().min(1),
            branch: branchSchema,
            paths: z.array(z.string().trim().min(1)).optional(),
            include_manifest: z.boolean().optional(),
          }),
        ),
        fallback_to_main: z.boolean().optional(),
        max_manifest_chars: z.number().int().positive().optional(),
        max_file_chars: z.number().int().positive().optional(),
      },
    },
    async (args) => jsonToolResult(await context.client.skills.focus(args)),
  );

  server.registerTool(
    "agent_api_list_skill_files",
    {
      title: "List Skill Files",
      description: "List files in a skill branch.",
      inputSchema: {
        ...skillIdSchema,
        branch: branchSchema,
        path: optionalString,
        fallback_to_main: z.boolean().optional(),
        ...paginationSchema,
      },
    },
    async ({ skill_id, ...params }) => jsonToolResult(await context.client.skills.listFiles(skill_id, params)),
  );

  server.registerTool(
    "agent_api_read_skill_file",
    {
      title: "Read Skill File",
      description: "Read a file from a skill branch.",
      inputSchema: {
        ...skillIdSchema,
        path: z.string().trim().min(1),
        branch: branchSchema,
        fallback_to_main: z.boolean().optional(),
        max_bytes: z.number().int().positive().optional(),
      },
    },
    async ({ skill_id, path, ...params }) => jsonToolResult(await context.client.skills.readFile(skill_id, path, params)),
  );

  server.registerTool(
    "agent_api_write_skill_file",
    {
      title: "Write Skill File",
      description: "Write text content to a skill branch file.",
      inputSchema: {
        ...skillIdSchema,
        path: z.string().trim().min(1),
        content: z.string(),
        branch: branchSchema,
      },
      annotations: {
        destructiveHint: true,
      },
    },
    async ({ skill_id, path, content, ...params }) =>
      jsonToolResult(await context.client.skills.writeFile(skill_id, path, content, params)),
  );

  server.registerTool(
    "agent_api_diff_skill",
    {
      title: "Diff Skill Branches",
      description: "Diff skill main and dev branches.",
      inputSchema: {
        ...skillIdSchema,
        path: optionalString,
        max_file_chars: z.number().int().positive().optional(),
        include_unchanged: z.boolean().optional(),
      },
    },
    async ({ skill_id, ...params }) => jsonToolResult(await context.client.skills.diff(skill_id, params)),
  );

  server.registerTool(
    "agent_api_accept_skill_dev",
    {
      title: "Accept Skill Dev Branch",
      description: "Promote a skill dev branch to main.",
      inputSchema: {
        ...skillIdSchema,
        strategy: z.enum(["patch", "mirror"]).optional(),
      },
      annotations: {
        destructiveHint: true,
      },
    },
    async ({ skill_id, strategy }) => jsonToolResult(await context.client.skills.acceptDev(skill_id, { strategy })),
  );
}
