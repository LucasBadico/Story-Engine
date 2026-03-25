CREATE TABLE IF NOT EXISTS relationship_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,
    tension_description TEXT NOT NULL,
    timeline_day INT,
    timeline_confidence VARCHAR(16),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_relationship_states_chapter ON relationship_states(chapter_id);
CREATE INDEX idx_relationship_states_chapter_number ON relationship_states(chapter_number DESC);
CREATE INDEX idx_relationship_states_tenant ON relationship_states(tenant_id);

CREATE TABLE IF NOT EXISTS relationship_state_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_state_id UUID NOT NULL REFERENCES relationship_states(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rs_entities_state ON relationship_state_entities(relationship_state_id);
CREATE INDEX idx_rs_entities_entity ON relationship_state_entities(entity_type, entity_id);
