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

