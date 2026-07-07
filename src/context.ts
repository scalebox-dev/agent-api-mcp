import { AgentAPI } from "@agent-api/sdk";
import type { AgentMcpConfig } from "./config.js";
import { requireApiKey } from "./config.js";

export interface AgentMcpContext {
  config: AgentMcpConfig;
  client: AgentAPI;
}

export function createContext(config: AgentMcpConfig): AgentMcpContext {
  return {
    config,
    client: new AgentAPI({
      apiKey: requireApiKey(config),
      baseURL: config.baseURL,
      timeout: config.timeoutMs,
      streamTimeout: config.streamTimeoutMs,
      maxRetries: config.maxRetries,
    }),
  };
}
