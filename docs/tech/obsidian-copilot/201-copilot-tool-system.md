# 201 — Copilot Tool System Reference

## Overview

How Obsidian Copilot's tool system works, from registration to execution. This is the system we integrate MCP tools into.

## Architecture

```
Tool Definition (Zod schema + handler)
    |
    v  createLangChainTool()
LangChain StructuredTool
    |
    v  registry.register({ tool, metadata })
ToolRegistry (singleton)
    |
    v  registry.getEnabledTools(enabledIds, vaultAvailable)
StructuredTool[] (filtered by settings)
    |
    v  chatModel.bindTools(tools)
Bound Model (tool schemas in LLM context)
    |
    v  ReAct loop (up to 4 iterations)
Tool Execution -> ToolMessage -> AI Response
```

## Key Files

| File | Purpose |
|---|---|
| `src/tools/ToolRegistry.ts` | Singleton registry — stores ToolDefinitions, filters by settings |
| `src/tools/createLangChainTool.ts` | Helper to create LangChain `StructuredTool` from Zod schema + handler |
| `src/tools/builtinTools.ts` | All builtin tool definitions + `initializeBuiltinTools()` |
| `src/tools/SearchTools.ts` | localSearch + webSearch tool implementations |
| `src/tools/TimeTools.ts` | Time-related tools |
| `src/tools/FileTreeTools.ts` | Vault file tree tool |
| `src/tools/memoryTools.ts` | User memory tool |
| `src/tools/ObsidianCliTools.ts` | Desktop-only CLI tools |
| `src/tools/toolManager.ts` | Lightweight execution utility (not registration) |
| `src/LLMProviders/chainRunner/AutonomousAgentChainRunner.ts` | ReAct agent loop — uses tools |
| `src/LLMProviders/chainRunner/CopilotPlusChainRunner.ts` | Single-pass tool planning — uses tools |

## ToolMetadata Interface

```typescript
// src/tools/ToolRegistry.ts:7-21
interface ToolMetadata {
  id: string;
  displayName: string;
  description: string;
  category: "search" | "time" | "file" | "media" | "mcp" | "memory" | "custom" | "cli";
  isAlwaysEnabled?: boolean;    // Always available (e.g. time tools)
  requiresVault?: boolean;      // Needs vault access
  customPromptInstructions?: string;  // Extra LLM guidance for this tool
  copilotCommands?: string[];   // @command aliases (e.g. @vault)
  timeoutMs?: number;           // Execution timeout
  isBackground?: boolean;       // Hidden from user
  isPlusOnly?: boolean;         // Requires Plus subscription
  requiresUserMessageContent?: boolean;  // Receives original user message
}
```

## Tool Registration Patterns

### Pattern A: Static Array (most tools)

```typescript
// builtinTools.ts:31-242
export const BUILTIN_TOOLS: ToolDefinition[] = [
  { tool: localSearchTool, metadata: { id: "localSearch", category: "search", ... } },
  { tool: webSearchTool, metadata: { id: "webSearch", category: "search", ... } },
  { tool: getCurrentTimeTool, metadata: { id: "getCurrentTime", category: "time", ... } },
  // ...
];
```

### Pattern B: Conditional (runtime-dependent)

```typescript
// builtinTools.ts:247-272
export function registerFileTreeTool(vault: Vault): void {
  const registry = ToolRegistry.getInstance();
  registry.register({
    tool: createGetFileTreeTool(vault.getRoot()),
    metadata: { id: "getFileTree", category: "file", ... }
  });
}

// builtinTools.ts:306-326
export function registerMemoryTool(): void {
  registry.register({ tool: updateMemoryTool, metadata: { id: "updateMemory", ... } });
}

// builtinTools.ts:332-478
export function registerCliTools(): void {
  // Desktop-only tools
  registry.register(dailyNoteTool);
  registry.register(randomReadTool);
}
```

### Pattern C: Initialization entry point

```typescript
// builtinTools.ts:487-524
export function initializeBuiltinTools(vault?: Vault): void {
  const registry = ToolRegistry.getInstance();
  registry.clear();
  registry.registerAll(BUILTIN_TOOLS);

  if (vault) {
    registerFileTreeTool(vault);
    registerTagListTool();
  }
  if (settings.enableSavedMemory) registerMemoryTool();
  if (isDesktopRuntime()) registerCliTools();
}
```

## How Chain Runners Get Tools

### AutonomousAgentChainRunner (ReAct loop)

```typescript
// AutonomousAgentChainRunner.ts:128-141
private getAvailableTools(): StructuredTool[] {
  const registry = ToolRegistry.getInstance();
  if (registry.getAllTools().length === 0) {
    initializeBuiltinTools(this.chainManager.app?.vault);
  }
  const enabledToolIds = new Set(settings.autonomousAgentEnabledToolIds || []);
  return registry.getEnabledTools(enabledToolIds, !!this.chainManager.app?.vault);
}
```

Tools are bound to the model and used in the ReAct loop:

```typescript
// AutonomousAgentChainRunner.ts:545-553
const availableTools = this.getAvailableTools();
const boundModel = chatModel.bindTools(availableTools);

// ReAct loop: model outputs tool_calls -> execute -> ToolMessage -> repeat
while (iteration < 4) {
  const response = await boundModel.invoke(messages);
  if (!response.tool_calls?.length) break;  // Final response
  // Execute tools, add ToolMessages, continue
}
```

### CopilotPlusChainRunner (single-pass)

Only uses a subset of tools for planning:

```typescript
// CopilotPlusChainRunner.ts:81-103
protected getAvailableToolsForPlanning(): StructuredTool[] {
  const allTools = registry.getAllTools().map(def => def.tool);
  return allTools.filter(tool =>
    ["getCurrentTime", "convertTimeBetweenTimezones",
     "getTimeInfoByEpoch", "getTimeRangeMs", "getFileTree"].includes(tool.name)
  );
}
```

## Creating a Tool (Full Example)

```typescript
import { z } from "zod";
import { createLangChainTool } from "@/tools/createLangChainTool";
import { ToolRegistry, ToolDefinition } from "@/tools/ToolRegistry";

// 1. Define Zod schema
const mySchema = z.object({
  query: z.string().min(1).describe("Search query"),
  limit: z.number().int().optional().describe("Max results"),
});

// 2. Create LangChain tool
const myTool = createLangChainTool({
  name: "myToolName",
  description: "What this tool does — the LLM reads this to decide when to use it",
  schema: mySchema,
  func: async ({ query, limit }) => {
    const result = await doSomething(query, limit);
    return JSON.stringify(result);
  },
});

// 3. Register with metadata
const registry = ToolRegistry.getInstance();
registry.register({
  tool: myTool,
  metadata: {
    id: "myToolName",
    displayName: "My Tool",
    description: "Human-readable description for settings UI",
    category: "custom",
  },
});
```

## Settings Integration

Tools are enabled/disabled via `settings.autonomousAgentEnabledToolIds` — an array of tool IDs. The `getEnabledTools()` method filters:

```typescript
// ToolRegistry.ts:72-97
getEnabledTools(enabledToolIds: Set<string>, vaultAvailable: boolean): StructuredTool[] {
  return Array.from(this.tools.values())
    .filter(def => {
      if (def.metadata.isAlwaysEnabled) return true;
      if (def.metadata.requiresVault && !vaultAvailable) return false;
      return enabledToolIds.has(def.metadata.id);
    })
    .map(def => def.tool);
}
```

## Test Patterns

`allTools.validation.test.ts` validates:
- No `z.any()` usage (weak typing)
- All tools have name, description, schema
- Array elements have specific types
- Optional fields properly marked
- Meaningful validation error messages

## Key Takeaways for MCP Integration

1. **ToolRegistry already supports `"mcp"` category** — no changes needed
2. **`createLangChainTool()` is the standard factory** — MCP tools use the same path
3. **Conditional registration pattern** (Pattern B) fits MCP — register when server connects, unregister on disconnect
4. **`getEnabledTools()` filtering** means MCP tools need to be in `autonomousAgentEnabledToolIds` or use `isAlwaysEnabled: true`
5. **ReAct loop handles everything** after registration — no agent changes needed
6. **30s timeout per tool** is the existing convention — matches well for MCP network calls
