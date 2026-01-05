-- Recriar tabelas antigas
CREATE TABLE IF NOT EXISTS event_characters (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    role TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT event_characters_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT event_characters_event_fk FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT event_characters_character_fk FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT event_characters_event_character_unique UNIQUE (event_id, character_id)
);

CREATE TABLE IF NOT EXISTS event_locations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    significance TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT event_locations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT event_locations_event_fk FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT event_locations_location_fk FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE,
    CONSTRAINT event_locations_event_location_unique UNIQUE (event_id, location_id)
);

CREATE TABLE IF NOT EXISTS event_artifacts (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    artifact_id TEXT NOT NULL,
    role TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT event_artifacts_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT event_artifacts_event_fk FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT event_artifacts_artifact_fk FOREIGN KEY (artifact_id) REFERENCES artifacts(id) ON DELETE CASCADE,
    CONSTRAINT event_artifacts_event_artifact_unique UNIQUE (event_id, artifact_id)
);

-- Migrar dados de volta
INSERT INTO event_characters (id, tenant_id, event_id, character_id, role, created_at)
SELECT id, tenant_id, event_id, entity_id, relationship_type, created_at
FROM event_references
WHERE entity_type = 'character';

INSERT INTO event_locations (id, tenant_id, event_id, location_id, significance, created_at)
SELECT id, tenant_id, event_id, entity_id, relationship_type, created_at
FROM event_references
WHERE entity_type = 'location';

INSERT INTO event_artifacts (id, tenant_id, event_id, artifact_id, role, created_at)
SELECT id, tenant_id, event_id, entity_id, relationship_type, created_at
FROM event_references
WHERE entity_type = 'artifact';

-- Remover tabela nova
DROP TABLE IF EXISTS event_references;

