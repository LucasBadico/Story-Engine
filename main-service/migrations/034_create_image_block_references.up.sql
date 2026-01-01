CREATE TABLE IF NOT EXISTS image_block_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_block_id UUID NOT NULL REFERENCES image_blocks(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT image_block_references_entity_type_check CHECK (entity_type IN ('scene', 'beat', 'chapter', 'character', 'location', 'artifact', 'event', 'world')),
    CONSTRAINT image_block_references_unique UNIQUE (image_block_id, entity_type, entity_id)
);

CREATE INDEX idx_image_block_references_image_block_id ON image_block_references(image_block_id);
CREATE INDEX idx_image_block_references_entity ON image_block_references(entity_type, entity_id);
CREATE INDEX idx_image_block_references_entity_id ON image_block_references(entity_id);

