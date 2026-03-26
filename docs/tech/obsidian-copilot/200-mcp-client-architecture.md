# 200 — MCP Client Architecture for Obsidian Copilot

## Status: Phase 1 Complete (v3.2.5-mcp.13)

End-to-end working: Obsidian Copilot Agent → MCP SSE → Story Engine → PostgreSQL → Canon/States/Events returned to chat.

## Overview

Adding MCP (Model Context Protocol) client support to Obsidian Copilot so the AI agent can call tools exposed by external MCP servers. This is a generic feature — any MCP server works, not just story-block's.

**Upstream issue:** [logancyang/obsidian-copilot#1451](https://github.com/logancyang/obsidian-copilot/issues/1451)

## Quick Start

```bash
# Start story engine services
make start        # api, grpc, gateway, worker
make mcp          # MCP server on :8082

# Build and deploy Copilot fork
make copilot                    # default vault (Obsidian/writing)
make copilot VAULT=other-vault  # specific vault under Documents/notes/

# In Obsidian:
# 1. Settings → Copilot → MCP tab → Add server (http://localhost:8082/sse)
# 2. Chat dropdown → select "agent"
# 3. Toggle brain icon ON
# 4. Ask about your story
```

## System Diagram

```
MCP Server (e.g. story-block at :8082)
    |
    | SSE / JSON-RPC 2.0
    v
McpClient (src/mcp/McpClient.ts)
    |  - SSEClientTransport from @modelcontextprotocol/sdk
    |  - connect(), listTools(), callTool(), disconnect()
    v
McpToolBridge (src/mcp/McpToolBridge.ts)
    |  - JSON Schema -> Zod via @n8n/json-schema-to-zod
    |  - Creates LangChain StructuredTool per MCP tool
    |  - Registers in ToolRegistry with category: "mcp"
    v
ToolRegistry (existing, src/tools/ToolRegistry.ts)
    |  - Already has "mcp" category in ToolMetadata
    |  - getEnabledTools() returns MCP tools alongside builtins
    v
AutonomousAgentChainRunner (existing)
    |  - bindTools() includes MCP tools
    |  - ReAct loop calls MCP tools like any other tool
    v
Chat UI (existing)
```

## Dependencies

| Package | Purpose | Size |
|---|---|---|
| `@modelcontextprotocol/sdk` | MCP client transport + protocol | Official SDK, v1.28.0 |
| `@n8n/json-schema-to-zod` | Runtime JSON Schema -> Zod objects | Battle-tested at n8n |

Both are runtime dependencies. No codegen or build-time tooling needed.

### Why these libraries

**MCP SDK** — The official TypeScript SDK. Provides `SSEClientTransport`, `Client` class with `listTools()` / `callTool()`. Used by Claude Desktop, Cursor, and all major MCP clients. No alternatives exist.

**@n8n/json-schema-to-zod** — MCP tools define parameters as JSON Schema. LangChain tools need Zod schemas. This library converts at runtime, returning real Zod objects (not code strings). Chosen over:
- `json-schema-to-zod` (StefanTerdell) — codegen only, explicitly "not meant for runtime"
- `z.fromJSONSchema()` — requires Zod v4, LangChain still on v3
- Hand-written converter — viable (~100 lines) but n8n's handles edge cases we'd miss

## New Files

```
obsidian-copilot/src/mcp/
  McpClient.ts          # Transport layer: SSE connection, tool discovery, tool execution
  McpToolBridge.ts      # Schema conversion + LangChain tool creation + ToolRegistry integration
  McpServerManager.ts   # Manages multiple server connections, lifecycle, reconnection
  __tests__/
    McpClient.test.ts
    McpToolBridge.test.ts
    McpServerManager.test.ts
```

## Existing Code Integration Points

### ToolRegistry (src/tools/ToolRegistry.ts)

Already supports `category: "mcp"` in the `ToolMetadata` type. No changes needed to the registry itself — we just call `registry.register()` with MCP tool definitions.

```typescript
// ToolMetadata.category already includes "mcp"
category: "search" | "time" | "file" | "media" | "mcp" | "memory" | "custom" | "cli"
```

### createLangChainTool (src/tools/createLangChainTool.ts)

Used to wrap each MCP tool as a LangChain `StructuredTool`:

```typescript
const langchainTool = createLangChainTool({
  name: mcpTool.name,
  description: mcpTool.description,
  schema: jsonSchemaToZod(mcpTool.inputSchema),  // runtime conversion
  func: async (args) => {
    const result = await mcpClient.callTool(mcpTool.name, args);
    return result;
  },
});
```

### builtinTools.ts — initializeBuiltinTools()

The initialization pattern to follow. MCP tools register alongside builtins:

```typescript
// Existing pattern (builtinTools.ts:487-524):
export function initializeBuiltinTools(vault?: Vault): void {
  const registry = ToolRegistry.getInstance();
  registry.clear();
  registry.registerAll(BUILTIN_TOOLS);
  if (vault) registerFileTreeTool(vault);
  if (settings.enableSavedMemory) registerMemoryTool();
  if (isDesktopRuntime()) registerCliTools();
}

// New: MCP tools registered via McpServerManager after builtin init
```

### AutonomousAgentChainRunner (lines 128-141)

Gets tools via `registry.getEnabledTools()`. MCP tools appear automatically once registered — no changes needed to the agent runner.

### main.ts (plugin lifecycle)

MCP servers connect on `onload()`, disconnect on `onunload()`:

```typescript
async onload() {
  initializeBuiltinTools(this.app.vault);
  await this.mcpServerManager.connectAll();  // NEW
}

onunload() {
  this.mcpServerManager.disconnectAll();  // NEW
}
```

## Component Details

### McpClient

Thin wrapper around `@modelcontextprotocol/sdk` Client + SSEClientTransport.

```typescript
class McpClient {
  private client: Client;
  private transport: SSEClientTransport;

  async connect(url: string): Promise<void>
  async listTools(): Promise<McpToolDefinition[]>
  async callTool(name: string, args: Record<string, unknown>): Promise<string>
  async disconnect(): Promise<void>
  isConnected(): boolean
}
```

**Connection lifecycle:**
1. Create `SSEClientTransport` with server URL
2. Create MCP `Client` with name/version
3. `client.connect(transport)` — starts SSE stream + MCP handshake
4. `client.listTools()` — discovers available tools
5. `client.callTool({ name, arguments })` — invokes a tool
6. `transport.close()` — disconnect

### McpToolBridge

Converts MCP tool definitions into LangChain tools and registers them.

```typescript
class McpToolBridge {
  constructor(private mcpClient: McpClient, private serverId: string)

  async registerTools(): Promise<void>
  // 1. mcpClient.listTools() -> MCP tool definitions
  // 2. For each tool: JSON Schema -> Zod -> createLangChainTool()
  // 3. registry.register({ tool, metadata: { category: "mcp" } })

  unregisterTools(): void
  // Remove all tools from this server from the registry
}
```

**JSON Schema -> Zod conversion:**

```
MCP inputSchema (JSON Schema)          Zod Schema
─────────────────────────────          ──────────
{ type: "string" }                 ->  z.string()
{ type: "number" }                 ->  z.number()
{ type: "integer" }                ->  z.number().int()
{ type: "boolean" }                ->  z.boolean()
{ type: "array", items: {...} }    ->  z.array(...)
{ type: "object", properties: {} } ->  z.object({...})
{ enum: ["a", "b"] }               ->  z.enum(["a", "b"])
+ description -> .describe()
+ required -> present vs .optional()
```

### McpServerManager

Manages server connections, settings integration, and lifecycle.

```typescript
class McpServerManager {
  private servers: Map<string, { client: McpClient, bridge: McpToolBridge }>

  async connectAll(): Promise<void>      // Connect all configured servers
  async connectServer(config): Promise<void>
  async disconnectServer(id): Promise<void>
  async disconnectAll(): Promise<void>
  async reconnectServer(id): Promise<void>
  getServerStatus(id): "connected" | "disconnected" | "error"
}
```

**Server config shape (stored in Copilot settings):**

```typescript
interface McpServerConfig {
  id: string;           // unique identifier
  name: string;         // display name (e.g. "Story Engine")
  transport: "sse";     // "stdio" in Phase 2
  url: string;          // SSE endpoint (e.g. "http://localhost:8082/sse")
  enabled: boolean;     // toggle on/off
}
```

## Settings Integration

New settings field in Copilot's settings schema:

```typescript
mcpServers: McpServerConfig[]
// Default: []
```

**Phase 1 UI** — minimal: text input for a single server URL in Copilot settings.
**Phase 2 UI** — full: add/edit/remove servers, connection status indicators.

## Tool Naming

MCP tools are namespaced to avoid collisions with builtin tools:

```
MCP tool name:    get_canon
Registered as:    mcp_storyengine_get_canon
Display name:     [Story Engine] Get Canon
```

Pattern: `mcp_{serverId}_{toolName}`

## Error Handling

| Scenario | Behavior |
|---|---|
| Server unreachable on startup | Log warning, skip — don't block plugin load |
| Server disconnects mid-session | Remove MCP tools from registry, log warning |
| Server reconnects | Re-discover tools, re-register |
| Tool call fails | Return error string to agent (ReAct loop handles retries) |
| Invalid JSON Schema from server | Skip that tool, log warning, register others |
| Timeout on tool call | 30s timeout (matches existing tool timeout pattern) |

## Phases

### Phase 1 — SSE Client MVP
- McpClient with SSE transport
- McpToolBridge with JSON Schema -> Zod
- McpServerManager (single server)
- Settings: single URL input
- Plugin lifecycle: connect/disconnect
- Tests

### Phase 2 — Multi-Server + stdio
- Multiple MCP server support
- stdio transport (spawn local process)
- Settings UI: server list with add/edit/remove
- Connection status indicators
- Auto-reconnect with backoff

### Phase 3 — Resources
- MCP resource discovery + reading
- Surface as context sources in Copilot
- Resource templates support

## Example: Story Engine MCP Integration

With this feature, Obsidian Copilot + story-block MCP server:

```
User: "What is Haruka's emotional state in the prologue?"

Agent reasoning:
1. Tool available: mcp_storyengine_get_character_state
2. Call: get_character_state({ story_id: "...", character_name: "haruka", as_of_chapter: 0 })
3. MCP server queries gateway -> returns character state snapshot
4. Agent synthesizes response with story context

Response: "In the prologue (Chapter 0, Day 0), Haruka is..."
```

The agent treats MCP tools exactly like builtin tools — the ReAct loop handles tool selection, execution, and response synthesis automatically.
