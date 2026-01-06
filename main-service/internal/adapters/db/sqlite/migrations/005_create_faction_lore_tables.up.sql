-- Create factions table
CREATE TABLE IF NOT EXISTS factions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    parent_id TEXT,
    name TEXT NOT NULL,
    type TEXT,
    description TEXT,
    beliefs TEXT,
    structure TEXT,
    symbols TEXT,
    hierarchy_level INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT factions_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT factions_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT factions_parent_fk FOREIGN KEY (parent_id) REFERENCES factions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_factions_tenant_id ON factions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_factions_world_id ON factions(world_id);
CREATE INDEX IF NOT EXISTS idx_factions_parent_id ON factions(parent_id);
CREATE INDEX IF NOT EXISTS idx_factions_world_parent ON factions(world_id, parent_id);

-- Create faction_references table
CREATE TABLE IF NOT EXISTS faction_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    faction_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    role TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT faction_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT faction_references_faction_fk FOREIGN KEY (faction_id) REFERENCES factions(id) ON DELETE CASCADE,
    CONSTRAINT faction_references_entity_type_check CHECK (entity_type IN (
        'character', 'location', 'artifact', 'event', 
        'lore', 'faction', 
        'lore_reference'
    )),
    CONSTRAINT faction_references_faction_entity_unique UNIQUE (faction_id, entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_faction_references_tenant_id ON faction_references(tenant_id);
CREATE INDEX IF NOT EXISTS idx_faction_references_faction_id ON faction_references(faction_id);
CREATE INDEX IF NOT EXISTS idx_faction_references_entity ON faction_references(entity_type, entity_id);

-- Create lores table
CREATE TABLE IF NOT EXISTS lores (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT NOT NULL,
    parent_id TEXT,
    name TEXT NOT NULL,
    category TEXT,
    description TEXT,
    rules TEXT,
    limitations TEXT,
    requirements TEXT,
    hierarchy_level INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT lores_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT lores_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    CONSTRAINT lores_parent_fk FOREIGN KEY (parent_id) REFERENCES lores(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_lores_tenant_id ON lores(tenant_id);
CREATE INDEX IF NOT EXISTS idx_lores_world_id ON lores(world_id);
CREATE INDEX IF NOT EXISTS idx_lores_parent_id ON lores(parent_id);
CREATE INDEX IF NOT EXISTS idx_lores_world_parent ON lores(world_id, parent_id);

-- Create lore_references table
CREATE TABLE IF NOT EXISTS lore_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    lore_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    relationship_type TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT lore_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT lore_references_lore_fk FOREIGN KEY (lore_id) REFERENCES lores(id) ON DELETE CASCADE,
    CONSTRAINT lore_references_entity_type_check CHECK (entity_type IN (
        'character', 'location', 'artifact', 'event', 
        'faction', 'lore',
        'faction_reference'
    )),
    CONSTRAINT lore_references_lore_entity_unique UNIQUE (lore_id, entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_lore_references_tenant_id ON lore_references(tenant_id);
CREATE INDEX IF NOT EXISTS idx_lore_references_lore_id ON lore_references(lore_id);
CREATE INDEX IF NOT EXISTS idx_lore_references_entity ON lore_references(entity_type, entity_id);

