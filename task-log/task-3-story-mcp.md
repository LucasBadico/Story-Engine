# Task 3 — Story Engine MCP (Thin Retrieval Wrapper)

## Goal
Build an MCP server that wraps the LLM Gateway's retrieval capabilities so a local LLM (via LM Studio chat) can query story memory through MCP protocol. The MCP is a **thin translation layer** — it maps MCP tools/resources to gateway HTTP endpoints. All heavy logic (extraction, embeddings, state tracking, timeline, semantic search) lives in the gateway.

## Prerequisites
- Task 1 complete — main service has entities + new state/timeline/reveal data, gateway has ingested + indexed them
- LLM Gateway running with new retrieval endpoints (added in this task)
- MCP SDK (TypeScript or Go)

## System Architecture

```
LM Studio Chat + LLM
        │ MCP protocol (stdio or SSE)
        ▼
  MCP Server (thin wrapper)
  │  - translates MCP tool calls → gateway HTTP calls
  │  - translates MCP resource reads → gateway HTTP calls
  │  - assembles context packs from multiple gateway responses
  │  - enforces retrieval mode (timeline-consistent / revealed / omniscient)
        │ HTTP API
        ▼
  LLM Gateway (retrieval layer)
  │  - semantic search (pgvector)
  │  - entity/concept lookup
  │  - state snapshot queries
  │  - timeline event queries
  │  - concept graph traversal
  │  - canon retrieval endpoint
        │ SQL + pgvector          │ gRPC (ingestion)
        ▼                         ▼
  Gateway Postgres + pgvector    Main Service (source of truth)
                                 │  - Character, Location, Artifact CRUD
                                 │  - CharacterState, RelationshipState (per chapter)
                                 │  - TimelineEvent, Reveal
                                 │  - publishes to gateway via gRPC + Redis queue
```

### What lives where

| Concern | Main Service | LLM Gateway | MCP |
|---|---|---|---|
| Entity CRUD (characters, locations, etc.) | Source of truth | Ingests via gRPC | — |
| Character/relationship state per chapter | Source of truth | `state_snapshots` (retrieval-optimized) | Queries gateway |
| Timeline events, reveals | Source of truth | `events`, `reveals` (retrieval-optimized) | Queries gateway |
| Embeddings + summaries | — | Generates + stores | — |
| Semantic search (pgvector) | — | Owns | Queries gateway |
| Concept graph | — | `concepts` + `concept_links` | Queries gateway |
| Canon versioning (valid_from/to) | — | Retrieval concern | Queries gateway |
| Dual-axis timeline filtering | — | SQL WHERE on chapter + day | Passes params to gateway |
| Retrieval mode logic | — | — | Decides what params to pass |
| Context assembly | — | — | Combines multiple responses |
| MCP protocol (stdio) | — | — | Owns |

## Dual Timeline Model

Two independent axes — chapter number (always present) and in-world timeline (optional):

| Axis | Always Present | Controls | Example |
|---|---|---|---|
| **Chapter number** (primary) | Yes | What reader/author knows | "At chapter 5, reader hasn't seen chapter 7 reveals" |
| **In-world timeline_day** (secondary) | No — nullable | What's true in the story world | "On Day -40, Haruka is still a hopeful wife" |

`timeline_confidence`: `explicit` (author stated) | `inferred` (LLM-derived, future) | NULL (absent)

## Knowledge Layers (retrieval contract)

| # | Layer | Gateway Source | Chapter Filter | Timeline Day Filter |
|---|---|---|---|---|
| 1 | Canon | `documents` (is_canonical=true) + `concepts` | `valid_from_chapter <= N AND (valid_to_chapter IS NULL OR valid_to_chapter > N)` | N/A — canon versioned by chapter only |
| 2 | Timeline | `events` | `chapter_number <= N` | `timeline_day <= D` (when available) |
| 3 | State | `state_snapshots` | `chapter_number <= N` | `timeline_day <= D` (fallback to chapter when NULL) |
| 4 | Reveal | `reveals` | `chapter_number <= N` | N/A — reveals are chapter-scoped |

## Retrieval Modes

| Mode | Canon Filter | State/Events Filter | Reveals Filter | Use Case |
|---|---|---|---|---|
| revealed_narrative (default) | `valid_from_chapter <= N` | `chapter_number <= N` | `chapter_number <= N` | Spoiler-safe writing/analysis |
| timeline_consistent | all canon | `timeline_day <= D` (fallback to chapter) | all | Writing a scene at a specific in-world moment |
| omniscient | all | all | all | Planning, editing, full picture |

## Scope

### Gateway: New Retrieval Endpoints (Go)

- [ ] `GET /api/v1/stories/{id}/identity` — premise, themes, metadata
- [ ] `GET /api/v1/stories/{id}/world-rules` — canon world rules
- [ ] `GET /api/v1/stories/{id}/characters` — list all character concepts
- [ ] `GET /api/v1/stories/{id}/characters/{name}/state?as_of_chapter=N&as_of_day=D&mode=M` — character state (chapter always required, day optional for timeline-consistent mode)
- [ ] `GET /api/v1/stories/{id}/relationships/{char_a}/{char_b}?as_of_chapter=N&as_of_day=D&mode=M` — relationship state with history
- [ ] `GET /api/v1/stories/{id}/state-changes?as_of_chapter=N&as_of_day=D&mode=M&limit=K` — recent state changes
- [ ] `GET /api/v1/stories/{id}/events?as_of_chapter=N&as_of_day=D&mode=M&limit=K` — timeline events
- [ ] `GET /api/v1/stories/{id}/scenes/{chapter}/{scene}` — scene context (characters, location, conflicts, recent events)
- [ ] `GET /api/v1/stories/{id}/concepts/{name}/graph?depth=N` — concept graph traversal
- [ ] `POST /api/v1/search` — extend existing search with: `as_of_chapter`, `as_of_day`, `mode`, `concept_tags[]` filters
- [ ] `GET /api/v1/stories/{id}/chapters/{num}/summary` — chapter summary

### MCP Server

#### Resources (stable knowledge objects)
- [ ] `story://stories/{id}/identity` → `GET /stories/{id}/identity`
- [ ] `story://stories/{id}/characters/{name}` → `GET /stories/{id}/characters/{name}/state` (omniscient)
- [ ] `story://stories/{id}/relationships/{a}_{b}` → `GET /stories/{id}/relationships/{a}/{b}` (omniscient)
- [ ] `story://stories/{id}/world-rules` → `GET /stories/{id}/world-rules`
- [ ] `story://stories/{id}/chapters/{num}/summary` → `GET /stories/{id}/chapters/{num}/summary`
- [ ] `story://stories/{id}/locations/{name}` → `GET /stories/{id}/concepts/{name}/graph?depth=0`
- [ ] `story://stories/{id}/artifacts/{name}` → `GET /stories/{id}/concepts/{name}/graph?depth=0`

#### Tools (dynamic retrieval)
- [ ] `search_story_memory(story_id, query, as_of_chapter, as_of_day?, mode?)` → `POST /search` with dual-axis params
- [ ] `get_character_state(story_id, character_name, as_of_chapter, as_of_day?, mode?)` → `GET /characters/{name}/state`
- [ ] `get_relationship_state(story_id, char_a, char_b, as_of_chapter, as_of_day?, mode?)` → `GET /relationships/{a}/{b}`
- [ ] `get_recent_state_changes(story_id, as_of_chapter, as_of_day?, mode?, limit?)` → `GET /state-changes`
- [ ] `get_scene_context(story_id, chapter_number, scene_number)` → `GET /scenes/{chapter}/{scene}`
- [ ] `connect_story_concepts(story_id, concept_name, depth?)` → `GET /concepts/{name}/graph`

#### Context Assembly
- [ ] `get_writing_context(story_id, chapter, scene)` — composite tool that calls multiple gateway endpoints and assembles a full context pack: story identity + scene characters + their states + active relationships + recent events + relevant world rules + relevant semantic chunks
- [ ] Token budget awareness — truncate/prioritize by relevance when context exceeds limit

#### Transport
- [x] stdio transport for local/Claude Desktop integration
- [x] SSE transport for LM Studio on Windows (`--sse=:8082`)
- [x] Configuration: gateway URL, default story_id, default tenant_id (flags + env vars)

### Tests
- [ ] MCP tool → gateway HTTP mapping: each tool calls correct endpoint with correct params
- [ ] Timeline-consistent mode: `get_character_state` at Day -40 returns "hopeful wife", not "collared" (Day 0)
- [ ] Timeline-consistent mode: entities without timeline_day fall back to chapter filter
- [ ] Revealed-narrative mode: chapter 3 query doesn't return chapter 7 secrets
- [ ] Omniscient mode: everything returned
- [ ] Context assembly: `get_writing_context` returns structured pack
- [ ] Context assembly: respects token budget
- [ ] Concept graph: `connect_story_concepts("crimson ribbon")` returns connected entities
- [ ] Resource reads return stable canonical data

## Dev Log

| # | Date | Event |
|---|------|-------|
| 1 | 2026-03-16 | MCP server v0.1 built. Go module `mcp-server/` using `mcp-go` v0.28.0 (stdio transport). 5 tools: search_story_memory, get_canon, get_character_state, get_timeline, get_writing_context. 2 resources: story identity, world rules. Gateway HTTP client matches actual canon endpoint response shape (canon/states/events/reveals arrays). Builds + vets clean. Simplified from original spec: uses single canon endpoint instead of 11 separate gateway endpoints — canon assembles all layers server-side. |
| 2 | 2026-03-16 | Session ended. Next session: local setup (task 2) then end-to-end test of full pipeline: import CLI → main service → Redis → gateway ingest → canon retrieval → MCP tools. |
| 3 | 2026-03-18 | Reveals removed from MCP: `RevealEntry` DTO deleted from gateway client, reveals sections removed from canon tool and writing_context tool. Reveals replaced by plots system (task-1 refactor). |
| 4 | 2026-03-18 | New `get_plots` tool added (6 tools total). Filters canon by `source_type=plot` and `plot_development`, groups developments under parent plot. Plots are versioned canonical documents — gateway does the LLM synthesis, MCP just reads and presents. |
| 5 | 2026-03-25 | **End-to-end pipeline validated.** Full chain working: Import CLI → Main Service API → Redis (DB 1) → Gateway Worker (embeddings via Ollama + summaries via Gemini) → Canon endpoint → MCP tools. 32 documents, 201 embedding chunks, 18 canonical docs. `get_canon` tool returns complete story data. |
| 6 | 2026-03-25 | **`is_canonical` bug fixed.** `DocumentRepository.Create` and `Update` were missing `valid_from_chapter`, `valid_to_chapter`, `is_canonical`, `chapter_number`, `story_id` columns in SQL. All 5 fields now persisted. Added `IsCanonical = true` to ingesters: story, chapter, character, location, artifact, world, plot. Added `StoryID`/`ChapterNumber` to chapter, content_block, character_state, relationship_state, timeline_event ingesters. `ListCanonical` query updated to include world entities (no `story_id`) via tenant fallback. |
| 7 | 2026-03-25 | **SSE transport added.** `--sse=:8082` flag starts HTTP/SSE server instead of stdio. Needed because LM Studio runs on Windows, can't use stdio via `wsl.exe`. LM Studio connects to `http://localhost:8082/sse`. Root Makefile updated: `make mcp` starts SSE on 8082. |
| 8 | 2026-03-25 | **Queue bugs fixed.** (a) Removed `SourceTypeScene`/`SourceTypeBeat` from `storySourceTypes` — no ingesters exist, items were stuck in pop/release loop. (b) `Ack()` in processing_queue.go now removes from both processing AND queue sets via Redis pipeline — prevents re-queued items from being stuck. |
| 9 | 2026-03-25 | **MCP basic milestone complete.** All 6 tools + 2 resources working. Full pipeline validated with real story data (Akiba Haruka, Prologue/Chapter 0). Next steps for writing interface: (a) VS Code + Continue/Cline with MCP + xAI Grok API, (b) **Obsidian chat plugins with MCP** — Obsidian Copilot, Smart Connections, or Obsidian MCP plugin could connect to SSE at `localhost:8082` + Grok API for uncensored creative analysis. Obsidian is ideal since story files are already markdown. Needs research: which plugin supports both MCP and custom API endpoints (xAI). |

## Decisions & Rationale

| # | Decision | Chosen | Why |
|---|----------|--------|-----|
| 1 | MCP is thin wrapper, not its own service | All logic in gateway | Gateway already has DB, embeddings, search — MCP just translates protocol |
| 2 | Retrieval mode as parameter, not separate tools | Single tool set with `mode` param | Simpler API surface — same tool works for writing, analysis, and planning |
| 3 | Context assembly in MCP not gateway | MCP combines multiple responses | Gateway endpoints stay single-purpose; MCP knows what the LLM needs |
| 4 | Dual transport: stdio + SSE | stdio for local, SSE for cross-OS | LM Studio on Windows can't pipe stdio through wsl.exe reliably. SSE on :8082 works across WSL2 boundary. |
| 5 | Gateway gets new retrieval endpoints | REST over existing DB | Gateway already owns the data — adding read endpoints is natural |
| 6 | No prose generation in MCP | Retrieval only | MCP provides memory, LLM provides creativity — clear separation |
| 7 | Token budget in MCP | MCP truncates context packs | Gateway doesn't know the LLM's context window — MCP does |
| 8 | Dual timeline: chapter (primary) + day (secondary) | `as_of_chapter` required, `as_of_day` optional on all tools | Chapter number always known; in-world timeline is story-dependent. Fallback ensures no data disappears. |
| 9 | timeline_confidence in schema | `explicit` / `inferred` / NULL | Schema ready for future LLM-inferred timelines without migration |

## Difficulties & Solutions

| # | Problem | Solution | Time |
|---|---------|----------|------|
| 1 | `is_canonical` always false | `DocumentRepository.Create/Update` SQL missing 5 columns added in migration 009. Fixed INSERT/UPDATE to include all canon fields. | 30 min |
| 2 | World entities (character, location, artifact) have no `story_id` | Updated `ListCanonical` to fallback to `tenant_id` match when `story_id` is NULL | 10 min |
| 3 | Scenes stuck in queue loop | `SourceTypeScene`/`SourceTypeBeat` in `storySourceTypes` but no ingester. Removed from array. | 5 min |
| 4 | Content block re-queued after processing | `Ack()` only removed from processing set, not queue set. Fixed to remove from both. | 5 min |
| 5 | LM Studio stdio via wsl.exe exits immediately | wsl.exe closes stdin pipe. Added SSE transport as alternative. | 15 min |

## Patterns Discovered

## Acceptance Criteria

- [ ] From LM Studio chat with MCP connected, ask "What is Haruka's emotional state in chapter 3?" → MCP returns character state snapshot (revealed-narrative, chapter-based)
- [ ] Ask "What is Haruka's state on Day -40?" → MCP returns timeline-consistent state (uses timeline_day filter, falls back to chapter for entities without it)
- [ ] Ask "What relationship context matters for Haruka and Tetsuo?" → returns relationship baseline + state changes
- [ ] Ask "What world rules apply?" → returns canon world rules
- [ ] Ask "What connects to the crimson ribbon?" → concept graph traversal returns connected entities/themes
- [ ] `get_writing_context(story, chapter, scene)` returns assembled context pack with all relevant data
- [ ] Timeline-consistent mode excludes future states
- [ ] Revealed-narrative mode excludes unrevealed chapter data
- [ ] Omniscient mode returns everything
- [ ] MCP connects to LM Studio via stdio
- [ ] All MCP tools return structured JSON, not prose
- [ ] Context packs stay within configurable token budget
