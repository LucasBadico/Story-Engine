-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table

-- First, drop indexes
DROP INDEX IF EXISTS idx_entity_relations_source_ref;
DROP INDEX IF EXISTS idx_entity_relations_target_ref;
DROP INDEX IF EXISTS idx_entity_relations_status;
DROP INDEX IF EXISTS idx_entity_relations_tenant_id;
DROP INDEX IF EXISTS idx_entity_relations_world_id;
DROP INDEX IF EXISTS idx_entity_relations_source;
DROP INDEX IF EXISTS idx_entity_relations_target;
DROP INDEX IF EXISTS idx_entity_relations_relation_type;
DROP INDEX IF EXISTS idx_entity_relations_context;

-- Create new table without the removed columns
CREATE TABLE entity_relations_new (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,  -- Now required
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,  -- Now required
    relation_type TEXT NOT NULL,
    context_type TEXT,
    context_id TEXT,
    attributes TEXT NOT NULL DEFAULT '{}',
    summary TEXT NOT NULL DEFAULT '',
    mirror_id TEXT,
    created_by_user_id TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT entity_relations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT entity_relations_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT entity_relations_mirror_fk FOREIGN KEY (mirror_id) REFERENCES entity_relations_new(id) ON DELETE SET NULL
);

-- Copy data (only rows with valid source_id and target_id)
INSERT INTO entity_relations_new (
    id, tenant_id, world_id, source_type, source_id, target_type, target_id,
    relation_type, context_type, context_id, attributes, summary, mirror_id,
    created_by_user_id, created_at, updated_at
)
SELECT 
    id, tenant_id, world_id, source_type, source_id, target_type, target_id,
    relation_type, context_type, context_id, attributes, summary, mirror_id,
    created_by_user_id, created_at, updated_at
FROM entity_relations
WHERE source_id IS NOT NULL AND target_id IS NOT NULL;

-- Drop old table
DROP TABLE entity_relations;

-- Rename new table
ALTER TABLE entity_relations_new RENAME TO entity_relations;

-- Recreate indexes
CREATE INDEX idx_entity_relations_tenant_id ON entity_relations(tenant_id);
CREATE INDEX idx_entity_relations_world_id ON entity_relations(world_id);
CREATE INDEX idx_entity_relations_source ON entity_relations(source_type, source_id);
CREATE INDEX idx_entity_relations_target ON entity_relations(target_type, target_id);
CREATE INDEX idx_entity_relations_relation_type ON entity_relations(relation_type);
CREATE INDEX idx_entity_relations_context ON entity_relations(context_type, context_id);

