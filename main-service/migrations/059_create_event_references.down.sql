-- Recriar tabelas antigas
CREATE TABLE IF NOT EXISTS event_characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    role VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, character_id)
);

CREATE TABLE IF NOT EXISTS event_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    significance TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, location_id)
);

CREATE TABLE IF NOT EXISTS event_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    role VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, artifact_id)
);

-- Migrar dados de volta
INSERT INTO event_characters (id, tenant_id, event_id, character_id, role, created_at)
SELECT id, tenant_id, event_id, entity_id, relationship_type, created_at
FROM event_references
WHERE entity_type = 'character'
ON CONFLICT DO NOTHING;

INSERT INTO event_locations (id, tenant_id, event_id, location_id, significance, created_at)
SELECT id, tenant_id, event_id, entity_id, relationship_type, created_at
FROM event_references
WHERE entity_type = 'location'
ON CONFLICT DO NOTHING;

INSERT INTO event_artifacts (id, tenant_id, event_id, artifact_id, role, created_at)
SELECT id, tenant_id, event_id, entity_id, relationship_type, created_at
FROM event_references
WHERE entity_type = 'artifact'
ON CONFLICT DO NOTHING;

-- Remover tabela nova
DROP TABLE IF EXISTS event_references;

