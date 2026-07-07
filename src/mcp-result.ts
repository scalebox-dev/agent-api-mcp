export function jsonToolResult(value: unknown) {
  return {
    content: [
      {
        type: "text" as const,
        text: stringify(value),
      },
    ],
    structuredContent: normalizeStructured(value),
  };
}

export function textToolResult(text: string) {
  return {
    content: [
      {
        type: "text" as const,
        text,
      },
    ],
  };
}

export function jsonResource(uri: string, value: unknown) {
  return {
    contents: [
      {
        uri,
        mimeType: "application/json",
        text: stringify(value),
      },
    ],
  };
}

export function textResource(uri: string, text: string, mimeType = "text/plain") {
  return {
    contents: [
      {
        uri,
        mimeType,
        text,
      },
    ],
  };
}

function stringify(value: unknown): string {
  return JSON.stringify(value, null, 2);
}

function normalizeStructured(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return { result: value };
}
