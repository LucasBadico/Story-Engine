CREATE TABLE IF NOT EXISTS image_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID REFERENCES chapters(id) ON DELETE SET NULL,
    order_num INT,
    kind VARCHAR(50) NOT NULL DEFAULT 'final',
    image_url TEXT NOT NULL,
    alt_text TEXT,
    caption TEXT,
    width INT,
    height INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT image_blocks_kind_check CHECK (kind IN ('final', 'alt_a', 'alt_b', 'draft', 'thumbnail')),
    CONSTRAINT image_blocks_order_positive CHECK (order_num IS NULL OR order_num > 0),
    CONSTRAINT image_blocks_dimensions_positive CHECK ((width IS NULL OR width > 0) AND (height IS NULL OR height > 0))
);

CREATE INDEX idx_image_blocks_chapter_id ON image_blocks(chapter_id);
CREATE INDEX idx_image_blocks_kind ON image_blocks(kind);


