import { describe, expect, it } from "vitest";
import { createAgentApiMcpServer } from "../src/server.js";

describe("server", () => {
  it("constructs the MCP server with an Agent API key", () => {
    const server = createAgentApiMcpServer({
      apiKey: "sk_test",
      baseURL: "http://localhost:18000",
      timeoutMs: 1000,
      streamTimeoutMs: 2000,
      maxRetries: 0,
    });

    expect(server.isConnected()).toBe(false);
  });
});
