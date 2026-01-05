CREATE TABLE IF NOT EXISTS event_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    relationship_type TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT event_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT event_references_event_fk FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT event_references_entity_type_check CHECK (entity_type IN (
        'character', 'location', 'artifact', 'faction', 'lore',
        'faction_reference', 'lore_reference'
    )),
    CONSTRAINT event_references_event_entity_unique UNIQUE (event_id, entity_type, entity_id)
);

CREATE INDEX idx_event_references_tenant_id ON event_references(tenant_id);
CREATE INDEX idx_event_references_event_id ON event_references(event_id);
CREATE INDEX idx_event_references_entity ON event_references(entity_type, entity_id);

-- Migrar dados de event_characters
INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, created_at)
SELECT id, tenant_id, event_id, 'character', character_id, role, created_at
FROM event_characters;

-- Migrar dados de event_locations
INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, created_at)
SELECT id, tenant_id, event_id, 'location', location_id, significance, created_at
FROM event_locations;

-- Migrar dados de event_artifacts
INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, created_at)
SELECT id, tenant_id, event_id, 'artifact', artifact_id, role, created_at
FROM event_artifacts;

-- Remover tabelas antigas
DROP TABLE IF EXISTS event_characters;
DROP TABLE IF EXISTS event_locations;
DROP TABLE IF EXISTS event_artifacts;

