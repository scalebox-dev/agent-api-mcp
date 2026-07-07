import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import * as z from "zod/v4";
import type { AgentMcpContext } from "../context.js";
import { jsonToolResult } from "../mcp-result.js";
import { optionalString, paginationSchema, volumeIdSchema } from "../schema.js";

export function registerVolumeTools(server: McpServer, context: AgentMcpContext): void {
  server.registerTool(
    "agent_api_list_volumes",
    {
      title: "List Volumes",
      description: "List durable Agent API volumes.",
      inputSchema: {
        ...paginationSchema,
        user_id: optionalString,
      },
    },
    async (args) => jsonToolResult(await context.client.volumes.list(args)),
  );

  server.registerTool(
    "agent_api_create_volume",
    {
      title: "Create Volume",
      description: "Create a durable Agent API volume.",
      inputSchema: {
        name: optionalString,
      },
    },
    async (args) => jsonToolResult(await context.client.volumes.create(args)),
  );

  server.registerTool(
    "agent_api_list_volume_entries",
    {
      title: "List Volume Entries",
      description: "List files and directories inside a volume.",
      inputSchema: {
        ...volumeIdSchema,
        path: optionalString,
        recursive: z.boolean().optional(),
        ...paginationSchema,
      },
    },
    async ({ volume_id, ...params }) => jsonToolResult(await context.client.volumes.listEntries(volume_id, params)),
  );

  server.registerTool(
    "agent_api_search_volume_entries",
    {
      title: "Search Volume Paths",
      description: "Search file and directory paths inside a volume.",
      inputSchema: {
        ...volumeIdSchema,
        query: z.string().trim().min(1),
        path: optionalString,
        ...paginationSchema,
      },
    },
    async ({ volume_id, ...params }) => jsonToolResult(await context.client.volumes.searchEntries(volume_id, params)),
  );

  server.registerTool(
    "agent_api_read_volume_file",
    {
      title: "Read Volume File",
      description: "Read a volume file as delivered text, extracted text, image URL, or base64 content.",
      inputSchema: {
        ...volumeIdSchema,
        path: z.string().trim().min(1),
        format: z.enum(["auto", "text", "base64", "url"]).optional(),
        max_bytes: z.number().int().positive().optional(),
      },
    },
    async ({ volume_id, path, ...params }) => jsonToolResult(await context.client.volumes.readFile(volume_id, path, params)),
  );

  server.registerTool(
    "agent_api_write_volume_file",
    {
      title: "Write Volume File",
      description: "Write text content to a volume file.",
      inputSchema: {
        ...volumeIdSchema,
        path: z.string().trim().min(1),
        content: z.string(),
      },
      annotations: {
        destructiveHint: true,
      },
    },
    async ({ volume_id, path, content }) => jsonToolResult(await context.client.volumes.writeFile(volume_id, path, content)),
  );

  server.registerTool(
    "agent_api_read_volume_lines",
    {
      title: "Read Volume Lines",
      description: "Read a line range from a text file in a volume.",
      inputSchema: {
        ...volumeIdSchema,
        path: z.string().trim().min(1),
        start_line: z.number().int().positive(),
        end_line: z.number().int().nonnegative().optional(),
      },
    },
    async ({ volume_id, path, ...params }) => jsonToolResult(await context.client.volumes.readLines(volume_id, path, params)),
  );

  server.registerTool(
    "agent_api_patch_volume_lines",
    {
      title: "Patch Volume Lines",
      description: "Replace a line range in a volume text file.",
      inputSchema: {
        ...volumeIdSchema,
        path: z.string().trim().min(1),
        start_line: z.number().int().positive(),
        end_line: z.number().int().nonnegative().optional(),
        replacement: z.string().optional(),
      },
      annotations: {
        destructiveHint: true,
      },
    },
    async ({ volume_id, path, ...params }) => jsonToolResult(await context.client.volumes.patchLines(volume_id, path, params)),
  );

  server.registerTool(
    "agent_api_grep_volume",
    {
      title: "Grep Volume",
      description: "Search text content inside volume files.",
      inputSchema: {
        ...volumeIdSchema,
        pattern: z.string().min(1),
        path: optionalString,
        case_sensitive: z.boolean().optional(),
        max_matches: z.number().int().positive().max(1000).optional(),
        ...paginationSchema,
      },
    },
    async ({ volume_id, ...params }) => jsonToolResult(await context.client.volumes.grep(volume_id, params)),
  );

  server.registerTool(
    "agent_api_summarize_volume",
    {
      title: "Summarize Volume",
      description: "Summarize volume contents and text previews.",
      inputSchema: {
        ...volumeIdSchema,
        path: optionalString,
      },
    },
    async ({ volume_id, ...params }) => jsonToolResult(await context.client.volumes.summarize(volume_id, params)),
  );
}
