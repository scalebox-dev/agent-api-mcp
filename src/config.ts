export interface AgentMcpConfig {
  apiKey?: string;
  baseURL: string;
  timeoutMs: number;
  streamTimeoutMs: number;
  maxRetries: number;
}

const DEFAULT_BASE_URL = "https://api.agentsway.dev";
const DEFAULT_TIMEOUT_MS = 600_000;
const DEFAULT_STREAM_TIMEOUT_MS = 3_600_000;
const DEFAULT_MAX_RETRIES = 2;

export function loadConfig(env: NodeJS.ProcessEnv = process.env): AgentMcpConfig {
  return {
    apiKey: nonEmpty(env.AGENT_API_KEY),
    baseURL: nonEmpty(env.AGENT_API_BASE_URL) ?? DEFAULT_BASE_URL,
    timeoutMs: positiveInt(env.AGENT_API_TIMEOUT_MS, DEFAULT_TIMEOUT_MS),
    streamTimeoutMs: positiveInt(env.AGENT_API_STREAM_TIMEOUT_MS, DEFAULT_STREAM_TIMEOUT_MS),
    maxRetries: positiveInt(env.AGENT_API_MAX_RETRIES, DEFAULT_MAX_RETRIES),
  };
}

export function requireApiKey(config: AgentMcpConfig): string {
  if (!config.apiKey) {
    throw new Error("AGENT_API_KEY is required to call Agent API");
  }
  return config.apiKey;
}

function nonEmpty(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function positiveInt(value: string | undefined, fallback: number): number {
  if (!value) return fallback;
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}
