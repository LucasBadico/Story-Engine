CREATE TABLE IF NOT EXISTS scene_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id UUID NOT NULL REFERENCES scenes(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scene_id, entity_type, entity_id),
    CONSTRAINT scene_references_entity_type_check CHECK (entity_type IN ('character', 'location', 'artifact'))
);

CREATE INDEX idx_scene_references_scene_id ON scene_references(scene_id);
CREATE INDEX idx_scene_references_entity ON scene_references(entity_type, entity_id);

