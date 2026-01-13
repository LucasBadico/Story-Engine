# Extract Pipeline Domain Split (Entities vs Relations)

Goal: separate extraction into two domain-specific pipelines with a lightweight orchestrator.
This avoids numeric phases and keeps each pipeline extensible and testable.

## Core Idea

- `EntitiesExtractor`: responsible for all entity discovery and resolution.
- `RelationsExtractor`: responsible for relation discovery/normalization/matching.
- `ExtractOrchestrator`: runs entities first, then relations (if enabled), and streams results.

## Operation Names (No Numeric Phases)

These names are used for:
- SSE progress events
- logs
- internal step identifiers

Entities:
- `entities.detect`
- `entities.candidates`
- `entities.resolve`
- `entities.result`

Relations:
- `relations.discover`
- `relations.normalize`
- `relations.match`
- `relations.result`

Optional substage for richer progress:
- `{ kind: "relations", stage: "discover", substage: "batch" }`

## Module Structure (LLM Gateway)

Current (after refactor start):
- `llm-gateway-service/internal/application/extract/`
  - `orchestrator.go`
  - `payload.go`
  - `events/`
    - `events.go`
  - `entities/`
    - `type_router.go`
    - `entity_extractor.go`
    - `entrypoint.go`
    - `match.go`
    - `payload.go`
    - `text_split.go`
    - `prompts/`
      - `phase1_entity_type_router.prompt`
      - `phase2_*_extractor.prompt`
      - `phase3_entity_match*.prompt`
  - `relations/`
    - `discovery.go`
    - `normalize.go`
    - `match.go`
    - `flow.go`
    - `payload.go`
    - `prompts/`
      - `phase5_relation_discovery.prompt`
      - `phase6_custom_relation_summary.prompt`

Notes:
- Entities and relations prompts live in their own domain folders.
- Orchestrator owns request/response assembly (stream + non-stream).
- Shared helpers can live in `extract/common/`.

## Suggested File Renames / Moves

These rename examples assume earlier files lived under `entity_extraction/`.

Entities (examples):
- `phase1_entity_type_detection.*` -> `entities/detect.*`
- `phase2_entity_candidates.*` -> `entities/candidates.*`
- `phase3_entity_matching.*` -> `entities/resolve.*`
- `phase4_entity_result.*` -> `entities/result.*`

Relations (examples):
- `phase5_relation_discovery.*` -> `relations/discover.*`
- `phase6_relation_normalize.*` -> `relations/normalize.*`
- `phase7_relation_match.*` -> `relations/match.*`
- `phase8_relation_result.*` -> `relations/result.*`

Orchestrator:
- `extractor_and_relations_extractor.go` -> `extract/orchestrator.go`

Prompts:
- `phase5_relation_discovery.prompt` -> `relations/prompts/phase5_relation_discovery.prompt`
- `phase6_custom_relation_summary.prompt` -> `relations/prompts/phase6_custom_relation_summary.prompt`

SSE Events:
- `entity_extract_sse_events.md` should map to the new names.

## Public API Behavior

No change in public shape required:
- stream: `result_entities`, `result_relations`
- non-stream: `{ entities, relations }`

Only the internal progress identifiers and filenames change.

## High-Level Interfaces (Pseudo)

```go
type ExtractOrchestrator struct {
  entities EntitiesExtractor
  relations RelationsExtractor
}

type ExtractRequest struct {
  TenantID uuid.UUID
  WorldID uuid.UUID
  Text string
  IncludeRelations bool
  // ...other tuning fields
}

type ExtractResult struct {
  Payload ExtractPayload
}

type EntitiesExtractor interface {
  Run(ctx, req) (EntitiesResult, error)
}

type RelationsExtractor interface {
  Run(ctx, req, entities EntitiesResult) (RelationsResult, error)
}
```

## Benefits

- No phase renumbering when inserting new steps.
- Clear ownership of prompts, tests, and models.
- Relations can evolve without touching entity steps.
- Easier to add additional domains later (e.g., `citations`, `events`).
