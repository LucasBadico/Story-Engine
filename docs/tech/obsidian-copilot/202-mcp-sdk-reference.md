# 202 — MCP TypeScript SDK Reference

## Package

```
npm install @modelcontextprotocol/sdk zod
```

- **Version:** 1.28.0 (2026-03-25)
- **Peer dep:** zod ^3.25 or ^4.0
- **Node:** >= 18
- **GitHub:** github.com/modelcontextprotocol/typescript-sdk

## Client Transports

| Class | Import | Use Case |
|---|---|---|
| `SSEClientTransport` | `@modelcontextprotocol/sdk/client/sse` | Connect to SSE servers (our case) |
| `StreamableHTTPClientTransport` | `@modelcontextprotocol/sdk/client/streamableHttp` | Newer HTTP transport (preferred for new servers) |
| `StdioClientTransport` | `@modelcontextprotocol/sdk/client/stdio` | Spawn local process, communicate via stdin/stdout |
| `WebSocketClientTransport` | `@modelcontextprotocol/sdk/client/websocket` | WebSocket transport |
| `InMemoryTransport` | `@modelcontextprotocol/sdk` | Testing (linked pairs) |

**Note:** `SSEClientTransport` is deprecated in favor of `StreamableHTTPClientTransport`, but many servers (including ours) still use SSE. Support both during migration.

## Connect + List + Call

```typescript
import { Client } from "@modelcontextprotocol/sdk/client";
import { SSEClientTransport } from "@modelcontextprotocol/sdk/client/sse";

// 1. Connect
const client = new Client({ name: "obsidian-copilot", version: "3.2.5-mcp.1" });
const transport = new SSEClientTransport(new URL("http://localhost:8082/sse"));
await client.connect(transport);

// 2. Discover tools
const { tools } = await client.listTools();
// tools: Array<{ name: string, description: string, inputSchema: JSONSchema }>

// 3. Call a tool
const result = await client.callTool({
  name: "get_canon",
  arguments: { story_id: "abc-123", as_of_chapter: 5 }
});
// result.content: Array<{ type: "text", text: string } | { type: "image", ... }>
// result.isError?: boolean

// 4. Disconnect
await client.close();
```

## Client API

```typescript
class Client {
  constructor(info: { name: string, version: string }, options?: ClientOptions)

  // Lifecycle
  connect(transport: Transport): Promise<void>
  close(): Promise<void>
  ping(): Promise<void>

  // Tools
  listTools(params?): Promise<{ tools: Tool[] }>
  callTool(params: { name: string, arguments: Record<string, unknown> }): Promise<CallToolResult>

  // Resources
  listResources(params?): Promise<{ resources: Resource[] }>
  readResource(params: { uri: string }): Promise<ReadResourceResult>
  listResourceTemplates(params?): Promise<{ resourceTemplates: ResourceTemplate[] }>
  subscribeResource(params): Promise<void>
  unsubscribeResource(params): Promise<void>

  // Prompts
  listPrompts(params?): Promise<{ prompts: Prompt[] }>
  getPrompt(params): Promise<GetPromptResult>

  // Server info (available after connect)
  getServerCapabilities(): ServerCapabilities | undefined
  getServerVersion(): Implementation | undefined
  getInstructions(): string | undefined

  // Logging
  setLoggingLevel(level): Promise<void>
}
```

## Tool Definition Shape (from listTools)

```typescript
interface Tool {
  name: string;           // "get_canon"
  description: string;    // "Retrieve canonical story data..."
  inputSchema: {          // JSON Schema for parameters
    type: "object";
    properties: {
      [key: string]: {
        type: string;
        description?: string;
        enum?: string[];
        // ... standard JSON Schema
      }
    };
    required?: string[];
  };
}
```

## CallToolResult Shape

```typescript
interface CallToolResult {
  content: Array<
    | { type: "text"; text: string }
    | { type: "image"; data: string; mimeType: string }
    | { type: "resource"; resource: { uri: string; text?: string; blob?: string } }
  >;
  isError?: boolean;
  structuredContent?: Record<string, unknown>;
}
```

## Transport Interface

All transports implement:

```typescript
interface Transport {
  start(): Promise<void>
  send(message: JSONRPCMessage): Promise<void>
  close(): Promise<void>
  onclose?: () => void
  onerror?: (error: Error) => void
  onmessage?: (message: JSONRPCMessage) => void
  sessionId?: string
}
```

## SSE Transport Details

```typescript
const transport = new SSEClientTransport(
  new URL("http://localhost:8082/sse"),
  {
    // Optional: custom fetch for auth headers
    eventSourceInit?: { fetch: customFetch },
    // Optional: request init for POST messages
    requestInit?: { headers: { "Authorization": "Bearer ..." } },
  }
);
```

The SSE transport:
1. Opens EventSource (GET) to receive server messages
2. Sends client messages via POST to the endpoint returned in the SSE stream
3. Handles reconnection via EventSource built-in retry

## Story-Block MCP Server Tools

Our MCP server at `:8082` exposes:

| Tool | Parameters | Returns |
|---|---|---|
| `get_canon` | story_id, as_of_chapter?, mode? | Full canon: documents, states, events |
| `search_story_memory` | story_id, query, as_of_chapter?, mode? | Semantic search results |
| `get_character_state` | story_id, character_name, as_of_chapter? | Character state snapshot |
| `get_timeline` | story_id, as_of_chapter?, limit? | Timeline events |
| `get_plots` | story_id, as_of_chapter? | Plot threads + developments |
| `get_writing_context` | story_id, chapter, scene | Assembled context pack |

Plus 2 resources:
- `story://stories/{id}/identity`
- `story://stories/{id}/world-rules`

## Dependencies Breakdown

| Dep | Why it's there |
|---|---|
| `zod` (peer) | Schema validation |
| `eventsource` + `eventsource-parser` | SSE client |
| `cross-spawn` | stdio transport |
| `ajv` + `ajv-formats` | JSON Schema validation for tool output |
| `content-type`, `raw-body` | HTTP parsing |
| `jose` | OAuth JWT (not needed for our use) |
| `express`, `hono` | Server transports (not needed for client) |

For client-only usage, the heavy deps (express, hono) are tree-shakeable — esbuild only bundles what we import.
