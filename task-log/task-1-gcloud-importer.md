# Task 1 — Google Drive Chapter Importer

## Goal
Build a file ingestion pipeline that reads Google Drive exported chapter folders and imports them into the **main service** (source of truth for entities), which then publishes to the **LLM Gateway** for embedding + retrieval via the existing gRPC + Redis queue pattern. The Obsidian plugin triggers the import.

## Architecture Decision: Main Service as Source of Truth

**Revised**: originally the plan was to bypass main service and have the gateway own everything. This was reversed because:

- The gateway already ingests FROM main service via gRPC (11 existing ingesters)
- Retrieval results must trace back to a canonical source — if the MCP returns "Haruka at chapter 5", the LLM needs to verify/update that in main service, not a gateway shadow copy
- Main service already owns Character, Location, Artifact, Story, Chapter, Scene, ContentBlock entities
- Building a second source of truth in the gateway creates divergence

**Flow:**
```
Google Drive Export
        │
        ▼
  Parser (reads v2 export files)
        │
        ▼
  Main Service (CRUD — source of truth)
  │  ├── Existing entities: characters, locations, artifacts, stories, chapters, scenes, content blocks
  │  └── New entities: character states per chapter, timeline events, reveals, relationship states
  │       │
  │       ▼ publishes via gRPC + Redis queue (existing pattern)
  │
  LLM Gateway (embed + chunk + summarize + index)
  │  ├── embedding_documents + embedding_chunks (existing)
  │  ├── Retrieval metadata: canon versioning, dual-axis timeline, concepts, state_snapshots
  │  ├── Summary generation (3 per entity) ← REUSE
  │  └── Entity extraction on prose ← REUSE
  │       │
  │       ▼
  │  Canon Retrieval Endpoint
  │  GET /api/v1/stories/{id}/canon?as_of_chapter=N
```

### What lives where

| Concern | Main Service (source of truth) | LLM Gateway (retrieval layer) |
|---|---|---|
| Character CRUD | ✓ (existing) | Ingests via gRPC |
| Location CRUD | ✓ (existing) | Ingests via gRPC |
| Artifact CRUD | ✓ (existing) | Ingests via gRPC |
| Story/Chapter/Scene CRUD | ✓ (existing) | Ingests via gRPC |
| Content Blocks | ✓ (existing) | Ingests via gRPC |
| **Character State per Chapter** | **✓ (NEW)** | Ingests via gRPC → `state_snapshots` |
| **Relationship State per Chapter** | **✓ (NEW)** | Ingests via gRPC → `state_snapshots` |
| **Timeline Events** | **✓ (NEW)** | Ingests via gRPC → `events` |
| **Reveals** | **✓ (NEW)** | Ingests via gRPC → `reveals` |
| Canon versioning (valid_from/to) | — | ✓ (retrieval concern) |
| Embeddings + summaries | — | ✓ (existing) |
| Semantic search | — | ✓ (existing) |
| Concept graph (concepts + links) | — | ✓ (retrieval-optimized) |
| Entity extraction from prose | — | ✓ (existing 8-phase pipeline) |

## Prerequisites
- Main service running with existing entity CRUD
- LLM Gateway running with existing ingestion pipeline (gRPC + Redis queue → embed + summarize)
- Google Drive export tool producing v2 folder structure (see `data/gcloud-exported-chapter/`)
- Obsidian plugin with API client

## Export Structure (input — v2 format)

```
gcloud-exported-chapter/
├── tabs_index.txt              # metadata: doc name, doc ID, export timestamp, tab list
├── 01_story_canon.md           # Story Canon: premise, themes, world rules, characters, relationships, locations, artifacts
├── 02_chapter_log.md           # Chapter Log: character states, relationship states, scenes, secrets, conflicts, timeline, arc context
├── 03_content.md               # Content: actual prose
├── 04_notes.md                 # Notes: author editorial notes with dates
└── story_canon/                # Granular: each canon section as its own file
    ├── 01_premise.md
    ├── 02_themes.md
    ├── 03_world_rules.md
    ├── 04_character_{name}.md
    ├── 09_location_{name}.md
    ├── 11_relationship_{a}_{b}.md
    ├── 15_artifact_{name}.md
    └── ...
```

### Key v2 format decisions (see `data/gcloud-exported-chapter/CHANGES.md`)
- "Story Bible" renamed to **Story Canon** — identity only, no state
- "Log" renamed to **Chapter Log** — all per-chapter state lives here
- "Key Objects" renamed to **Artifacts** — matches gateway entity type `artefact`
- Character profiles stripped of "current state" → moved to Chapter Log
- Relationship baselines stripped of tension → moved to Chapter Log as "Relationship States"
- New section "New Information Revealed" in Chapter Log → feeds Reveal layer

## Dual Timeline Model

Story data is indexed on **two independent axes**. Chapter number is always present. In-world timeline is optional.

### Chapter Number (primary axis — always present)
- Order the author writes/exports: Prologue = 0, Chapter 1 = 1, Chapter 2 = 2...
- **Always known** at ingestion time (passed as parameter)
- Controls: what the reader/author knows at a given point

### In-World Timeline (secondary axis — optional)
- Chronological order in the story universe: Day -63, Day -40, Day 0...
- **Not always present** — depends on whether the author provides explicit timeline markers
- Parsed from Chapter Log `## Timeline` section when available
- `timeline_confidence`: `explicit` (author stated it) | `inferred` (LLM-derived, future feature) | NULL (absent)
- Controls: what's true in the story world at a given moment

### Why both matter

Stories can be non-linear. The prologue (chapter 0) shows Day 0 (collaring), then chapters 1-7 go back to Day -63 through Day -1:

| Chapter | In-World Day | Haruka's State |
|---|---|---|
| Prologue (ch 0) | Day 0 | collared, surrendered |
| Chapter 1 | Day -63 | hopeful wife |
| Chapter 5 | Day -40 | cracking |
| Chapter 8 | Day -1 | experiment begins |

### Retrieval Modes (which axis filters what)

| Mode | Canon Filter | State Filter | Events Filter | Reveals Filter | Use Case |
|---|---|---|---|---|---|
| **Revealed-narrative** (default) | `valid_from_chapter <= N` | `chapter_number <= N` | `chapter_number <= N` | `chapter_number <= N` | Spoiler-safe writing/analysis |
| **Timeline-consistent** | all canon | `timeline_day <= D` | `timeline_day <= D` | all | Writing a scene at a specific in-world moment |
| **Omniscient author** | all | all | all | all | Planning, editing, full picture |

Timeline-consistent mode falls back to chapter-only filtering for entities that have no `timeline_day`.

## Canon Temperature Model

Story canon has three temperatures. All use the **same retrieval filter** but different update patterns.

### Cold (almost never changes)
- Premise, themes
- Main service: stored as story metadata
- Gateway: `embedding_documents` (is_canonical=true, valid_from_chapter: 0, valid_to_chapter: NULL) + 3 summary chunks

### Warm (updates on major story events)
- Character base profiles, locations, artifacts, world rules, relationship baselines
- Main service: existing entity CRUD (Character, Location, Artifact, Lore, CharacterRelationship)
- Gateway: `embedding_documents` with chapter ranges, versioned when content changes

### Hot (changes every chapter)
- Character emotional state, goals, secrets, clothes, relationship tension, active conflicts
- Main service: **new** CharacterState, RelationshipState entities (per chapter)
- Gateway: `state_snapshots` (JSONB) + `embedding_documents` + summary chunks for semantic search

## Scope

### Part A — Main Service: New Entities ✅

#### New Domain Models ✅
- [x] `CharacterState` — character_id, chapter_id, chapter_number, emotion, goal, secrets (text[]), clothes, physical_state, timeline_day (nullable), timeline_confidence (nullable), metadata (JSONB)
- [x] `RelationshipState` — source_character_id, target_character_id, chapter_id, chapter_number, tension_description, timeline_day (nullable), timeline_confidence (nullable), metadata (JSONB)
- [x] `TimelineEvent` — story_id, chapter_id, chapter_number, timeline_day (nullable), timeline_confidence (nullable), description, entities_involved (JSONB)
- [x] `Reveal` — story_id, chapter_id, chapter_number, fact, entity_ids (JSONB)

#### Migrations (main service) ✅
- [x] 067: Create `character_states` table
- [x] 068: Create `relationship_states` table
- [x] 069: Create `timeline_events` table
- [x] 070: Create `reveals` table

#### Repositories (main service) ✅
- [x] `CharacterStateRepository` — Create, GetByID, GetLatestByCharacter(asOfChapter), ListByChapter, ListByCharacter, Delete
- [x] `RelationshipStateRepository` — Create, GetByID, GetLatestByPair(asOfChapter), ListByChapter, Delete
- [x] `TimelineEventRepository` — Create, GetByID, ListByStory(asOfChapter, limit), ListByChapter, Delete
- [x] `RevealRepository` — Create, GetByID, ListByStory(asOfChapter), ListByChapter, Delete

#### Use Cases (main service) ✅
- [x] CharacterState: create, get, get_latest, list_by_chapter, delete (5 files)
- [x] RelationshipState: create, get, get_latest, list_by_chapter, delete (5 files)
- [x] TimelineEvent: create, get, list_by_story, list_by_chapter, delete (5 files)
- [x] Reveal: create, get, list_by_story, list_by_chapter, delete (5 files)

#### HTTP Endpoints (main service) ✅
- [x] `POST /characters/{character_id}/states` — create character state
- [x] `GET /character-states/{id}` — get by ID
- [x] `GET /characters/{character_id}/states/latest?as_of_chapter=N` — get latest
- [x] `GET /chapters/{chapter_id}/character-states` — list by chapter
- [x] `DELETE /character-states/{id}` — delete
- [x] `POST /relationship-states` — create relationship state
- [x] `GET /relationship-states/{id}` — get by ID
- [x] `GET /relationship-states/latest?source_character_id=X&target_character_id=Y&as_of_chapter=N` — get latest
- [x] `GET /chapters/{chapter_id}/relationship-states` — list by chapter
- [x] `DELETE /relationship-states/{id}` — delete
- [x] `POST /stories/{story_id}/timeline-events` — create timeline event
- [x] `GET /timeline-events/{id}` — get by ID
- [x] `GET /stories/{story_id}/timeline-events?as_of_chapter=N&limit=50` — list by story
- [x] `GET /chapters/{chapter_id}/timeline-events` — list by chapter
- [x] `DELETE /timeline-events/{id}` — delete
- [x] `POST /stories/{story_id}/reveals` — create reveal
- [x] `GET /reveals/{id}` — get by ID
- [x] `GET /stories/{story_id}/reveals?as_of_chapter=N` — list by story
- [x] `GET /chapters/{chapter_id}/reveals` — list by chapter
- [x] `DELETE /reveals/{id}` — delete

#### gRPC Service Extensions ✅
- [x] `CreateCharacterState` / `GetCharacterState` / `GetLatestCharacterState` / `ListCharacterStatesByChapter` / `DeleteCharacterState` RPCs
- [x] `CreateRelationshipState` / `GetRelationshipState` / `ListRelationshipStatesByChapter` / `DeleteRelationshipState` RPCs
- [x] `CreateTimelineEvent` / `GetTimelineEvent` / `ListTimelineEventsByStory` / `ListTimelineEventsByChapter` / `DeleteTimelineEvent` RPCs
- [x] `CreatePlot` / `GetPlot` / `ListPlotsByStory` / `UpdatePlot` / `DeletePlot` / `AddPlotEntity` / `CreatePlotDevelopment` / `ListPlotDevelopments` RPCs
- [x] Proto compilation (protoc + Go plugins installed on WSL2)
- [x] 4 gRPC handlers + server registration + wired in cmd/api-grpc/main.go
- [x] 4 integration test files (14 tests total)

### Part B — Main Service: Parser + Import Orchestrator ✅

#### Parser Module ✅ (8 files, 9 unit tests pass)
- [x] `ParseTabsIndex(content string)` — extract document name, ID, export timestamp, tab list
- [x] `ParseStoryCanon(content string)` — extract all sections: premise, themes, world rules, characters, relationships, locations, artifacts
- [x] Character parsing — name, age, role, description from `### Name \- role \- age` format
- [x] Location parsing — name, sub-locations, symbols
- [x] Artifact parsing — name, description, symbolism
- [x] World rules parsing — numbered rules list
- [x] Themes parsing — bullet-point list
- [x] Relationships parsing — `### A ↔ B` blocks
- [x] `ParseChapterLog(content string)` — character states, relationship states, scenes, secrets, conflicts, timeline, reveals, arc context
- [x] Character state parsing — key-value pairs (`**Emotion**: value`)
- [x] Relationship state parsing — `A ↔ B` header + description
- [x] Scene parsing — `N\. Title` format with descriptions
- [x] Timeline parsing — extract `Day N:` integer + set confidence="explicit"
- [x] `ParseContent(content string)` — split prose by `### ` scene headers
- [x] `ParseNotes(content string)` — parse dated author notes
- [x] `ParseGranularCanonFile(filename, content string)` — route by filename prefix (character_, location_, artifact_, relationship_, premise, themes, world_rules)

#### Import Orchestrator ✅
- [x] `ImportChapterUseCase.Execute(ctx, input)` — full orchestration
- [x] Reads all files from export folder (tabs_index, story_canon, chapter_log, content, notes)
- [x] Tries granular `story_canon/` files first, falls back to combined `01_story_canon.md`
- [x] Deduplicates characters, locations, artifacts by case-insensitive name match
- [x] Creates new entities via use cases (characters, locations, artifacts)
- [x] Creates character states per character per chapter (maps name → ID)
- [x] Creates relationship states (resolves both character names → IDs)
- [x] Creates timeline events from chapter log timeline data
- [x] Creates reveals from chapter log new info revealed
- [x] Returns ImportReport with created/skipped/failed counts
- [x] Individual entity failures logged, don't abort whole import

#### CLI Import Command ✅
- [x] `cmd/import-chapter/main.go` — CLI that reads folder, wires deps, runs orchestrator, prints report
- [x] Flags: `-f` folder, `-t` tenant-id, `-w` world-id, `-s` story-id, `-n` chapter-number, `--chapter-id` (optional)
- [x] Wires repos + use cases same pattern as api-http/main.go
- [x] Sets ingestion queue when `LLM_GATEWAY_ENABLED=true`
- [x] Formatted report output with created/skipped/failed counts

#### Remaining orchestrator work
- [x] Chapter creation (resolve existing or create new)
- [x] Scene creation from parsed chapter log scenes
- [x] Content block creation from parsed prose
- [ ] Notes storage as chapter metadata
- [ ] Scene placement + beat inference via LLM Gateway (see below)

### Part C — LLM Gateway: Retrieval Layer

#### Gateway Retrieval Metadata ✅
- [x] Migration 009: extend `embedding_documents` (canon versioning) + `embedding_chunks` (dual-axis timeline)
- [x] Migration 010: create `concepts`, `concept_links`, `state_snapshots`, `events`, `reveals` tables
- [x] Domain models: Concept, ConceptLink, StateSnapshot, Event, Reveal
- [x] Extended Document + Chunk models with canon/timeline fields
- [x] Repository implementations (5 new repos)
- [x] Integration tests (22 tests across 5 files)

#### Gateway: Scene Placement + Beat Inference (NEW — LLM inference, not ingestion)

The orchestrator creates scenes (from log) and content blocks (from prose) independently. This step uses the LLM to connect them: which prose belongs to which scene, and what beats exist within each scene.

One LLM call returns the full chapter structure:
- **Scene boundaries** — which paragraphs map to which scene
- **Beat segmentation** — sub-scene structural units within each scene
- **Beat typing** — setup, turn, reveal, conflict, climax, resolution, hook, transition (existing domain types)
- **Beat intent extraction** — what each beat accomplishes narratively

```
Input:  scene list (titles + descriptions) + prose text
Output: [{scene: 1, paragraphs: [1-8], beats: [{type: "setup", intent: "...", paragraphs: [1-3]}, ...]}, ...]
```

- [x] Gateway: `InferChapterStructureUseCase` — accepts scene list + prose, calls LLM (Gemini Pro), returns structured placement + beats
- [x] Gateway: prompt template for scene placement + beat inference (`infer_chapter_structure.prompt`)
- [x] Gateway: `POST /api/v1/infer-chapter-structure` endpoint
- [x] Gateway: dedicated `gemini-inference` provider with configurable model, output tokens (16384), timeout (120s)
- [x] Orchestrator: after creating scenes + parsing content, call gateway inference endpoint
- [x] Orchestrator: create beats from inferred beat data (via existing `createBeat` use case)
- [x] Orchestrator: link content blocks to beats AND scenes via ContentAnchors
- [x] Gateway client (`gateway_client.go`): HTTP client for inference endpoint (180s timeout for Gemini Pro)
- [x] Fallback: without gateway, creates basic content blocks linked to chapter only (existing behavior preserved)
- [x] CLI: `--gateway-url` flag (optional, also reads `LLM_GATEWAY_URL` env)

#### Gateway: New Ingesters ✅
- [x] `IngestCharacterStateUseCase` — fetches via gRPC, stores state_snapshot (entity_type="character") + embedding document + chunks
- [x] `IngestRelationshipStateUseCase` — resolves entity names, stores state_snapshot (entity_type="relationship") + embedding document + chunks
- [x] `IngestTimelineEventUseCase` — stores event in retrieval table + embedding document + chunks
- [x] `IngestPlotUseCase` — fetches plot + developments, stores each development as event + embedding document + chunks
- [x] 4 new SourceType constants (character_state, relationship_state, timeline_event, plot)
- [x] 5 new gRPC client methods + 8 DTO types + proto converters + mock stubs
- [x] Worker wiring: DebouncedWorker + cmd/worker/main.go + source type routing

#### Gateway: Canon Retrieval Endpoint ✅
- [x] `DocumentRepository.ListCanonical(storyID, asOfChapter)` — filters by is_canonical + chapter range, uses index from migration 009
- [x] `DocumentRepository.ListByStoryAndSourceType(storyID, sourceType)` — for warm layer queries
- [x] `RetrieveCanonUseCase` — assembles 4 layers (cold+warm canon, hot states, events, reveals) with 3 retrieval modes
- [x] `GET /api/v1/stories/{id}/canon?as_of_chapter=N&as_of_day=D&mode=M&event_limit=K` — HTTP handler wired in cmd/http/main.go
- [x] Mock DocumentRepository updated with new methods

#### Gateway: Remaining retrieval work
- [ ] Extend `ChunkRepository.SearchSimilar()` — add `as_of_chapter`, `as_of_day`, `mode` filters
- [ ] LLM extraction on prose: run content through existing entity-extract pipeline after structured import

### Part D — Obsidian Plugin: UI Integration
- [ ] Command: "Import Chapter from Google Drive Export" — opens folder picker
- [ ] Import config modal: select target story + world, set chapter number, preview folder contents
- [ ] Progress view: reuse existing SSE streaming pattern
- [ ] Import report: summary of created/updated/skipped/failed

### Tests

#### Main Service Tests ✅
- [x] Parser unit tests using real export data as fixtures (9 tests, all pass)
- [x] Story Canon parser: extracts all characters, locations, artifacts, relationships
- [x] Chapter Log parser: extracts character states, relationship states, scenes, timeline, reveals
- [x] Content parser: splits prose by scene headers
- [x] Notes parser: extracts dated entries
- [x] Granular canon file routing: character, relationship, artifact
- [x] CharacterState repo: CRUD + GetLatestByCharacter at chapter N + timeline fields (6 tests)
- [x] RelationshipState repo: CRUD + GetLatestByPair at chapter N (4 tests)
- [x] TimelineEvent repo: CRUD + ListByStory at chapter N + limit (5 tests)
- [x] Reveal repo: CRUD + ListByStory at chapter N + entity_ids roundtrip (4 tests)
- [ ] Import orchestrator: full flow integration test

#### Gateway Tests ✅
- [x] ConceptRepository: CRUD, case-insensitive find, type filtering, unique constraint (6 tests)
- [x] ConceptLinkRepository: CRUD, pair lookup, chapter-range filtering, version closing (5 tests)
- [x] StateSnapshotRepository: chapter-based latest, day-based latest, DISTINCT ON, timeline_confidence (4 tests)
- [x] EventRepository: chapter filtering, day filtering, limit, delete (4 tests)
- [x] RevealRepository: chapter filtering, entity_ids round-trip, delete (3 tests)
- [ ] New ingester tests: CharacterState, RelationshipState, TimelineEvent, Plot
- [ ] Canon retrieval: `as_of_chapter=5` returns correct cold + warm + hot data
- [ ] Canon retrieval: timeline-consistent mode with day filter + fallback

## Dev Log

| # | Date | Event |
|---|------|-------|
| 1 | 2026-03-15 | Task spec written. Export format v2 designed (canon/log split). CHANGES.md created. |
| 2 | 2026-03-15 | Gateway: migrations 009-010 created (canon versioning + 5 narrative tables). Domain models created (Concept, ConceptLink, StateSnapshot, Event, Reveal). Document + Chunk extended. |
| 3 | 2026-03-15 | Gateway: 5 port interfaces + 5 Postgres repository implementations created. |
| 4 | 2026-03-15 | Gateway: 22 integration tests across 5 test files. All compile + pass vet. |
| 5 | 2026-03-15 | Architecture revised: main service is source of truth, gateway is retrieval layer. Parser + import moves to main service. Gateway keeps retrieval metadata + new ingesters. |
| 6 | 2026-03-15 | Main service: migrations 067-070 created (character_states, relationship_states, timeline_events, reveals). |
| 7 | 2026-03-15 | Main service: 4 domain models created (CharacterState, RelationshipState in world package; TimelineEvent, Reveal in story package). |
| 8 | 2026-03-15 | Main service: 4 port interfaces + 4 Postgres repository implementations. |
| 9 | 2026-03-15 | Main service: 19 integration tests across 4 repo test files (character_state: 6, relationship_state: 4, timeline_event: 5, reveal: 4). |
| 10 | 2026-03-15 | Main service: 20 use cases created (5 per entity: create, get, get_latest/list_by_story, list_by_chapter, delete). Audit log entity types added. |
| 11 | 2026-03-15 | Main service: 4 HTTP handlers created (20 endpoints total). Wired into main.go with repos, use cases, ingestion queue, and routes. |
| 12 | 2026-03-15 | Main service: parser module created (8 files, pure Go, no external deps). 9 unit tests pass against real export data fixtures. |
| 13 | 2026-03-15 | Main service: import orchestrator created. Reads export folder, parses all files, deduplicates by name, calls use cases, returns import report. |
| 14 | 2026-03-15 | Main service: CLI `cmd/import-chapter` created. Wires deps, runs orchestrator, prints formatted report. Replaces planned HTTP endpoint. |
| 15 | 2026-03-15 | Orchestrator extended: now creates chapter (resolve or create), scenes (from log), content blocks (from prose). States linked to resolved chapterID. CLI updated with new deps. |
| 16 | 2026-03-15 | Scoped new feature: LLM-powered scene placement + beat inference. Gateway infers which prose maps to which scene, and segments beats within scenes. One LLM call returns full chapter structure. |
| 17 | 2026-03-15 | Gateway: `POST /api/v1/infer-chapter-structure` endpoint built. Use case + prompt + handler + wired. |
| 18 | 2026-03-15 | Gateway: dedicated inference provider `gemini-inference` using Gemini Pro. Configurable via `INFERENCE_MODEL`, `INFERENCE_MAX_OUTPUT_TOKENS` (16384), `INFERENCE_TIMEOUT_SECONDS` (120). RouterModel refactored to functional options. |
| 19 | 2026-03-15 | Orchestrator: inference wired. Calls gateway → creates beats per scene → creates content blocks per beat → links via ContentAnchors to both beat and scene. Falls back to basic content blocks without gateway. |
| 20 | 2026-03-15 | CLI: added `--gateway-url` flag. Creates GatewayClient, beatRepo, createBeatUseCase, contentAnchorRepo when URL is set. "beat" added to report output. |
| 21 | 2026-03-15 | Major refactor: `reveals` replaced with `plots` system (plots + plot_developments + plot_entities + plot_development_entities). `relationship_states` and `timeline_events` now use secondary entity link tables. `character_states` keeps strong FKs (character_id + chapter_id), drops world_id. |
| 22 | 2026-03-15 | Export format updated: "New Information Revealed" → "Plot Threads" section with `### Title — type — action` format and `**Entities**: role: type: name` links. Parser + tests updated. |
| 23 | 2026-03-15 | Proto files created for 4 new services: character_state (5 RPCs), relationship_state (4 RPCs), timeline_event (5 RPCs), plot (8 RPCs). Awaiting protoc compilation. |
| 24 | 2026-03-16 | gRPC toolchain installed on WSL2: protoc v29.3, protoc-gen-go, protoc-gen-go-grpc. All protos compiled (`make proto-gen`). New `docs/` structure created following novelist-app pattern (1xx product, 2xx tech, 3xx how-to). First doc: `300-main.grpc-toolchain-setup.md` with WSL2 known issues section (DNS fix, missing unzip, Postgres auto-start). |
| 25 | 2026-03-16 | Main service: 4 gRPC handlers created (character_state: 5 RPCs, relationship_state: 4 RPCs, timeline_event: 5 RPCs, plot: 8 RPCs). 4 Register methods added to server.go. Wired into cmd/api-grpc/main.go with repos, use cases, handlers, and registration. |
| 26 | 2026-03-16 | Main service: 4 gRPC handler test files created (integration tests). Test infrastructure updated: TestHandlers struct + registration in testing/server.go. Test helper extended with world + character handlers for prerequisite creation. CharacterState tests: 4 (create+get, getLatest with chapter filtering, delete). RelationshipState: 3. TimelineEvent: 3. Plot: 4 (including developments with as_of_chapter filter). |
| 27 | 2026-03-16 | Gateway: 4 new ingesters created (IngestCharacterState, IngestRelationshipState, IngestTimelineEvent, IngestPlot). Each does dual-write: embedding documents + chunks for semantic search, AND retrieval tables (state_snapshots, events) for structured queries. 4 new SourceType constants. 5 new gRPC client methods + 8 DTO types + proto conversion helpers. Mock client updated. Worker wired: DebouncedWorker constructor + processItem routing + storySourceTypes list. cmd/worker/main.go wired with repos + use cases + summary generators. |
| 28 | 2026-03-16 | Gateway: Canon retrieval endpoint built. `GET /api/v1/stories/{id}/canon?as_of_chapter=N&as_of_day=D&mode=M&event_limit=K`. RetrieveCanonUseCase assembles 4 layers: cold+warm (canonical documents with chapter-range filtering), hot states (state_snapshots), events, reveals. Three retrieval modes: revealed_narrative (default, spoiler-safe), timeline_consistent (day filter + fallback), omniscient (everything). DocumentRepository extended with ListCanonical + ListByStoryAndSourceType. Wired into cmd/http/main.go. **Task 1 core pipeline complete.** |
| 29 | 2026-03-18 | Main service: plot_developments refactored — dropped `plot_development_entities` table, added `notes` (text) + `timeline_event_id` (FK → timeline_events). Entity context now flows through linked timeline events. Migration 071. Proto recompiled. All layers updated (domain, repo, use case, handlers, gateway client). |
| 30 | 2026-03-18 | Gateway: reveals system removed — migration 011 drops table, all dead code removed (domain, repo, tests, canon retrieval, MCP). Reveals were replaced by plots in refactor #21 but gateway code was never cleaned up. |
| 31 | 2026-03-18 | Gateway: plot ingestion rewritten with LLM synthesis. Versioned plot documents (warm canon pattern) with Gemini-generated narrative summaries. Individual development documents with detailed info. New `SourceTypePlotDevelopment`. Prompt: `synthesize_plot_state.prompt`. Canon retrieval returns plots for free via `ListCanonical`. |

## Decisions & Rationale

| # | Decision | Chosen | Why |
|---|----------|--------|-----|
| 1 | ~~Bypass main-service~~ → **Main service is source of truth** | Import → main service → publishes → gateway | Retrieval must trace back to canonical source. Gateway already ingests from main service via gRPC. Building a shadow copy creates divergence. |
| 2 | Two ingestion paths (structured + LLM) | Structured for canon/log, LLM for prose | Canon/log are already structured — no LLM waste. Prose needs discovery. |
| 3 | Structured ingestion runs before LLM extraction | Sequential | Extraction needs existing entities to match against |
| 4 | Three-temperature canon model | Cold / warm / hot | Different update frequencies, same retrieval filter. No full snapshot duplication, no diff replay. |
| 5 | Warm canon: chapter-ranged versions | `valid_from_chapter` / `valid_to_chapter` on gateway embedding_documents | Chapter number is always present (primary axis). Simple range query, no reconstruction. |
| 6 | Warm versioning by content comparison | Diff against existing canon on ingestion | Auto-detects when canon entry actually changed vs just re-exported unchanged |
| 7 | Hot state as snapshots not diffs | Full state per entity per chapter | Clean retrieval — `DISTINCT ON + ORDER BY DESC`, no replay chain |
| 8 | Hot state ALSO gets embedding_documents + summaries | Dual storage: JSONB for structured queries, chunks for semantic search | "How is Haruka feeling?" needs to find the right state via vector search |
| 9 | Reuse `GenerateSummaryUseCase` for all entity types | Same prompt pipeline, 3 summaries per entity | Summary chunks massively improve semantic search recall — proven pattern |
| 10 | Reuse existing entity-extract orchestrator for prose | Phases 1-7 already do entity discovery + matching + relationship finding | No new extraction code needed — just feed prose as input text |
| 11 | Granular files preferred over combined | Parse `story_canon/` if present, fall back to `01_story_canon.md` | Granular files are cleaner to parse, less ambiguous section boundaries |
| 12 | Export format v2: canon/log split | Story Canon (identity) + Chapter Log (state) | Parser never guesses where identity ends and state begins — mechanical split |
| 13 | "Key Objects" → "Artifacts" in export | Matches gateway entity type `artefact` | One vocabulary across export, gateway, and retrieval |
| 14 | New entities in main service (CharacterState, RelationshipState, TimelineEvent, Reveal) | Main service owns them, gateway ingests them | These are structured data with canonical identity — they belong in source of truth, not retrieval layer |
| 15 | Gateway keeps concepts + concept_links tables | Retrieval-optimized, not main service shadow | Gateway needs entity graph for concept traversal + dedup — populated by gateway ingesters, not imported from main service |
| 16 | Parser lives in main service | Not in gateway or plugin | Main service is where entities are created — parser feeds directly into CRUD |
| 17 | Dual timeline: chapter_number (primary) + timeline_day (secondary) | Both fields on every record | Stories are non-linear — prologue (ch 0) may show Day 0 while ch 1 goes back to Day -63. Need both axes for correct retrieval. |
| 18 | chapter_number NOT NULL, timeline_day nullable | Primary axis always present, secondary optional | Not all stories have explicit in-world timelines. Chapter order is always known at ingestion. |
| 19 | timeline_confidence field | 'explicit' / 'inferred' / NULL | Author-stated timeline is trusted. LLM-inferred timeline (future feature) needs a confidence flag. Schema supports it from day one. |
| 20 | Timeline-consistent mode falls back to chapter | `WHERE (timeline_day <= D OR (timeline_day IS NULL AND chapter_number <= N))` | Entities without timeline_day shouldn't disappear from results — use chapter as fallback |
| 21 | Canon retrieval as single assembled endpoint | One GET returns all 4 layers (canon + states + events + reveals) | MCP/LLM consumers need full context in one call, not N+1 queries. Layers are cheap to assemble server-side. |
| 22 | Route uses `/api/v1/stories/{id}/canon` (path param) | Consistent with REST, story_id in path not query | Natural hierarchy: stories → canon. Query params for filters (mode, chapter, day). |

## Remaining Work (prioritized)

| # | What | Where | Depends on | Status |
|---|---|---|---|---|
| ~~1~~ | ~~Wire inference into orchestrator~~ | ~~Main service~~ | — | ✅ Done |
| ~~2~~ | ~~gRPC service definitions for new entities~~ | ~~Main service (proto + handlers)~~ | — | ✅ Done — protoc installed, 4 protos compiled, 4 gRPC handlers (22 RPCs), 4 test files (14 tests), wired in server.go + api-grpc/main.go |
| ~~3~~ | ~~Gateway ingesters for new entities~~ | ~~Gateway~~ | ~~#2~~ | ✅ Done — 4 ingesters with dual-write (embeddings + retrieval tables), 5 gRPC client methods, worker wiring, source type routing |
| ~~4~~ | ~~Canon retrieval endpoint~~ | ~~Gateway~~ | ~~#3~~ | ✅ Done — `GET /api/v1/stories/{id}/canon?as_of_chapter=N&as_of_day=D&mode=M`. Assembles cold+warm+hot across 4 layers. 3 retrieval modes. DocumentRepository extended. |
| 5 | End-to-end test | Both | ~~#4~~ | Run CLI against real data with DB running. Verify full flow: import → entities → Redis publish → gateway ingest → canon retrieval. |
| 6 | Task 2: MCP server | New service | ~~#4~~ | Thin wrapper over gateway retrieval endpoints. Separate task — **now unblocked**. |

## Difficulties & Solutions

| # | Problem | Solution | Time |
|---|---------|----------|------|

## Patterns Discovered

## Acceptance Criteria

- [ ] `go run ./cmd/import-chapter -f data/gcloud-exported-chapter -t X -w X -s X -n 0` → all characters created in main service (Haruka, Shouta, M-sama, Tetsuo, Sayaka)
- [ ] All locations created in main service (Rouge Seasons Hotel, Akiba Duplex)
- [ ] All artifacts created in main service (Mirrors, Phone messages, Vanilla-rose diffuser, Crimson ribbon, etc.)
- [ ] Character states created per character per chapter with dual timeline fields
- [ ] Relationship states created per relationship per chapter
- [ ] Timeline events created with chapter_number + timeline_day (explicit) + timeline_confidence
- [ ] Reveals created with chapter_number
- [ ] Main service publishes all entities to gateway via existing gRPC + Redis queue
- [ ] Gateway embeds + summarizes all received entities (3 summary chunks each)
- [ ] Gateway stores canon documents with chapter ranges (cold/warm versioning)
- [ ] Gateway stores state_snapshots from ingested character/relationship states
- [ ] `GET /api/v1/stories/{id}/canon?as_of_chapter=5` returns assembled canon with correct cold + warm + hot data
- [ ] `GET /api/v1/stories/{id}/canon?as_of_chapter=5&as_of_day=-40&mode=timeline_consistent` returns state true at Day -40
- [ ] `GET /api/v1/stories/{id}/canon?mode=omniscient` returns everything
- [ ] Prose from Content tab processed through entity-extract pipeline, matches against imported entities
- [ ] Re-importing same chapter: unchanged entities skipped, states updated
- [ ] Obsidian plugin can trigger import and display results
