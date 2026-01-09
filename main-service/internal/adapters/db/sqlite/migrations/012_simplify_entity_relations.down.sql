-- Restore the original table structure with all columns

-- Drop indexes
DROP INDEX IF EXISTS idx_entity_relations_tenant_id;
DROP INDEX IF EXISTS idx_entity_relations_world_id;
DROP INDEX IF EXISTS idx_entity_relations_source;
DROP INDEX IF EXISTS idx_entity_relations_target;
DROP INDEX IF EXISTS idx_entity_relations_relation_type;
DROP INDEX IF EXISTS idx_entity_relations_context;

-- Create table with original structure
CREATE TABLE entity_relations_new (
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
    status TEXT NOT NULL DEFAULT 'confirmed',
    mirror_id TEXT,
    created_by_user_id TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT entity_relations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT entity_relations_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT entity_relations_mirror_fk FOREIGN KEY (mirror_id) REFERENCES entity_relations_new(id) ON DELETE SET NULL,
    CONSTRAINT entity_relations_status_check CHECK (status IN ('draft', 'confirmed', 'orphan'))
);

-- Copy data, using source_id/target_id as source_ref/target_ref
INSERT INTO entity_relations_new (
    id, tenant_id, world_id, source_type, source_id, source_ref, target_type, target_id, target_ref,
    relation_type, context_type, context_id, attributes, summary, status, mirror_id,
    created_by_user_id, created_at, updated_at
)
SELECT 
    id, tenant_id, world_id, source_type, source_id, source_id, target_type, target_id, target_id,
    relation_type, context_type, context_id, attributes, summary, 'confirmed', mirror_id,
    created_by_user_id, created_at, updated_at
FROM entity_relations;

-- Drop simplified table
DROP TABLE entity_relations;

-- Rename new table
ALTER TABLE entity_relations_new RENAME TO entity_relations;

-- Recreate all indexes
CREATE INDEX idx_entity_relations_tenant_id ON entity_relations(tenant_id);
CREATE INDEX idx_entity_relations_world_id ON entity_relations(world_id);
CREATE INDEX idx_entity_relations_source ON entity_relations(source_type, source_id);
CREATE INDEX idx_entity_relations_target ON entity_relations(target_type, target_id);
CREATE INDEX idx_entity_relations_relation_type ON entity_relations(relation_type);
CREATE INDEX idx_entity_relations_status ON entity_relations(status);
CREATE INDEX idx_entity_relations_context ON entity_relations(context_type, context_id);
CREATE INDEX idx_entity_relations_source_ref ON entity_relations(source_ref);
CREATE INDEX idx_entity_relations_target_ref ON entity_relations(target_ref);

