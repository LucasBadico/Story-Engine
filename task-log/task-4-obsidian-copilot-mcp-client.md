# Task 4 — Obsidian Copilot MCP Client Support

## Goal
Add MCP (Model Context Protocol) client support to Obsidian Copilot so it can connect to external MCP servers and use their tools within the existing agent/tool system. This is a generic feature — not story-block specific. The PR goes back to the upstream repo (logancyang/obsidian-copilot, issue #1451).

## Context
- Obsidian Copilot is AGPL-3.0 — any modifications must be open-sourced under AGPL
- The feature is generic MCP client infrastructure, not tied to story-block
- Story-block's MCP server already runs on SSE at `localhost:8082` with 6 tools + 2 resources
- Copilot repo symlinked into story-block at `obsidian-copilot/` (not tracked by git)

## Dev Loop

```
1. Edit code in WSL:  ~/code/obsidian-copilot/src/
2. Bump version:      Update manifest.json version (3.2.5-mcp.N)
3. Build + deploy:    make copilot                    # default vault (Obsidian/writing)
                      make copilot VAULT=other-vault  # specific vault
4. Reload:            In Obsidian → Settings → Community plugins → disable/enable Copilot (MCP Fork)
5. Test:              Use Copilot chat in Obsidian with MCP server running (make mcp)
6. Repeat
```

**Make targets:**
- `make copilot` — build + deploy to default vault
- `make copilot VAULT=my-vault` — build + deploy to specific vault (child of `Documents/notes/`)
- `make copilot-build` — build only
- `make copilot-deploy` — deploy only (skip build)

**Notes:**
- Plugin name shows as **"Copilot (MCP Fork)"** to confirm custom build
- `data.json` in the vault plugin dir is NOT overwritten — it holds user settings
- All vaults assumed children of `/mnt/c/Users/Lucas Badico/Documents/notes/`
- MCP server must be running: `make mcp` (SSE on `:8082`)
- Ollama must be running on Windows for chat model + embeddings

## Architecture Analysis

### Existing Tool System
- **ToolRegistry** (`src/tools/ToolRegistry.ts`): singleton registry with `ToolDefinition` = LangChain `StructuredTool` + `ToolMetadata`
- **Category `"mcp"` already exists** in `ToolMetadata.category` type — but nothing uses it yet
- **AutonomousAgentChainRunner**: ReAct loop with `bindTools()` — up to 4 iterations, native tool calling
- **CopilotPlusChainRunner**: single-pass tool planning + execution
- **createLangChainTool** (`src/tools/createLangChainTool.ts`): helper to create LangChain tools from Zod schemas
- **LangChain integration**: tools use `@langchain/core/tools` `StructuredTool`, Zod for validation

### What Needs to Be Built

#### 1. MCP Client Transport (`src/mcp/McpClient.ts`)
- SSE transport client (connect to `http://<host>:<port>/sse`)
- stdio transport client (for local MCP servers)
- Connection lifecycle: connect, reconnect, disconnect
- Tool discovery: `tools/list` → parse MCP tool definitions
- Tool execution: `tools/call` → invoke MCP tool with args, return result

#### 2. MCP → LangChain Tool Bridge (`src/mcp/McpToolBridge.ts`)
- Convert MCP tool JSON Schema → Zod schema (for LangChain compatibility)
- Create `StructuredTool` for each MCP tool via `createLangChainTool`
- Register each tool in `ToolRegistry` with `category: "mcp"`
- Handle MCP tool results → format for LangChain `ToolMessage`

#### 3. MCP Server Manager (`src/mcp/McpServerManager.ts`)
- Manage multiple MCP server connections
- Configuration: list of servers with name, transport type, URL/command
- Auto-reconnect on disconnect
- Tool refresh on reconnect (server tools may change)

#### 4. Settings UI
- MCP servers configuration panel
- Add/edit/remove MCP servers
- Per-server: name, transport (SSE/stdio), URL/command, enabled toggle
- Connection status indicator

#### 5. Plugin Integration
- Wire MCP server manager into `main.ts` plugin lifecycle
- Connect on plugin load, disconnect on unload
- Register MCP tools alongside builtin tools

### Data Flow
```
MCP Server (SSE at :8082)
    ↕ SSE/JSON-RPC
McpClient (transport layer)
    ↕ MCP protocol
McpToolBridge (schema conversion)
    ↕ LangChain StructuredTool
ToolRegistry (category: "mcp")
    ↕ bindTools()
AutonomousAgentChainRunner (ReAct loop)
    ↕ streaming
Chat UI
```

## Scope

### Phase 1 — Core MCP Client (MVP)
- [ ] MCP SSE transport client (connect, tool discovery, tool execution)
- [ ] JSON Schema → Zod conversion for MCP tool parameters
- [ ] MCP tool → LangChain StructuredTool bridge
- [ ] Register MCP tools in ToolRegistry with `category: "mcp"`
- [ ] Settings: MCP server URL configuration (single server)
- [ ] Plugin lifecycle: connect on load, disconnect on unload
- [ ] Test with story-block MCP server (6 tools)

### Phase 2 — Multi-Server + stdio
- [ ] Multiple MCP server support
- [ ] stdio transport for local MCP servers
- [ ] Settings UI: add/edit/remove servers
- [ ] Connection status indicators
- [ ] Auto-reconnect

### Phase 3 — Resource Support
- [ ] MCP resource discovery (`resources/list`)
- [ ] MCP resource reading (`resources/read`)
- [ ] Surface resources in Copilot UI (e.g., as context sources)

### Tests
- [ ] MCP SSE client: connection, tool discovery, tool execution
- [ ] JSON Schema → Zod: string, number, boolean, array, object, enum, optional
- [ ] MCP tool bridge: creates valid LangChain tools from MCP tool definitions
- [ ] ToolRegistry: MCP tools registered with correct category
- [ ] Agent loop: MCP tools available and callable in ReAct flow

## Dev Log

| # | Date | Event |
|---|------|-------|
| 1 | 2026-03-25 | Task created. Copilot repo analyzed — ToolRegistry already has `"mcp"` category, LangChain tool system is well-structured. Architecture plan defined. Issue #1451 open in upstream. |
| 2 | 2026-03-25 | Dev loop established. Copilot forked, manifest tagged as "Copilot (MCP Fork)" v3.2.5-mcp.1. Build + deploy to Obsidian vault at `Documents/notes/Obsidian/writing`. Custom name confirmed visible in Obsidian community plugins. Vault shared via symlink: Windows `C:\Users\Lucas Badico\Documents\notes` → WSL `~/notes`. |
| 3 | 2026-03-25 | **Phase 1 core built.** 3 new files: McpClient.ts (SSE transport), McpToolBridge.ts (JSON Schema→Zod→LangChain), McpServerManager.ts (lifecycle). 72 tests passing. Dependencies: `@modelcontextprotocol/sdk`, `@n8n/json-schema-to-zod`. Settings: `mcpServers` config + MCP tab in settings UI. Plugin wiring in main.ts (connect on load, disconnect on unload, reconnect on settings change). |
| 4 | 2026-03-25 | **Plus checks bypassed.** `useIsPlusUser()` returns true, `isPlusChain()` returns true, agent runner Plus check bypassed. All premium features unlocked: agent mode, web search, PDF/image context, projects. UI labels: "(free)" removed, "copilot plus" → "agent", sparkles icon → bot icon. |
| 5 | 2026-03-25 | **CORS fix on MCP server.** Obsidian sends from `app://obsidian.md` origin. Added `corsMiddleware()` in `mcp-server/cmd/mcp/main.go` wrapping SSE handler. Allowed headers: Content-Type, Authorization, mcp-protocol-version. |
| 6 | 2026-03-25 | **MCP tools registered successfully.** 6 tools discovered and registered: get_canon, get_character_state, get_plots, get_timeline, get_writing_context, search_story_memory. Confirmed in Obsidian dev console. |
| 7 | 2026-03-25 | **Schema validation fix.** Agent sent empty `{}` to get_canon but `as_of_chapter` was required. Made all `as_of_chapter` params optional across all MCP tools (canon, timeline, plot, character) with default 9999 (show all). |
| 8 | 2026-03-25 | **Tool selection fix.** Mistral Nemo defaulted to vault search for story queries. Added MCP-specific system prompt guidance in AGENT_LOOP_GUIDANCE mapping question types to MCP tools. Tool IDs shortened from UUID-based to name-based (`mcp_storyblock_get_canon`). |
| 9 | 2026-03-25 | **End-to-end milestone achieved.** Full pipeline working: User asks "what is the canon of my story?" in Obsidian Copilot → Agent selects get_canon MCP tool → SSE call to MCP server (:8082) → Gateway API (:8081) → PostgreSQL → Canon response with artifacts, character states, relationships, timeline events, plot threads returned to chat. v3.2.5-mcp.13. |

## Decisions & Rationale

| # | Decision | Chosen | Why |
|---|----------|--------|-----|
| 1 | Generic MCP client, not story-block specific | Upstream PR | AGPL license requires open-sourcing. Generic feature benefits everyone. Story-block server is just one MCP server among many. |
| 2 | SSE transport first | Phase 1 priority | Story-block MCP server uses SSE. stdio can come in Phase 2. |
| 3 | Bridge MCP → LangChain tools | Reuse existing agent infrastructure | AutonomousAgentChainRunner already handles tool discovery, binding, execution, and ReAct loops. MCP tools just need to be registered as LangChain tools. |
| 4 | JSON Schema → Zod conversion | Required by LangChain tool interface | LangChain uses Zod for tool parameter validation. MCP uses JSON Schema. Need a bridge. |

## Difficulties & Solutions

| # | Problem | Solution | Time |
|---|---------|----------|------|
| 1 | CORS blocks SSE from Obsidian (`app://obsidian.md` origin) | Added `corsMiddleware()` to MCP server with `Access-Control-Allow-Origin: *` and `mcp-protocol-version` header | 15 min |
| 2 | `logInfo` silent in Obsidian console (requires debug mode) | Added temporary `console.log` with `[MCP]` prefix for connection debugging | 5 min |
| 3 | Agent sends empty `{}` args, Zod rejects (required `as_of_chapter`) | Made all `as_of_chapter` params optional in MCP server tools, default 9999 | 10 min |
| 4 | Agent picks vault search over MCP tools for story queries | Added MCP-specific guidance in `AGENT_LOOP_GUIDANCE` system prompt mapping question types to tools | 10 min |
| 5 | UUID-based tool IDs (`mcp_1f446e47-...`) confuse small models | Changed to name-based IDs (`mcp_storyblock_get_canon`) using sanitized server name | 5 min |
| 6 | `initializeBuiltinTools()` calls `registry.clear()` wiping MCP tools | Added `clearBuiltins()` to ToolRegistry that preserves `category: "mcp"` tools | 5 min |

## Acceptance Criteria

- [ ] Configure MCP server URL in Copilot settings → plugin connects on enable
- [ ] MCP tools appear in agent tool list alongside builtin tools
- [ ] Ask "What is Haruka's emotional state?" in Copilot chat → agent calls `get_canon` or `get_character_state` via MCP → returns story data
- [ ] MCP server disconnect → tools gracefully removed, no crash
- [ ] MCP server reconnect → tools re-registered
- [ ] PR submitted to logancyang/obsidian-copilot addressing issue #1451
