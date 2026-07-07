import * as z from "zod/v4";

export const optionalString = z.string().trim().min(1).optional();

export const paginationSchema = {
  limit: z.number().int().positive().max(100).optional(),
  page_token: optionalString,
};

export const responseIdSchema = {
  response_id: z.string().trim().min(1).describe("Agent API response id."),
};

export const volumeIdSchema = {
  volume_id: z.string().trim().min(1).describe("Agent API durable volume id."),
};

export const skillIdSchema = {
  skill_id: z.string().trim().min(1).describe("Agent API skill id."),
};

export const branchSchema = z.enum(["main", "dev"]).optional();
