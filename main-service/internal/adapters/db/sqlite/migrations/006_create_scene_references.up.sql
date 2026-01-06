-- Create scene_references table (created after scenes, characters, locations, artifacts)
CREATE TABLE IF NOT EXISTS scene_references (
    id TEXT PRIMARY KEY,
    scene_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CONSTRAINT scene_references_scene_fk FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE,
    CONSTRAINT scene_references_entity_type_check CHECK (entity_type IN ('character', 'location', 'artifact')),
    CONSTRAINT scene_references_unique UNIQUE (scene_id, entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_scene_references_scene_id ON scene_references(scene_id);
CREATE INDEX IF NOT EXISTS idx_scene_references_entity ON scene_references(entity_type, entity_id);

