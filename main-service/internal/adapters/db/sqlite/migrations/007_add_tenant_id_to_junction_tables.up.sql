-- Add tenant_id column to archetype_traits table
-- SQLite doesn't support ADD COLUMN with constraints, so we need to recreate the table

-- Step 1: Create new archetype_traits table with tenant_id
CREATE TABLE IF NOT EXISTS archetype_traits_new (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    archetype_id TEXT NOT NULL,
    trait_id TEXT NOT NULL,
    default_value TEXT,
    created_at TEXT NOT NULL,
    CONSTRAINT archetype_traits_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT archetype_traits_archetype_fk FOREIGN KEY (archetype_id) REFERENCES archetypes(id) ON DELETE CASCADE,
    CONSTRAINT archetype_traits_trait_fk FOREIGN KEY (trait_id) REFERENCES traits(id) ON DELETE CASCADE,
    CONSTRAINT archetype_traits_archetype_trait_unique UNIQUE (archetype_id, trait_id)
);

-- Step 2: Copy existing data (get tenant_id from archetype)
INSERT INTO archetype_traits_new (id, tenant_id, archetype_id, trait_id, default_value, created_at)
SELECT at.id, a.tenant_id, at.archetype_id, at.trait_id, at.default_value, at.created_at
FROM archetype_traits at
JOIN archetypes a ON at.archetype_id = a.id;

-- Step 3: Drop old table
DROP TABLE archetype_traits;

-- Step 4: Rename new table
ALTER TABLE archetype_traits_new RENAME TO archetype_traits;

-- Step 5: Recreate indexes
CREATE INDEX idx_archetype_traits_tenant_id ON archetype_traits(tenant_id);
CREATE INDEX idx_archetype_traits_archetype_id ON archetype_traits(archetype_id);
CREATE INDEX idx_archetype_traits_trait_id ON archetype_traits(trait_id);

-- Add tenant_id column to scene_references table

-- Step 1: Create new scene_references table with tenant_id
CREATE TABLE IF NOT EXISTS scene_references_new (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    scene_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CONSTRAINT scene_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT scene_references_scene_fk FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE,
    CONSTRAINT scene_references_entity_type_check CHECK (entity_type IN ('character', 'location', 'artifact')),
    CONSTRAINT scene_references_unique UNIQUE (scene_id, entity_type, entity_id)
);

-- Step 2: Copy existing data (get tenant_id from scene)
INSERT INTO scene_references_new (id, tenant_id, scene_id, entity_type, entity_id, created_at)
SELECT sr.id, s.tenant_id, sr.scene_id, sr.entity_type, sr.entity_id, sr.created_at
FROM scene_references sr
JOIN scenes s ON sr.scene_id = s.id;

-- Step 3: Drop old table
DROP TABLE scene_references;

-- Step 4: Rename new table
ALTER TABLE scene_references_new RENAME TO scene_references;

-- Step 5: Recreate indexes
CREATE INDEX idx_scene_references_tenant_id ON scene_references(tenant_id);
CREATE INDEX idx_scene_references_scene_id ON scene_references(scene_id);
CREATE INDEX idx_scene_references_entity ON scene_references(entity_type, entity_id);

