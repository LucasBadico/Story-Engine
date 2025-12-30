CREATE TABLE IF NOT EXISTS prose_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id UUID NOT NULL,
    kind VARCHAR(50) NOT NULL DEFAULT 'final',
    content TEXT NOT NULL,
    word_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT prose_blocks_scene_fk FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE,
    CONSTRAINT prose_blocks_kind_check CHECK (kind IN ('final', 'alt_a', 'alt_b', 'cleaned', 'localized', 'draft')),
    CONSTRAINT prose_blocks_word_count_positive CHECK (word_count >= 0)
);

CREATE INDEX idx_prose_blocks_scene_id ON prose_blocks(scene_id);
CREATE INDEX idx_prose_blocks_kind ON prose_blocks(kind);

