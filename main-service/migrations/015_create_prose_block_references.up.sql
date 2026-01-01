CREATE TABLE IF NOT EXISTS prose_block_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prose_block_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT prose_block_references_prose_block_fk FOREIGN KEY (prose_block_id) REFERENCES prose_blocks(id) ON DELETE CASCADE,
    CONSTRAINT prose_block_references_entity_type_check CHECK (entity_type IN ('scene', 'beat', 'character', 'location', 'trait')),
    CONSTRAINT prose_block_references_unique UNIQUE (prose_block_id, entity_type, entity_id)
);

CREATE INDEX idx_prose_block_references_prose_block_id ON prose_block_references(prose_block_id);
CREATE INDEX idx_prose_block_references_entity ON prose_block_references(entity_type, entity_id);
CREATE INDEX idx_prose_block_references_entity_id ON prose_block_references(entity_id);

