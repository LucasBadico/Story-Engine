-- Create worlds table (rpg_system_id without FK constraint)
CREATE TABLE IF NOT EXISTS worlds (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    rpg_system_id TEXT,
    name TEXT NOT NULL,
    description TEXT,
    genre TEXT,
    is_implicit INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT worlds_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_worlds_tenant_id ON worlds(tenant_id);
CREATE INDEX idx_worlds_implicit ON worlds(tenant_id, is_implicit);
CREATE INDEX idx_worlds_rpg_system_id ON worlds(rpg_system_id);

-- Create locations table
CREATE TABLE IF NOT EXISTS locations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    parent_id TEXT,
    name TEXT NOT NULL,
    type TEXT,
    description TEXT,
    hierarchy_level INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT locations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT locations_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT locations_parent_fk FOREIGN KEY (parent_id) REFERENCES locations(id) ON DELETE CASCADE
);

CREATE INDEX idx_locations_world_id ON locations(world_id);
CREATE INDEX idx_locations_tenant_id ON locations(tenant_id);
CREATE INDEX idx_locations_parent_id ON locations(parent_id);
CREATE INDEX idx_locations_world_parent ON locations(world_id, parent_id);

-- Create traits table
CREATE TABLE IF NOT EXISTS traits (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT,
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT traits_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT traits_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX idx_traits_tenant_id ON traits(tenant_id);
CREATE INDEX idx_traits_category ON traits(tenant_id, category);

-- Create archetypes table
CREATE TABLE IF NOT EXISTS archetypes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT archetypes_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT archetypes_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX idx_archetypes_tenant_id ON archetypes(tenant_id);

-- Create characters table (current_class_id without FK constraint)
CREATE TABLE IF NOT EXISTS characters (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    archetype_id TEXT,
    current_class_id TEXT,
    class_level INTEGER DEFAULT 1,
    name TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT characters_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT characters_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT characters_archetype_fk FOREIGN KEY (archetype_id) REFERENCES archetypes(id) ON DELETE SET NULL
);

CREATE INDEX idx_characters_world_id ON characters(world_id);
CREATE INDEX idx_characters_tenant_id ON characters(tenant_id);
CREATE INDEX idx_characters_archetype_id ON characters(archetype_id);
CREATE INDEX idx_characters_current_class ON characters(current_class_id);

-- Create character_traits table
CREATE TABLE IF NOT EXISTS character_traits (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    trait_id TEXT NOT NULL,
    trait_name TEXT NOT NULL,
    trait_category TEXT,
    trait_description TEXT,
    value TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT character_traits_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT character_traits_character_fk FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT character_traits_trait_fk FOREIGN KEY (trait_id) REFERENCES traits(id) ON DELETE CASCADE,
    CONSTRAINT character_traits_character_trait_unique UNIQUE (character_id, trait_id)
);

CREATE INDEX idx_character_traits_character_id ON character_traits(character_id);
CREATE INDEX idx_character_traits_trait_id ON character_traits(trait_id);
CREATE INDEX idx_character_traits_tenant_id ON character_traits(tenant_id);

-- Create archetype_traits table
CREATE TABLE IF NOT EXISTS archetype_traits (
    id TEXT PRIMARY KEY,
    archetype_id TEXT NOT NULL,
    trait_id TEXT NOT NULL,
    default_value TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT archetype_traits_archetype_fk FOREIGN KEY (archetype_id) REFERENCES archetypes(id) ON DELETE CASCADE,
    CONSTRAINT archetype_traits_trait_fk FOREIGN KEY (trait_id) REFERENCES traits(id) ON DELETE CASCADE,
    CONSTRAINT archetype_traits_archetype_trait_unique UNIQUE (archetype_id, trait_id)
);

CREATE INDEX idx_archetype_traits_archetype_id ON archetype_traits(archetype_id);
CREATE INDEX idx_archetype_traits_trait_id ON archetype_traits(trait_id);

-- Create artifacts table
CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    rarity TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT artifacts_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT artifacts_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE
);

CREATE INDEX idx_artifacts_world_id ON artifacts(world_id);
CREATE INDEX idx_artifacts_tenant_id ON artifacts(tenant_id);

-- Create artifact_references table
CREATE TABLE IF NOT EXISTS artifact_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    artifact_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CONSTRAINT artifact_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT artifact_references_artifact_fk FOREIGN KEY (artifact_id) REFERENCES artifacts(id) ON DELETE CASCADE,
    CONSTRAINT artifact_references_entity_type_check CHECK (entity_type IN ('character', 'location')),
    CONSTRAINT artifact_references_artifact_entity_unique UNIQUE (artifact_id, entity_type, entity_id)
);

CREATE INDEX idx_artifact_references_artifact_id ON artifact_references(artifact_id);
CREATE INDEX idx_artifact_references_tenant_id ON artifact_references(tenant_id);
CREATE INDEX idx_artifact_references_entity ON artifact_references(entity_type, entity_id);

