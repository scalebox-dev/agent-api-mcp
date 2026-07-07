import { describe, expect, it } from "vitest";
import { loadConfig, requireApiKey } from "../src/config.js";

describe("config", () => {
  it("loads defaults", () => {
    const config = loadConfig({});

    expect(config.baseURL).toBe("https://api.agentsway.dev");
    expect(config.timeoutMs).toBe(600_000);
    expect(config.streamTimeoutMs).toBe(3_600_000);
    expect(config.maxRetries).toBe(2);
  });

  it("requires an api key for Agent API calls", () => {
    expect(() => requireApiKey(loadConfig({}))).toThrow("AGENT_API_KEY is required");
  });

  it("loads explicit environment values", () => {
    const config = loadConfig({
      AGENT_API_KEY: "sk_test",
      AGENT_API_BASE_URL: "http://localhost:18000",
      AGENT_API_TIMEOUT_MS: "1000",
      AGENT_API_STREAM_TIMEOUT_MS: "2000",
      AGENT_API_MAX_RETRIES: "4",
    });

    expect(config).toEqual({
      apiKey: "sk_test",
      baseURL: "http://localhost:18000",
      timeoutMs: 1000,
      streamTimeoutMs: 2000,
      maxRetries: 4,
    });
  });
});
