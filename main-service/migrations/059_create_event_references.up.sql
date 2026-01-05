CREATE TABLE IF NOT EXISTS event_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    relationship_type VARCHAR(100), -- "role" para character/artifact, "significance" para location
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, entity_type, entity_id),
    CONSTRAINT event_references_entity_type_check CHECK (entity_type IN (
        'character', 'location', 'artifact', 'faction', 'lore', 
        'faction_reference', 'lore_reference'
    ))
);

CREATE INDEX idx_event_references_tenant_id ON event_references(tenant_id);
CREATE INDEX idx_event_references_event_id ON event_references(event_id);
CREATE INDEX idx_event_references_entity ON event_references(entity_type, entity_id);

-- Migrar dados de event_characters
INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, created_at)
SELECT id, tenant_id, event_id, 'character', character_id, role, created_at
FROM event_characters
ON CONFLICT (event_id, entity_type, entity_id) DO NOTHING;

-- Migrar dados de event_locations
INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, created_at)
SELECT id, tenant_id, event_id, 'location', location_id, significance, created_at
FROM event_locations
ON CONFLICT (event_id, entity_type, entity_id) DO NOTHING;

-- Migrar dados de event_artifacts
INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, created_at)
SELECT id, tenant_id, event_id, 'artifact', artifact_id, role, created_at
FROM event_artifacts
ON CONFLICT (event_id, entity_type, entity_id) DO NOTHING;

-- Remover tabelas antigas
DROP TABLE IF EXISTS event_characters;
DROP TABLE IF EXISTS event_locations;
DROP TABLE IF EXISTS event_artifacts;

