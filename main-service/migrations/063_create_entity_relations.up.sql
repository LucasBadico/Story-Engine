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

