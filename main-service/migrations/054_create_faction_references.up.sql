CREATE TABLE IF NOT EXISTS faction_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    faction_id UUID NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    role VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(faction_id, entity_type, entity_id),
    CONSTRAINT faction_references_entity_type_check CHECK (entity_type IN (
        'character', 'location', 'artifact', 'event', 
        'lore', 'faction', 
        'lore_reference'
    ))
);

CREATE INDEX idx_faction_references_tenant_id ON faction_references(tenant_id);
CREATE INDEX idx_faction_references_faction_id ON faction_references(faction_id);
CREATE INDEX idx_faction_references_entity ON faction_references(entity_type, entity_id);

