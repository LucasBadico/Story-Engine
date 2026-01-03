-- Rollback: Revert content_blocks back to prose_blocks and recreate image_blocks

-- ============================================
-- Recreate image_blocks tables
-- ============================================

CREATE TABLE IF NOT EXISTS image_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
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
CREATE INDEX idx_image_blocks_tenant_id ON image_blocks(tenant_id);

CREATE TABLE IF NOT EXISTS image_block_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    image_block_id UUID NOT NULL REFERENCES image_blocks(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT image_block_references_entity_type_check CHECK (entity_type IN ('scene', 'beat', 'chapter', 'character', 'location', 'artifact', 'event', 'world', 'rpg_system', 'rpg_skill', 'rpg_class', 'inventory_item')),
    CONSTRAINT image_block_references_unique UNIQUE (image_block_id, entity_type, entity_id)
);

CREATE INDEX idx_image_block_references_image_block_id ON image_block_references(image_block_id);
CREATE INDEX idx_image_block_references_entity ON image_block_references(entity_type, entity_id);
CREATE INDEX idx_image_block_references_entity_id ON image_block_references(entity_id);
CREATE INDEX idx_image_block_references_tenant_id ON image_block_references(tenant_id);

-- ============================================
-- Revert content_block_references to prose_block_references
-- ============================================

ALTER TABLE content_block_references RENAME TO prose_block_references;
ALTER TABLE prose_block_references RENAME COLUMN content_block_id TO prose_block_id;

ALTER TABLE prose_block_references RENAME CONSTRAINT content_block_references_content_block_fk TO prose_block_references_prose_block_fk;
ALTER TABLE prose_block_references RENAME CONSTRAINT content_block_references_entity_type_check TO prose_block_references_entity_type_check;
ALTER TABLE prose_block_references RENAME CONSTRAINT content_block_references_unique TO prose_block_references_unique;

ALTER INDEX idx_content_block_references_content_block_id RENAME TO idx_prose_block_references_prose_block_id;
ALTER INDEX idx_content_block_references_entity RENAME TO idx_prose_block_references_entity;
ALTER INDEX idx_content_block_references_entity_id RENAME TO idx_prose_block_references_entity_id;
ALTER INDEX idx_content_block_references_tenant_id RENAME TO idx_prose_block_references_tenant_id;

-- Revert tenant foreign key constraint name (if exists)
ALTER TABLE prose_block_references RENAME CONSTRAINT content_block_references_tenant_fk TO prose_block_references_tenant_fk;

-- Update foreign key to reference old table name
ALTER TABLE prose_block_references DROP CONSTRAINT prose_block_references_prose_block_fk;
ALTER TABLE prose_block_references ADD CONSTRAINT prose_block_references_prose_block_fk 
    FOREIGN KEY (prose_block_id) REFERENCES prose_blocks(id) ON DELETE CASCADE;

-- ============================================
-- Revert content_blocks to prose_blocks
-- ============================================

-- Drop new columns
ALTER TABLE content_blocks DROP COLUMN IF EXISTS metadata;
ALTER TABLE content_blocks DROP COLUMN IF EXISTS type;

-- Drop new indexes
DROP INDEX IF EXISTS idx_content_blocks_metadata;
DROP INDEX IF EXISTS idx_content_blocks_type;

-- Rename table
ALTER TABLE content_blocks RENAME TO prose_blocks;

-- Revert constraint names
ALTER TABLE prose_blocks RENAME CONSTRAINT content_blocks_chapter_fk TO prose_blocks_chapter_fk;
ALTER TABLE prose_blocks RENAME CONSTRAINT content_blocks_kind_check TO prose_blocks_kind_check;
ALTER TABLE prose_blocks RENAME CONSTRAINT content_blocks_order_positive TO prose_blocks_order_positive;
ALTER TABLE prose_blocks RENAME CONSTRAINT content_blocks_word_count_positive TO prose_blocks_word_count_positive;

-- Revert index names
ALTER INDEX idx_content_blocks_chapter_id RENAME TO idx_prose_blocks_chapter_id;
ALTER INDEX idx_content_blocks_chapter_order RENAME TO idx_prose_blocks_chapter_order;
ALTER INDEX idx_content_blocks_kind RENAME TO idx_prose_blocks_kind;
ALTER INDEX idx_content_blocks_tenant_id RENAME TO idx_prose_blocks_tenant_id;

-- Revert tenant foreign key constraint name
ALTER TABLE prose_blocks RENAME CONSTRAINT content_blocks_tenant_fk TO prose_blocks_tenant_fk;

