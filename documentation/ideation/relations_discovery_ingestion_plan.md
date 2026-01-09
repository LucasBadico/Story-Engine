# Relations Discovery + Ingestion Plan

> **⚠️ STATUS (Updated 2026-01-08)**: Code does NOT compile. All 3 entrypoints (api-http, api-grpc, api-offline) are broken.
> Old reference types were removed but interfaces/handlers still reference them.
> **Priority**: Complete Phase 0 (Emergency Fix) before any other work.

## Goal
Discover and store relations between entities in a world, make them searchable, and allow UI to reconcile temporary entities with final IDs.

## Why
- Relation context improves search and downstream LLM reasoning.
- We need a deterministic way to store relations without always calling LLM.
- UI must handle relations that involve entities not yet created.

## Concepts

### Relation Record
A first-class entity that links two entities:
- `id`
- `tenant_id` (required for multi-tenant filtering)
- `world_id`
- `source_entity_id` (nullable when temporary)
- `target_entity_id` (nullable when temporary)
- `source_ref` (string: real ID or temp ID)
- `target_ref` (string: real ID or temp ID)
- `relation_type` (string)
- `confidence` (float)
- `evidence_spans` (offsets or excerpts)
- `status` (draft | confirmed | orphan)
- `created_by_user_id` (nullable - for audit)
- `created_at`
- `updated_at`

### Unified Relations Table (single source of truth)
A single, polymorphic table can replace `*_references` and `character_relationships`:

Suggested columns:
- `id`
- `tenant_id`
- `world_id`
- `source_type` (e.g. character, event, lore, content_block)
- `source_id` (nullable if temp)
- `source_ref` (string: real ID or `tmp:<uuid>`)
- `target_type`
- `target_id` (nullable if temp)
- `target_ref` (string: real ID or `tmp:<uuid>`)
- `relation_type` (e.g. member_of, role, significance, mentions)
- `context_type` (optional: event, lore, faction, scene, content_block)
- `context_id` (optional)
- `attributes` (JSONB: description, notes, bidirectional, role, etc)
- `confidence` (float)
- `evidence_spans` (JSONB)
- `summary`
- `status` (draft | confirmed | orphan)
- `created_by_user_id` (nullable)
- `created_at`
- `updated_at`

Rationale:
- covers entity↔entity, content↔entity, and entity↔context relations
- supports temp IDs and discovery pipeline
- avoids many narrow tables while keeping semantics in `relation_type` and `attributes`

### Evidence Spans Schema
```json
{
  "spans": [
    { "start": 0, "end": 150, "text": "Excerpt from source content..." }
  ]
}
```

### Attributes Schema by Relation Type
| relation_type | attributes schema |
|---------------|-------------------|
| `parent_of`, `child_of`, `sibling_of`, `spouse_of` | `{ "biological": bool }` |
| `ally_of`, `enemy_of` | `{ "strength": float, "since": "date" }` |
| `member_of`, `leader_of` | `{ "role": "string", "since": "date" }` |
| `role` (event→character) | `{ "role_name": "protagonist" }` |
| `significance` (event→location) | `{ "significance": "primary" }` |
| `mentions`, `references` | `{ "notes": "string" }` |

### Bidirectional Relations (Auto-Mirror)
When creating a relation, the system **automatically creates the inverse relation**.

**Inverse Type Mapping:**
| Original | Inverse |
|----------|---------|
| `parent_of` | `child_of` |
| `child_of` | `parent_of` |
| `sibling_of` | `sibling_of` (symmetric) |
| `spouse_of` | `spouse_of` (symmetric) |
| `ally_of` | `ally_of` (symmetric) |
| `enemy_of` | `enemy_of` (symmetric) |
| `member_of` | `has_member` |
| `has_member` | `member_of` |
| `leader_of` | `led_by` |
| `led_by` | `leader_of` |
| `located_in` | `contains` |
| `contains` | `located_in` |
| `owns` | `owned_by` |
| `owned_by` | `owns` |
| `mentor_of` | `mentored_by` |
| `mentored_by` | `mentor_of` |

**Implementation:**
```go
var inverseRelations = map[string]string{
    "parent_of":    "child_of",
    "child_of":     "parent_of",
    "sibling_of":   "sibling_of",
    "spouse_of":    "spouse_of",
    "ally_of":      "ally_of",
    "enemy_of":     "enemy_of",
    "member_of":    "has_member",
    "has_member":   "member_of",
    "leader_of":    "led_by",
    "led_by":       "leader_of",
    "located_in":   "contains",
    "contains":     "located_in",
    "owns":         "owned_by",
    "owned_by":     "owns",
    "mentor_of":    "mentored_by",
    "mentored_by":  "mentor_of",
}

func GetInverseRelationType(relationType string) string {
    if inverse, ok := inverseRelations[relationType]; ok {
        return inverse
    }
    return relationType // fallback: symmetric
}
```

**Example:**
```
Input:  CreateRelation(A, "child_of", B)
Result: 
  - Relation 1: A → child_of → B
  - Relation 2: B → parent_of → A (auto-created)
```

**Linking mirror relations:**
Both relations should have a `mirror_id` field pointing to each other:
```sql
ALTER TABLE entity_relations ADD COLUMN mirror_id UUID REFERENCES entity_relations(id);
```

This allows:
- Deleting one → automatically delete the mirror
- Updating one → prompt to update the mirror
- Querying only "primary" relations (`WHERE mirror_id IS NULL OR id < mirror_id`)
### World-Level Relation Discovery
We should support a batch job that scans the entire world content to propose relations:
- Input: all entity descriptions + content blocks for the world.
- Output: relation candidates across the whole world (not just a single request).
- Usage: seed a relation catalog for search and downstream reasoning.

### Summary (no LLM)
We store a synthetic summary for each relation:
```
<source_name> <relation_type> <target_name>
```
Optionally append evidence snippets. This becomes an embedding entry for retrieval.

### Temporary IDs
When entity creation is pending:
- create a temporary ID: `tmp:<uuid>`
- relation refs can point to these IDs
- when entity is created, resolve temp IDs → real IDs
- `status` stays `draft` while any temp ref exists
- temp IDs are request-scoped, so we can safely reconcile after the request completes

## Pipeline

### Phase 2 (Entity Extraction)
- Extract entity candidates with summaries.
- Assign `tmp:<uuid>` for entities not yet persisted.

### Phase 2.5 (Relation Candidate Generation)
- Use heuristics to propose relations:
  - co-occurrence in sentence/paragraph
  - pattern matching: “X is Y's brother”, “X belongs to Y”, etc.
- Output candidate list:
  - `source_ref`, `target_ref`, `relation_type`, `confidence`, `evidence`

### Phase 3 (Relation Validation)
- Optional LLM validation:
  - Given full text + candidate relation
  - Confirm relation type or reject
  - If LLM is used, prefer validation/classification over free-form creation

### World Relation Discovery Job (batch - llm-gateway-service)
**Nota**: Este job será implementado no `llm-gateway-service`, não no `main-service`. O worker:
- Busca entidades via gRPC do `main-service`
- Aplica heurísticas para detectar relações
- Opcionalmente valida com LLM
- Cria relações descobertas via gRPC no `main-service` (que então gera summary e enfileira embedding)

Fluxo:
- Worker no `llm-gateway-service` escuta eventos ou endpoint interno
- Busca entidades do world via gRPC (`main-service`)
- Gera candidatos de relações com heurísticas
- Opcionalmente valida com LLM
- Chama `main-service` via gRPC para criar relações confirmadas
- `main-service` cria a relação e enfileira para embedding (como relações manuais)

### Persistence
- Insert relation record into DB.
- Generate synthetic summary and enqueue ingestion as a new embedding chunk (type `relations`).

## Ingestion Strategy
- We ingest relation summaries as additional chunks.
- These use the same metadata and world_id as entities.
- Relation chunks should be included in search filters.
- Summaries are generated without LLM to keep ingestion deterministic.

## Migration Notes (unify all relations)
Existing tables can map into the unified table:
- `character_relationships` → `source_type=character`, `target_type=character`, `relation_type=relationship_type`, `attributes.description/bidirectional`
- `event_references` → `source_type=event`, `target_type=entity_type`, `relation_type=relationship_type`, `attributes.notes`
- `lore_references` → `source_type=lore`, `target_type=entity_type`, `relation_type=relationship_type`, `attributes.notes`
- `faction_references` → `source_type=faction`, `target_type=entity_type`, `relation_type=role`, `attributes.notes`
- `artifact_references` → `source_type=artifact`, `target_type=entity_type`, `relation_type=references`
- `scene_references` → `source_type=scene`, `target_type=entity_type`, `relation_type=mentions`
- `content_block_references` → **NÃO MIGRAR** - será renomeado para `content_anchors` (ver Phase 11)

### Semantic Distinction: Relations vs Anchors
| Aspect | `entity_relations` | `content_anchors` (ex-content_block_references) |
|--------|-------------------|------------------------------------------------|
| Purpose | Semantic relationships between entities | Text mentions within prose content |
| Examples | "John is father of Mary", "Sword belongs to Hero" | "In paragraph 3, mention of 'John'" |
| Discovery | LLM/heuristic detection | User-created or text parsing |
| Bidirectional | Yes (auto-mirror) | No (anchor is one-way) |
| Embedding | Yes (relation summaries) | No |

## Unified Schema (SQL)

### Uniqueness Validation (Application Level)
**Note**: Uniqueness is enforced at the **application level**, not database level, for flexibility.

The uniqueness check validates the **combination** of these fields:
- `tenant_id` + `source_ref` + `target_ref` + `relation_type` + `context_type` + `context_id`

This means:
- ✅ A → `child_of` → B (allowed)
- ✅ B → `parent_of` → A (allowed - different direction, auto-mirrored)
- ❌ A → `child_of` → B (duplicate - rejected by application)
- ✅ A → `ally_of` → B (allowed - different relation_type)

**Implementation**: The `CreateRelationUseCase` checks for duplicates before creating a new relation.

### Postgres
```sql
CREATE TABLE IF NOT EXISTS entity_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL,
    source_id UUID,
    source_ref VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id UUID,
    target_ref VARCHAR(100) NOT NULL,
    relation_type VARCHAR(100) NOT NULL,
    context_type VARCHAR(50),
    context_id UUID,
    attributes JSONB NOT NULL DEFAULT '{}',
    confidence REAL,
    evidence_spans JSONB,
    summary TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    mirror_id UUID,  -- Points to the auto-created inverse relation
    created_by_user_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Constraints
    CONSTRAINT entity_relations_ref_unique UNIQUE (tenant_id, source_ref, target_ref, relation_type, COALESCE(context_type, ''), COALESCE(context_id, '00000000-0000-0000-0000-000000000000'::UUID)),
    CONSTRAINT entity_relations_status_check CHECK (status IN ('draft', 'confirmed', 'orphan')),
    CONSTRAINT entity_relations_mirror_fk FOREIGN KEY (mirror_id) REFERENCES entity_relations(id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_entity_relations_tenant_id ON entity_relations(tenant_id);
CREATE INDEX idx_entity_relations_world_id ON entity_relations(world_id);
CREATE INDEX idx_entity_relations_source ON entity_relations(source_type, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX idx_entity_relations_target ON entity_relations(target_type, target_id) WHERE target_id IS NOT NULL;
CREATE INDEX idx_entity_relations_relation_type ON entity_relations(relation_type);
CREATE INDEX idx_entity_relations_status ON entity_relations(status);
CREATE INDEX idx_entity_relations_context ON entity_relations(context_type, context_id) WHERE context_type IS NOT NULL;
CREATE INDEX idx_entity_relations_source_ref ON entity_relations(source_ref);
CREATE INDEX idx_entity_relations_target_ref ON entity_relations(target_ref);
CREATE INDEX idx_entity_relations_temp_refs ON entity_relations(status) WHERE source_ref LIKE 'tmp:%' OR target_ref LIKE 'tmp:%';
```

### SQLite
```sql
CREATE TABLE IF NOT EXISTS entity_relations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    source_type TEXT NOT NULL,
    source_id TEXT,
    source_ref TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT,
    target_ref TEXT NOT NULL,
    relation_type TEXT NOT NULL,
    context_type TEXT,
    context_id TEXT,
    attributes TEXT NOT NULL DEFAULT '{}',
    confidence REAL,
    evidence_spans TEXT,
    summary TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    mirror_id TEXT,  -- Points to the auto-created inverse relation
    created_by_user_id TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    -- Foreign keys
    CONSTRAINT entity_relations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT entity_relations_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT entity_relations_mirror_fk FOREIGN KEY (mirror_id) REFERENCES entity_relations(id) ON DELETE SET NULL,
    -- Constraints
    CONSTRAINT entity_relations_status_check CHECK (status IN ('draft', 'confirmed', 'orphan'))
);

-- Note: Uniqueness validation is handled at application level, not database level

-- Indexes
CREATE INDEX IF NOT EXISTS idx_entity_relations_tenant_id ON entity_relations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_world_id ON entity_relations(world_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_source ON entity_relations(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_target ON entity_relations(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_relation_type ON entity_relations(relation_type);
CREATE INDEX IF NOT EXISTS idx_entity_relations_status ON entity_relations(status);
CREATE INDEX IF NOT EXISTS idx_entity_relations_context ON entity_relations(context_type, context_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_source_ref ON entity_relations(source_ref);
CREATE INDEX IF NOT EXISTS idx_entity_relations_target_ref ON entity_relations(target_ref);
```

## Indexes (minimum)
- `tenant_id` (multi-tenant filtering)
- `world_id`
- `source_type, source_id` (partial index where not null)
- `target_type, target_id` (partial index where not null)
- `source_ref` (for temp ID lookups)
- `target_ref` (for temp ID lookups)
- `relation_type`
- `status`
- `context_type, context_id` (partial index where not null)
- temp refs index (for cleanup jobs)

## Migration Plan (high level)
> **Note:** This is a greenfield system - no backfill needed, just replace tables.

1) Create `entity_relations` table in both Postgres and SQLite.
2) Drop old tables: `character_relationships`, `*_references`.
3) Remove old repository implementations.
4) Update services to use new `EntityRelationRepository`.


## Endpoint/Query Updates (impact map)
When unifying relations, endpoints that read/write `*_references` and `character_relationships`
should query `entity_relations` using a consistent set of filters.

### Common Query Patterns
- **List by source**: `WHERE tenant_id = ? AND source_type = ? AND source_id = ? AND status != 'orphan'`
- **List by target**: `WHERE tenant_id = ? AND target_type = ? AND target_id = ? AND status != 'orphan'`
- **List by context**: `WHERE tenant_id = ? AND context_type = ? AND context_id = ? AND status != 'orphan'`
- **Delete by source+target**: `WHERE tenant_id = ? AND source_type = ? AND source_id = ? AND target_type = ? AND target_id = ?`

### Pagination & Filtering (all LIST endpoints)
Query params:
- `?limit=20&cursor=<opaque_string>` - Cursor-based pagination (default limit=50, max=100)
- `?relation_type=ally_of` - Filter by relation type
- `?status=confirmed` - Filter by status (default excludes 'orphan')
- `?confidence_min=0.7` - Minimum confidence threshold
- `?order_by=created_at&order_dir=desc` - Sorting

**Cursor format:** Base64-encoded `{id}:{created_at}` for stable pagination.

Response includes:
```json
{
  "data": [...],
  "pagination": {
    "limit": 20,
    "next_cursor": "YWJjMTIzOjIwMjQtMDEtMTVUMTI6MDA6MDBa",
    "has_more": true
  }
}
```

**Why cursor over offset:**
- Stable results when data changes between pages
- Better performance on large datasets (no `OFFSET N` scan)
- Prevents skipping/duplicating items on insert/delete

### Endpoint Mapping (HTTP + gRPC)
- **Characters relationships**
  - `GET /api/v1/characters/{id}/relationships`
    - list `entity_relations` where `(source_type=character AND source_id=id) OR (target_type=character AND target_id=id)`
    - Note: For bidirectional relations, both directions are stored in a single record. The query returns relations where the character is either source OR target.
  - `POST /api/v1/characters/{id}/relationships`
    - insert `source_type=character`, `source_id=id`, `target_type=character`, `target_id=...`, `relation_type=...`
    - For bidirectional: store `attributes.bidirectional=true`
  - `PUT /api/v1/character-relationships/{id}` / `DELETE /api/v1/character-relationships/{id}`
    - update/delete by `id`
- **Events references**
  - `GET /api/v1/events/{id}/references`
    - list `source_type=event`, `source_id=id`
  - `POST /api/v1/events/{id}/references`
    - insert `source_type=event`, `source_id=id`, `target_type=<entity_type>`, `target_id=<entity_id>`, `relation_type=<relationship_type>`
  - `DELETE /api/v1/events/{id}/references/{entity_type}/{entity_id}`
    - delete by `source_type=event`, `source_id=id`, `target_type=entity_type`, `target_id=entity_id`
- **Lore references**
  - `GET /api/v1/lores/{id}/references`
    - list `source_type=lore`, `source_id=id`
  - `POST /api/v1/lores/{id}/references`
    - insert `source_type=lore`, `source_id=id`, `target_type=<entity_type>`, `target_id=<entity_id>`
  - `DELETE /api/v1/lores/{id}/references/{entity_type}/{entity_id}`
    - delete by `source_type=lore`, `source_id=id`, `target_type=entity_type`, `target_id=entity_id`
- **Faction references**
  - `GET /api/v1/factions/{id}/references`
    - list `source_type=faction`, `source_id=id`
  - `POST /api/v1/factions/{id}/references`
    - insert `source_type=faction`, `source_id=id`, `target_type=<entity_type>`, `target_id=<entity_id>`
  - `DELETE /api/v1/factions/{id}/references/{entity_type}/{entity_id}`
    - delete by `source_type=faction`, `source_id=id`, `target_type=entity_type`, `target_id=entity_id`
- **Artifact references**
  - `GET /api/v1/artifacts/{id}/references`
    - list `source_type=artifact`, `source_id=id`
  - `POST /api/v1/artifacts/{id}/references`
    - insert `source_type=artifact`, `source_id=id`, `target_type=<entity_type>`, `target_id=<entity_id>`
  - `DELETE /api/v1/artifacts/{id}/references/{entity_type}/{entity_id}`
    - delete by `source_type=artifact`, `source_id=id`, `target_type=entity_type`, `target_id=entity_id`
- **Scene references**
  - `GET /api/v1/scenes/{id}/references`
    - list `source_type=scene`, `source_id=id`
  - `POST /api/v1/scenes/{id}/references`
    - insert `source_type=scene`, `source_id=id`, `target_type=<entity_type>`, `target_id=<entity_id>`
  - `DELETE /api/v1/scenes/{id}/references/{entity_type}/{entity_id}`
    - delete by `source_type=scene`, `source_id=id`, `target_type=entity_type`, `target_id=entity_id`
- **Content block references** → Renomear para **`content_anchors`** (ver Phase 11)
  - Será renomeado para `content_anchors` para clarificar que são âncoras de texto
  - NÃO são relações semânticas entre entidades - são menções no prose content
  - Endpoints mantém compatibilidade: `/api/v1/content-blocks/{id}/references` (deprecated) → `/api/v1/content-blocks/{id}/anchors` (new)

### New Generic Endpoints
- `GET /api/v1/worlds/{world_id}/relations` - List all relations in a world
- `GET /api/v1/relations/{id}` - Get relation by ID
- `POST /api/v1/relations` - Create generic relation
- `PUT /api/v1/relations/{id}` - Update relation
- `DELETE /api/v1/relations/{id}` - Delete relation
- `POST /api/v1/relations/resolve-temp` - Resolve temporary IDs batch
- ~~`POST /api/v1/worlds/{world_id}/relations/discover`~~ - **Removido**: Discovery pipeline será implementado no `llm-gateway-service`
- ~~`GET /api/v1/worlds/{world_id}/relations/discover/{job_id}`~~ - **Removido**: Discovery pipeline será implementado no `llm-gateway-service`

## UI Behavior
- UI receives:
  - entities found (with temp IDs)
  - relations found (with temp refs)
- After entity creation:
  - backend returns a map `tmp_id -> entity_id`
  - UI updates relations in memory
  - backend updates relation record, flips status to `confirmed`
- UI should block relation creation while any temp ref exists.

## Cascade Behavior
When an entity is deleted, what happens to its relations?

**Recommended approach:**
1. If `source_id` entity is deleted → relation status changes to `orphan`
2. If `target_id` entity is deleted → relation status changes to `orphan`
3. Orphan relations are:
   - Excluded from search results by default
   - Cleaned up by a scheduled job after N days (configurable, default 30)
   - Can be manually restored if entity is recreated

**Alternative (simpler):** Hard delete cascade - but loses relation history.

**Implementation:**
- Add trigger/hook on entity delete to update related `entity_relations` status
- Or use scheduled job to detect and mark orphans

## Temporary ID Cleanup
Temp IDs (`tmp:<uuid>`) need cleanup if the request fails:

1. **TTL approach:** Relations with temp refs older than 24h are deleted by cleanup job.
2. **Request-scoped:** Store request_id in a separate column, cleanup on request failure.

**Cleanup Job (recommended):**
```sql
DELETE FROM entity_relations 
WHERE status = 'draft' 
  AND (source_ref LIKE 'tmp:%' OR target_ref LIKE 'tmp:%')
  AND created_at < NOW() - INTERVAL '24 hours';
```

## Discovery Job Table (llm-gateway-service)
**Nota**: Esta tabela será criada no `llm-gateway-service`, não no `main-service`. O `main-service` apenas fornece gRPC endpoints para o worker buscar entidades e criar relações.

For tracking world-level batch discovery jobs:

### Postgres
```sql
CREATE TABLE IF NOT EXISTS relation_discovery_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    progress_pct INT NOT NULL DEFAULT 0,
    total_entities INT,
    processed_entities INT DEFAULT 0,
    relations_found INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT relation_discovery_jobs_status_check CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'))
);

CREATE INDEX idx_relation_discovery_jobs_world ON relation_discovery_jobs(world_id);
CREATE INDEX idx_relation_discovery_jobs_status ON relation_discovery_jobs(status);
```

## Rate Limiting (LLM Calls)
For Phase 3 (LLM validation):

- **Default limit:** 10 requests/minute per tenant
- **Backoff strategy:** Exponential backoff with jitter (1s, 2s, 4s, 8s, max 60s)
- **Queue:** Use existing job queue infrastructure
- **Timeout:** 30s per LLM call
- **Retry:** Max 3 retries for transient failures

## Architecture Components

### Domain Entity
File: `main-service/internal/core/relation/entity_relation.go`
```go
type EntityRelation struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    WorldID         uuid.UUID
    SourceType      string
    SourceID        *uuid.UUID
    SourceRef       string
    TargetType      string
    TargetID        *uuid.UUID
    TargetRef       string
    RelationType    string
    ContextType     *string
    ContextID       *uuid.UUID
    Attributes      map[string]interface{}
    Confidence      *float64
    EvidenceSpans   []EvidenceSpan
    Summary         string
    Status          RelationStatus
    MirrorID        *uuid.UUID  // Points to auto-created inverse relation
    CreatedByUserID *uuid.UUID
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type EvidenceSpan struct {
    Start int    `json:"start"`
    End   int    `json:"end"`
    Text  string `json:"text"`
}

type RelationStatus string
const (
    RelationStatusDraft     RelationStatus = "draft"
    RelationStatusConfirmed RelationStatus = "confirmed"
    RelationStatusOrphan    RelationStatus = "orphan"
)

// Inverse relation type mapping
var inverseRelations = map[string]string{
    "parent_of":    "child_of",
    "child_of":     "parent_of",
    "sibling_of":   "sibling_of",
    "spouse_of":    "spouse_of",
    "ally_of":      "ally_of",
    "enemy_of":     "enemy_of",
    "member_of":    "has_member",
    "has_member":   "member_of",
    "leader_of":    "led_by",
    "led_by":       "leader_of",
    "located_in":   "contains",
    "contains":     "located_in",
    "owns":         "owned_by",
    "owned_by":     "owns",
    "mentor_of":    "mentored_by",
    "mentored_by":  "mentor_of",
}

func GetInverseRelationType(relationType string) string {
    if inverse, ok := inverseRelations[relationType]; ok {
        return inverse
    }
    return relationType // fallback: symmetric
}

// CreateMirrorRelation creates the inverse relation
func (r *EntityRelation) CreateMirrorRelation() *EntityRelation {
    mirror := &EntityRelation{
        ID:              uuid.New(),
        TenantID:        r.TenantID,
        WorldID:         r.WorldID,
        SourceType:      r.TargetType,
        SourceID:        r.TargetID,
        SourceRef:       r.TargetRef,
        TargetType:      r.SourceType,
        TargetID:        r.SourceID,
        TargetRef:       r.SourceRef,
        RelationType:    GetInverseRelationType(r.RelationType),
        ContextType:     r.ContextType,
        ContextID:       r.ContextID,
        Attributes:      r.Attributes,
        Confidence:      r.Confidence,
        EvidenceSpans:   r.EvidenceSpans,
        Status:          r.Status,
        MirrorID:        &r.ID,
        CreatedByUserID: r.CreatedByUserID,
        CreatedAt:       r.CreatedAt,
        UpdatedAt:       r.UpdatedAt,
    }
    r.MirrorID = &mirror.ID
    return mirror
}
```

### Repository Interface
File: `main-service/internal/ports/repositories/entity_relation.go`
```go
type EntityRelationRepository interface {
    // CRUD
    Create(ctx context.Context, r *relation.EntityRelation) error
    CreateWithMirror(ctx context.Context, r *relation.EntityRelation) (*relation.EntityRelation, error) // Creates both relation and mirror
    GetByID(ctx context.Context, tenantID, id uuid.UUID) (*relation.EntityRelation, error)
    Update(ctx context.Context, r *relation.EntityRelation) error
    Delete(ctx context.Context, tenantID, id uuid.UUID) error // Also deletes mirror if exists
    
    // List with cursor pagination
    ListBySource(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID, opts ListOptions) (*ListResult, error)
    ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, opts ListOptions) (*ListResult, error)
    ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, opts ListOptions) (*ListResult, error)
    
    // Temp ID resolution
    ResolveTemporaryRef(ctx context.Context, tenantID uuid.UUID, tempRef string, realID uuid.UUID) error
    
    // Maintenance
    MarkOrphan(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) error
    CleanupExpiredDrafts(ctx context.Context, olderThan time.Time) (int, error)
}

// Cursor-based pagination
type ListOptions struct {
    Cursor        *string         // Opaque cursor from previous response
    Limit         int             // Default 50, max 100
    RelationType  *string
    Status        *RelationStatus
    MinConfidence *float64
    OrderBy       string          // "created_at", "confidence"
    OrderDir      string          // "asc", "desc"
    ExcludeMirrors bool           // If true, only returns primary relations (id < mirror_id)
}

type ListResult struct {
    Items      []*relation.EntityRelation
    NextCursor *string
    HasMore    bool
}

// Cursor encoding/decoding
type Cursor struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}

func EncodeCursor(c Cursor) string {
    data, _ := json.Marshal(c)
    return base64.StdEncoding.EncodeToString(data)
}

func DecodeCursor(s string) (*Cursor, error) {
    data, err := base64.StdEncoding.DecodeString(s)
    if err != nil {
        return nil, err
    }
    var c Cursor
    if err := json.Unmarshal(data, &c); err != nil {
        return nil, err
    }
    return &c, nil
}
```

### Use Cases
File: `main-service/internal/application/relation/`
- `create_relation.go` - Create new relation
- `list_relations.go` - List with filters
- `update_relation.go` - Update relation
- `delete_relation.go` - Delete relation
- `resolve_temp_ids.go` - Resolve temp → real IDs
- ~~`discover_world_relations.go`~~ - **Removido**: Discovery job será implementado no `llm-gateway-service`

### gRPC Proto
File: `main-service/proto/entity_relation.proto`
```protobuf
message EntityRelation {
    string id = 1;
    string tenant_id = 2;
    string world_id = 3;
    string source_type = 4;
    optional string source_id = 5;
    string source_ref = 6;
    string target_type = 7;
    optional string target_id = 8;
    string target_ref = 9;
    string relation_type = 10;
    optional string context_type = 11;
    optional string context_id = 12;
    string attributes_json = 13;
    optional float confidence = 14;
    string summary = 15;
    string status = 16;
    google.protobuf.Timestamp created_at = 17;
    google.protobuf.Timestamp updated_at = 18;
}

service EntityRelationService {
    rpc CreateRelation(CreateRelationRequest) returns (CreateRelationResponse);
    rpc GetRelation(GetRelationRequest) returns (EntityRelation);
    rpc ListRelationsBySource(ListRelationsBySourceRequest) returns (ListRelationsResponse);
    rpc ListRelationsByTarget(ListRelationsByTargetRequest) returns (ListRelationsResponse);
    rpc UpdateRelation(UpdateRelationRequest) returns (EntityRelation);
    rpc DeleteRelation(DeleteRelationRequest) returns (google.protobuf.Empty);
}
```

## Embedding Ingestion Trigger
When a relation is created/updated:
1. Generate summary: `"<source_name> <relation_type> <target_name>"`
2. Enqueue embedding job to `llm-gateway-service`
3. Chunk type: `relation`
4. Metadata: `{ "relation_id": "...", "world_id": "...", "source_type": "...", "target_type": "..." }`

On relation delete → delete corresponding embedding chunk.

## Open Questions
- ~~Which relation types are in v1?~~ (resolved - see below)
- How strict should evidence validation be before persistence?
  - **Proposal:** No validation in v1, store as-is
- How much relation detection should be heuristic vs LLM?
  - **Proposal:** Start 100% heuristic, add LLM validation as opt-in
- Confidence threshold for auto-confirm?
  - **Proposal:** confidence >= 0.8 auto-confirms, else stays draft

## Relation Types v1 (complete)
### Entity↔Entity (with auto-mirror)
| Primary | Inverse (auto-created) |
|---------|------------------------|
| `parent_of` | `child_of` |
| `child_of` | `parent_of` |
| `sibling_of` | `sibling_of` (symmetric) |
| `spouse_of` | `spouse_of` (symmetric) |
| `ally_of` | `ally_of` (symmetric) |
| `enemy_of` | `enemy_of` (symmetric) |
| `member_of` | `has_member` |
| `has_member` | `member_of` |
| `leader_of` | `led_by` |
| `led_by` | `leader_of` |
| `located_in` | `contains` |
| `contains` | `located_in` |
| `owns` | `owned_by` |
| `owned_by` | `owns` |
| `mentor_of` | `mentored_by` |
| `mentored_by` | `mentor_of` |

### Context Relations (from existing tables)
- `mentions` (content_block → entity, scene → entity)
- `references` (event/lore/faction/artifact → entity)
- `role` (event → character: "protagonist", "antagonist", "witness")
- `significance` (event → location: "primary", "secondary")

### Legacy (for migration)
- `relationship` (maps from character_relationships.relationship_type)

---

## Implementation TODO

### ⚠️ Phase 0: Emergency Fix - Make Code Compile (CRITICAL)

**Status**: 🔴 ALL ENTRYPOINTS BROKEN - Code does not compile

**Root Cause**: Old reference types and repository interfaces were removed from `core/` but still referenced in:
- Repository interfaces (`ports/repositories/`)
- gRPC mappers (`transport/grpc/mappers/`)
- gRPC handlers (`transport/grpc/handlers/`)
- HTTP handlers (`transport/http/handlers/`)
- Main entrypoints (`cmd/api-http/`, `cmd/api-grpc/`, `cmd/api-offline/`)

#### 0.1 Remove Old Repository Interfaces
- [x] **0.1a** Remove `CharacterRelationshipRepository` interface from `internal/ports/repositories/character.go`
- [x] **0.1b** Remove `ArtifactReferenceRepository` interface from `internal/ports/repositories/artifact.go`
- [x] **0.1c** Remove `EventReferenceRepository` interface from `internal/ports/repositories/event.go`
- [x] **0.1d** Remove `FactionReferenceRepository` interface from `internal/ports/repositories/faction.go`
- [x] **0.1e** Remove `LoreReferenceRepository` interface from `internal/ports/repositories/lore.go`

#### 0.2 Fix gRPC Mappers
- [x] **0.2a** Remove/Update `SceneReferenceToProto` in `internal/transport/grpc/mappers/scene_mapper.go` (uses `story.SceneReference` which doesn't exist)
- [x] **0.2b** Remove/Update `CharacterRelationshipToProto` in `internal/transport/grpc/mappers/character_mapper.go`
- [x] **0.2c** Remove/Update `EventReferenceToProto` in `internal/transport/grpc/mappers/event_mapper.go`
- [x] **0.2d** Remove/Update `FactionReferenceToProto` in `internal/transport/grpc/mappers/faction_mapper.go`
- [x] **0.2e** Remove/Update `LoreReferenceToProto` in `internal/transport/grpc/mappers/lore_mapper.go`
- [x] **0.2f** Remove/Update `ArtifactReferenceToProto` in `internal/transport/grpc/mappers/artifact_mapper.go`
- [x] **0.2g** Fix `entity_relation_mapper.go` line 57: `Confidence` type mismatch (`*float64` vs `*float32`)

#### 0.3 Fix gRPC Handlers
- [x] **0.3a** Update `internal/transport/grpc/handlers/character_handler.go` - remove old CharacterRelationship methods or adapt to entity_relations
- [x] **0.3b** Update `internal/transport/grpc/handlers/event_handler.go` - adapt EventReference to entity_relations
- [x] **0.3c** Update `internal/transport/grpc/handlers/faction_handler.go` - adapt FactionReference to entity_relations
- [x] **0.3d** Update `internal/transport/grpc/handlers/lore_handler.go` - adapt LoreReference to entity_relations
- [x] **0.3e** Update `internal/transport/grpc/handlers/artifact_handler.go` - adapt ArtifactReference to entity_relations
- [x] **0.3f** Update `internal/transport/grpc/handlers/scene_handler.go` - adapt SceneReference to entity_relations

#### 0.4 Fix HTTP Handlers
- [x] **0.4a** Update `internal/transport/http/handlers/character_handler.go` - remove import of `characterrelationshipapp` and adapt relationship methods

#### 0.5 Fix Main Entrypoints
- [x] **0.5a** Update `cmd/api-http/main.go` - remove character_relationship import, inject relationapp for character relationships
- [x] **0.5b** Update `cmd/api-grpc/main.go` - same fixes as api-http
- [x] **0.5c** Update `cmd/api-offline/main.go` - major rewrite:
  - Remove all old reference repository instantiations
  - Use `sqlite.NewEntityRelationRepository()`
  - Create `relationapp` use cases
  - Update all handlers to use new dependencies

#### 0.6 Fix Test Files (can be done later, but need to compile)
- [ ] **0.6a** Update or skip `*_handler_test.go` files that reference old types

---

### Phase 1: Foundation (Core Infrastructure)
- [x] **1.1** Create domain entity `EntityRelation` in `main-service/internal/core/relation/entity_relation.go`
- [x] **1.2** Create `RelationStatus` enum and validation methods
- [x] **1.3** Create `EvidenceSpan` struct with JSON marshaling
- [x] **1.4** Create `inverseRelations` map and `GetInverseRelationType()` function
- [x] **1.5** Create `CreateMirrorRelation()` method on EntityRelation
- [x] **1.6** Define repository interface with cursor pagination in `main-service/internal/ports/repositories/entity_relation.go`
- [x] **1.7** Create `Cursor` struct with encode/decode functions
- [x] **1.8** Create Postgres migration: `entity_relations` table with all indexes + `mirror_id`
- [x] **1.9** Create SQLite migration: `entity_relations` table with all indexes + `mirror_id`
- [x] **1.10** Implement Postgres repository `main-service/internal/adapters/db/postgres/entity_relation_repository.go`
- [x] **1.11** Implement SQLite repository `main-service/internal/adapters/db/sqlite/entity_relation_repository.go`
- [x] **1.12** Implement `CreateWithMirror()` that creates both relation and inverse in single transaction
- [ ] **1.13** Write repository unit tests (Postgres + SQLite)

### Phase 2: Use Cases
- [x] **2.1** Create `CreateRelation` use case
- [x] **2.2** Create `GetRelation` use case  
- [x] **2.3** Create `ListRelationsBySource` use case with pagination/filters
- [x] **2.4** Create `ListRelationsByTarget` use case with pagination/filters
- [x] **2.5** Create `ListRelationsByWorld` use case with pagination/filters
- [x] **2.6** Create `UpdateRelation` use case
- [x] **2.7** Create `DeleteRelation` use case
- [x] **2.8** Create `ResolveTemporaryIds` use case (batch)
- [ ] **2.9** Write use case unit tests

### Phase 3: HTTP Transport
- [x] **3.1** Create HTTP handler `entity_relation_handler.go`
- [x] **3.2** Add generic routes: `GET/POST /api/v1/relations`, `GET/PUT/DELETE /api/v1/relations/{id}`
- [x] **3.3** Add world-scoped route: `GET /api/v1/worlds/{world_id}/relations`
- [x] **3.4** Add resolve temp IDs route: `POST /api/v1/relations/resolve-temp`
- [x] **3.5** Register routes in router
- [x] **3.6** Add pagination response wrapper
- [ ] **3.7** Write HTTP handler tests

### Phase 4: gRPC Transport
- [x] **4.1** Create proto definition `proto/entity_relation.proto`
- [ ] **4.2** Generate Go code from proto (user needs to run: `make proto` or `protoc`)
- [x] **4.3** Create gRPC service implementation
- [x] **4.4** Create gRPC ↔ domain mappers
- [x] **4.5** Register service in gRPC server

### Phase 5: Migration - Greenfield (No Backfill)
- [x] **5.1** Create migration to drop old tables: `character_relationships`, `event_references`, `lore_references`, `faction_references`, `artifact_references`, `scene_references` (note: `content_block_references` will be migrated in Phase 6)
- [x] **5.2** Remove old repository implementations (Postgres + SQLite)
- [x] **5.3** Remove old domain entities and use cases

### Phase 6: Update Existing Endpoints
- [x] **6.1** ~~Update `character_relationship_handler.go`~~ - Merged into 0.4a (Emergency Fix)
- [x] **6.2** Update `event_reference` use cases to use `entity_relations` (add_reference, get_references, remove_reference, update_reference)
- [x] **6.2b** Update `delete_event` use case to use `entity_relations` (mark relations as orphan)
- [x] **6.2c** Update `main.go` for event use cases to use new `entity_relations` dependencies
- [x] **6.2d** Update `character/get_events.go` to use `entity_relations`
- [x] **6.3** Update `lore_reference` use cases to use `entity_relations` (add_reference, get_references, remove_reference, update_reference, delete_lore)
- [x] **6.3b** Update `main.go` for lore use cases to use new `entity_relations` dependencies
- [x] **6.4** Update `faction_reference` use cases to use `entity_relations` (add_reference, get_references, remove_reference, update_reference, delete_faction)
- [x] **6.4b** Update `main.go` for faction use cases to use new `entity_relations` dependencies
- [x] **6.5** Update `artifact_reference` use cases to use `entity_relations` (add_reference, get_references, remove_reference, create_artifact, update_artifact, delete_artifact)
- [x] **6.5b** Update `main.go` for artifact use cases to use new `entity_relations` dependencies
- [x] **6.6** Update `scene_reference` use cases to use `entity_relations` (add_reference, get_references, remove_reference)
- [x] **6.6b** Update `main.go` for scene use cases to use new `entity_relations` dependencies
- [~] **6.7** ~~Update `content_block_reference` handlers to use `entity_relations`~~ **SKIPPED**: Será renomeado para `content_anchors` (ver Phase 11)
- [x] **6.8** Adapt `character_relationship` endpoints to use `entity_relations` (maintain compatibility with existing DTO structure):
  - [x] **6.8a** Create adapter layer in `CharacterHandler` to convert old DTO format to `entity_relations`
  - [x] **6.8b** Create `CharacterRelationshipDTO` for response compatibility (map from `EntityRelation`)
  - [x] **6.8c** Update `ListRelationships` - query entity_relations where source_type=character OR target_type=character
  - [x] **6.8d** Update `CreateRelationship` - create entity_relation with proper mapping
  - [x] **6.8e** Update `UpdateRelationship` - update entity_relation, map attributes
  - [x] **6.8f** Update `DeleteRelationship` - delete entity_relation (and mirror)
  - [x] **6.8g** Update `main.go` to inject `relationapp` use cases into `CharacterHandler`
- [x] **6.9** Update Postman collection with new endpoints

### Phase 7: Cascade & Cleanup
- [x] **7.1** Add hook/trigger on entity delete to mark relations as `orphan` (added to character, location, scene, event, lore, faction, artifact)
- [x] **7.2** Create cleanup job: delete expired draft relations with temp refs (`CleanupExpiredDraftsUseCase`)
- [x] **7.3** Create cleanup job: delete orphan relations older than N days (`CleanupOrphanRelationsUseCase`)
- [x] **7.4** Add configuration for cleanup intervals and retention periods (`CleanupConfig` in config)
- [ ] **7.5** Test cascade behavior (manual testing required)

### Phase 8: Embedding Ingestion
- [x] **8.1** Create summary generator function (no LLM) - `SummaryGenerator` com `GenerateSummary` e `GenerateSummaryFromRefs`
- [x] **8.2** Add relation create/update hook to enqueue embedding job - hooks em `CreateRelationUseCase` e `UpdateRelationUseCase` (apenas para `confirmed`)
- [x] **8.3** Add relation delete hook to delete embedding chunk - delete automático via llm-gateway-service (verificação periódica)
- [x] **8.4** Define `relation` chunk type in llm-gateway-service - usar `sourceType="relation"` na ingestion queue
- [ ] **8.5** Test embedding search with relation summaries (manual testing required)

### Phase 9: Discovery Pipeline (Future - llm-gateway-service)
**Nota**: Todo o discovery pipeline será implementado no `llm-gateway-service`, não no `main-service`. O `main-service` apenas fornece os gRPC endpoints para o worker buscar entidades e criar relações.

- [ ] **9.1** (llm-gateway) Create discovery worker/job que escuta eventos ou endpoint
- [ ] **9.2** (llm-gateway) Implement heuristic relation detection (co-occurrence, patterns)
- [ ] **9.3** (llm-gateway) Add LLM validation (optional, Phase 3 of pipeline)
- [ ] **9.4** (llm-gateway) Add rate limiting for LLM calls
- [ ] **9.5** (llm-gateway) Worker chama `main-service` via gRPC para criar relações descobertas
- [ ] **9.6** (llm-gateway) Gerencia estado do job (própria tabela ou mecanismo interno)
- [ ] **9.7** (llm-gateway) Test discovery with sample world

### Phase 10: Documentation & Cleanup
- [ ] **10.1** Update API documentation with new endpoints
- [ ] **10.2** Update Postman collection
- [ ] **10.3** Update README with relation types and usage

### Phase 11: Rename content_block_references → content_anchors
**Rationale**: `content_block_references` are NOT semantic entity relations. They are **text anchors** (mentions within prose content). Renaming clarifies this distinction.

#### 11.1 Database Migration
- [x] **11.1a** Create Postgres migration to rename table: `content_block_references` → `content_anchors`
- [x] **11.1b** Create SQLite migration to rename table: `content_block_references` → `content_anchors`
- [x] **11.1c** Update indexes and constraints accordingly

#### 11.2 Domain Layer
- [x] **11.2a** Rename `internal/core/story/content_block_reference.go` → `content_anchor.go`
- [x] **11.2b** Rename struct `ContentBlockReference` → `ContentAnchor`
- [x] **11.2c** Update all type references in core

#### 11.3 Repository Layer
- [x] **11.3a** Rename `ContentBlockReferenceRepository` interface → `ContentAnchorRepository`
- [x] **11.3b** Rename Postgres implementation file and struct
- [x] **11.3c** Rename SQLite implementation file and struct
- [x] **11.3d** Update table name in all queries

#### 11.4 Application Layer
- [x] **11.4a** Rename use cases in `application/story/content_block/`:
  - `create_content_block_reference.go` → `create_content_anchor.go`
  - `list_content_block_references_by_content_block.go` → `list_content_anchors_by_content_block.go`
  - `list_content_blocks_by_entity.go` (uses anchors internally)
  - `delete_content_block_reference.go` → `delete_content_anchor.go`
- [x] **11.4b** Update use case struct names and methods

#### 11.5 Transport Layer
- [x] **11.5a** Rename HTTP handler: `content_block_reference_handler.go` → `content_anchor_handler.go`
- [x] **11.5b** Update HTTP routes (keep old routes as aliases for backwards compatibility):
  - `POST /api/v1/content-blocks/{id}/anchors` (new)
  - `POST /api/v1/content-blocks/{id}/references` (deprecated alias)
- [x] **11.5c** Rename gRPC proto messages if applicable
- [x] **11.5d** Update gRPC handlers

#### 11.6 Main Entrypoints
- [x] **11.6a** Update `cmd/api-http/main.go` with new handler names
- [x] **11.6b** Update `cmd/api-grpc/main.go` with new handler names
- [x] **11.6c** Update `cmd/api-offline/main.go` with new handler names

#### 11.7 Tests
- [x] **11.7a** Rename test files to match new names
- [x] **11.7b** Update test assertions and mocks

### Phase 12: API Simplification - Remove LLM-only fields from main-service ✅ COMPLETED
**Rationale**: Os campos `source_ref`/`target_ref`, `confidence`, `evidence_spans`, e `status` existiam para o pipeline de LLM/entity extraction onde entidades podem não existir ainda. Mas na API pública de criar relações, ambas entidades já devem existir. Manter esses campos na API cria inconsistências.

**Decisão de Design:**
- Removemos completamente `source_ref`, `target_ref`, `confidence`, `evidence_spans`, e `status` do main-service
- API pública (`POST /api/v1/relations`) aceita apenas `source_id`/`target_id` (obrigatórios)
- Todas as relações criadas via API são automaticamente `confirmed` (sem conceito de draft/orphan no main-service)
- O pipeline de LLM será implementado no `llm-gateway-service` com seu próprio schema

#### 12.1 Core Domain ✅
- [x] **12.1a** Remover `SourceRef`, `TargetRef`, `Confidence`, `EvidenceSpans`, `Status` da struct `EntityRelation`
- [x] **12.1b** Atualizar `NewEntityRelation` para aceitar apenas IDs (não refs)
- [x] **12.1c** Remover `RelationStatus` enum (draft, orphan) - apenas confirmed implícito
- [x] **12.1d** Atualizar `Validate()` para validar apenas campos necessários
- [x] **12.1e** Atualizar `CreateMirrorRelation` para refletir mudanças

#### 12.2 Repository Layer ✅
- [x] **12.2a** Atualizar `EntityRelationRepository` interface (remover métodos de refs)
- [x] **12.2b** Atualizar Postgres repository (queries sem colunas removidas)
- [x] **12.2c** Atualizar SQLite repository (queries sem colunas removidas)

#### 12.3 Application Layer ✅
- [x] **12.3a** Atualizar `CreateRelationInput` - remover refs, confidence, evidence_spans, status
- [x] **12.3b** Atualizar `UpdateRelationInput` - remover confidence, evidence_spans, status
- [x] **12.3c** Remover `ResolveTemporaryIdsUseCase` - não mais necessário
- [x] **12.3d** Remover `CleanupOrphanRelationsUseCase` - não mais necessário (sem status orphan)
- [x] **12.3e** Remover `CleanupExpiredDraftsUseCase` - não mais necessário (sem status draft)
- [x] **12.3f** Atualizar `SummaryGenerator` - usar IDs diretamente

#### 12.4 HTTP Transport ✅
- [x] **12.4a** Atualizar `EntityRelationHandler.Create` - remover refs do request body
- [x] **12.4b** Atualizar `EntityRelationHandler.Update` - remover confidence, evidence_spans, status
- [x] **12.4c** Remover endpoint `POST /api/v1/relations/resolve-temp`
- [x] **12.4d** Remover parâmetros de filtro status e confidence_min no ListOptions
- [x] **12.4e** Atualizar `NewEntityRelationHandler` - remover resolveTemporaryIdsUseCase

#### 12.5 gRPC Transport ✅
- [x] **12.5a** Atualizar `EntityRelationHandler` gRPC - remover refs, confidence, evidence_spans, status
- [x] **12.5b** Atualizar mapper `EntityRelationToProto` - refletir campos removidos
- [x] **12.5c** Atualizar `NewEntityRelationHandler` - remover resolveTemporaryIdsUseCase

#### 12.6 Main Entrypoints ✅
- [x] **12.6a** Atualizar `cmd/api-http/main.go` - remover criação de resolveTemporaryIdsUseCase
- [x] **12.6b** Atualizar `cmd/api-grpc/main.go` - remover criação de resolveTemporaryIdsUseCase

#### 12.7 Database Migration ✅
- [x] **12.7a** Criar migration Postgres `065_simplify_entity_relations` - remover colunas, tornar source_id/target_id NOT NULL
- [x] **12.7b** Criar migration SQLite `012_simplify_entity_relations` - recriar tabela sem colunas removidas

#### 12.8 Future: LLM Pipeline (llm-gateway-service)
- [ ] **12.8a** Criar schema próprio no llm-gateway para relations descobertas (com confidence, evidence_spans, status draft)
- [ ] **12.8b** Worker de discovery cria relações "candidatas" localmente
- [ ] **12.8c** Quando confirmadas, chama main-service via gRPC para persistir relação final

---

## Priority Order (Updated)
1. ✅ **Phase 0** (Emergency Fix) - Make code compile
2. ✅ **Phase 6** (Update Endpoints) - Including 6.8 Character Relationship Compatibility
3. ✅ **Phase 7-8** (Cascade + Embedding) - Data integrity + Search
4. ✅ **Phase 12** (API Simplification) - Remove LLM-only fields from main-service
5. 📋 **Phase 1.13, 2.9, 3.7** (Tests) - Add missing unit tests
6. 📋 **Phase 11** (content_anchors rename) - Clarify semantics
7. ⚠️ **Phase 4** (gRPC full support) - If needed by clients
8. 📋 **Phase 9** (Discovery) - Advanced feature (llm-gateway-service)
9. 📋 **Phase 10** (Documentation) - Final cleanup

---

## Estimated Effort (Updated)
| Phase | Effort | Status |
|-------|--------|--------|
| 0. Emergency Fix | ~~1-2 days~~ | ✅ Done |
| 1. Foundation | ~~3-4 days~~ | ✅ Done |
| 2. Use Cases | ~~2 days~~ | ✅ Done |
| 3. HTTP Transport | ~~1-2 days~~ | ✅ Done |
| 4. gRPC Transport | 0.5 day | ⚠️ Needs fixes |
| 5. Migration | ~~0.5 day~~ | ✅ Done |
| 6. Update Endpoints | ~~1 day~~ | ✅ Done |
| 7. Cascade & Cleanup | ~~1 day~~ | ✅ Done |
| 8. Embedding Ingestion | ~~1-2 days~~ | ✅ Done |
| 9. Discovery Pipeline | 3-5 days | 📋 Future (llm-gateway) |
| 10. Documentation | 0.5 day | 📋 Pending |
| 11. content_anchors rename | 1-2 days | 📋 New |
| 12. API Simplification | ~~1 day~~ | ✅ Done |
| 1.13, 2.9, 3.7 Tests | 2-3 days | 📋 Pending |
| **Remaining** | **~6-10 days** |
