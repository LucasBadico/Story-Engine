CREATE TABLE IF NOT EXISTS lore_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    lore_id UUID NOT NULL REFERENCES lores(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    relationship_type VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(lore_id, entity_type, entity_id),
    CONSTRAINT lore_references_entity_type_check CHECK (entity_type IN (
        'character', 'location', 'artifact', 'event', 
        'faction', 'lore',
        'faction_reference'
    ))
);

CREATE INDEX idx_lore_references_tenant_id ON lore_references(tenant_id);
CREATE INDEX idx_lore_references_lore_id ON lore_references(lore_id);
CREATE INDEX idx_lore_references_entity ON lore_references(entity_type, entity_id);

