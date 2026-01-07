Ingestion Reindex + Pipeline Architecture (Ideation)

Goal
- Allow reingestion for a tenant, optionally filtered by entity type.
- Support multiple pipelines (base ingestion + optional enrichers).
- Keep eventual consistency and avoid blocking user-facing flows.

Core Concept
- A single Reingest Service is the source of truth for reindex requests.
- Two adapters call the same Reingest Service:
  - CLI for ops
  - Admin HTTP endpoint for product
- Pipelines are independent workers consuming dedicated queues.
- Base ingestion emits events that trigger downstream pipelines.

Reingest Service (Core Use Case)
Inputs
- tenant_id (required)
- types (optional list of entity types)
- dry_run (optional)
- batch_size (optional)

Behavior
- Resolve entity IDs from main-service for the tenant, filtered by types.
- Enqueue each entity to the base ingestion queue.
- Return counts and requested filters.

Entity Types (initial)
- story, chapter, scene, beat, content_block
- world, character, location, event, artifact, faction, lore

Adapters
CLI
- Command: reingest --tenant <uuid> --types story,character --batch 100
- Calls Reingest Service with filters

Admin HTTP
- POST /admin/tenants/{tenant_id}/reingest
- Body:
  {
    "types": ["story", "character"],
    "batch_size": 100,
    "dry_run": false
  }

Queue and Pipeline Model
Queues
- ingestion_base (base ingest + embed)
- pipeline_dialog_mood
- pipeline_character_traits

Rules
- Base ingestion is the only source that writes embeddings.
- Enrichment pipelines consume events or queue items derived from base ingestion.
- Each pipeline is isolated for scaling and failure handling.

Event Contract (Proposed)
Event name: entity.ingested
Payload:
{
  "tenant_id": "uuid",
  "source_type": "story",
  "source_id": "uuid",
  "version": 1,
  "occurred_at": "2026-01-06T00:00:00Z",
  "metadata": {
    "world_id": "uuid",
    "content_kind": "draft"
  }
}

Additional events (optional)
- entity.changed (when upstream changes detected)
- reingest.requested (for auditing and async ops)

Pipeline Triggering
- Base worker emits entity.ingested
- Orchestrator (or light router) enqueues to downstream pipelines based on:
  - source_type
  - metadata conditions
  - feature flags

Pipeline Status Tracking (Optional)
Table: pipeline_runs
- id (uuid)
- tenant_id (uuid)
- source_type (text)
- source_id (uuid)
- pipeline (text)
- status (pending, running, success, failed)
- started_at, finished_at
- attempt (int)
- error_message (text)

This enables:
- retries with backoff
- observability for partial failures

LangGraph Integration (Future)
Pattern
- Each pipeline can be a LangGraph graph.
- Worker passes input (entity + metadata) to graph.
- Graph outputs:
  - enriched data to store
  - optional new events

Benefits
- multi-step workflows
- conditional branching (eg. only classify dialog content)
- checkpointing and retries

Consistency Model
- Eventual consistency only.
- User-facing reads should not depend on pipeline completion.

Open Questions
- Do we want a dedicated event bus or reuse Redis streams?
- Should base ingestion enqueue downstream pipelines directly or via an event router?
- Where to store pipeline metadata for traceability (DB vs logs)?
