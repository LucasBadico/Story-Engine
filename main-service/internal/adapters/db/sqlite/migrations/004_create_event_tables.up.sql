-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT,
    description TEXT,
    timeline TEXT,
    importance INTEGER DEFAULT 5,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT events_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT events_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE
);

CREATE INDEX idx_events_world_id ON events(world_id);
CREATE INDEX idx_events_tenant_id ON events(tenant_id);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_timeline ON events(timeline);

-- Create event_characters table
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

CREATE INDEX idx_event_characters_event_id ON event_characters(event_id);
CREATE INDEX idx_event_characters_character_id ON event_characters(character_id);
CREATE INDEX idx_event_characters_tenant_id ON event_characters(tenant_id);

-- Create event_locations table
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

CREATE INDEX idx_event_locations_event_id ON event_locations(event_id);
CREATE INDEX idx_event_locations_location_id ON event_locations(location_id);
CREATE INDEX idx_event_locations_tenant_id ON event_locations(tenant_id);

-- Create event_artifacts table
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

CREATE INDEX idx_event_artifacts_event_id ON event_artifacts(event_id);
CREATE INDEX idx_event_artifacts_artifact_id ON event_artifacts(artifact_id);
CREATE INDEX idx_event_artifacts_tenant_id ON event_artifacts(tenant_id);

