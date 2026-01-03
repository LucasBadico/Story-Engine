-- Unify prose_blocks and image_blocks into content_blocks
-- This migration renames prose_blocks to content_blocks, adds type and metadata columns,
-- and drops image_blocks tables (dev mode - no data migration needed)

-- ============================================
-- Rename prose_blocks to content_blocks
-- ============================================

ALTER TABLE prose_blocks RENAME TO content_blocks;

-- Update foreign key constraint name
ALTER TABLE content_blocks RENAME CONSTRAINT prose_blocks_chapter_fk TO content_blocks_chapter_fk;
ALTER TABLE content_blocks RENAME CONSTRAINT prose_blocks_kind_check TO content_blocks_kind_check;
ALTER TABLE content_blocks RENAME CONSTRAINT prose_blocks_order_positive TO content_blocks_order_positive;
ALTER TABLE content_blocks RENAME CONSTRAINT prose_blocks_word_count_positive TO content_blocks_word_count_positive;

-- Update index names
ALTER INDEX idx_prose_blocks_chapter_id RENAME TO idx_content_blocks_chapter_id;
ALTER INDEX idx_prose_blocks_chapter_order RENAME TO idx_content_blocks_chapter_order;
ALTER INDEX idx_prose_blocks_kind RENAME TO idx_content_blocks_kind;
ALTER INDEX idx_prose_blocks_tenant_id RENAME TO idx_content_blocks_tenant_id;

-- Update tenant foreign key constraint name
ALTER TABLE content_blocks RENAME CONSTRAINT prose_blocks_tenant_fk TO content_blocks_tenant_fk;

-- ============================================
-- Add new columns: type and metadata
-- ============================================

ALTER TABLE content_blocks ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'text';
ALTER TABLE content_blocks ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}';

-- Add constraint for type values
ALTER TABLE content_blocks ADD CONSTRAINT content_blocks_type_check 
    CHECK (type IN ('text', 'image', 'video', 'audio', 'embed', 'link'));

-- Create index for type
CREATE INDEX idx_content_blocks_type ON content_blocks(type);

-- Create GIN index for metadata JSONB queries
CREATE INDEX idx_content_blocks_metadata ON content_blocks USING GIN (metadata);

-- ============================================
-- Rename prose_block_references to content_block_references
-- ============================================

ALTER TABLE prose_block_references RENAME TO content_block_references;

-- Update foreign key constraint name and column reference
ALTER TABLE content_block_references RENAME CONSTRAINT prose_block_references_prose_block_fk TO content_block_references_content_block_fk;
ALTER TABLE content_block_references RENAME COLUMN prose_block_id TO content_block_id;

-- Update foreign key to reference new table name
ALTER TABLE content_block_references DROP CONSTRAINT content_block_references_content_block_fk;
ALTER TABLE content_block_references ADD CONSTRAINT content_block_references_content_block_fk 
    FOREIGN KEY (content_block_id) REFERENCES content_blocks(id) ON DELETE CASCADE;

-- Update constraint names
ALTER TABLE content_block_references RENAME CONSTRAINT prose_block_references_entity_type_check TO content_block_references_entity_type_check;
ALTER TABLE content_block_references RENAME CONSTRAINT prose_block_references_unique TO content_block_references_unique;

-- Update index names
ALTER INDEX idx_prose_block_references_prose_block_id RENAME TO idx_content_block_references_content_block_id;
ALTER INDEX idx_prose_block_references_entity RENAME TO idx_content_block_references_entity;
ALTER INDEX idx_prose_block_references_entity_id RENAME TO idx_content_block_references_entity_id;
ALTER INDEX idx_prose_block_references_tenant_id RENAME TO idx_content_block_references_tenant_id;

-- Update tenant foreign key constraint name (if exists)
ALTER TABLE content_block_references RENAME CONSTRAINT prose_block_references_tenant_fk TO content_block_references_tenant_fk;

-- ============================================
-- Drop image_blocks tables (dev mode - no data migration)
-- ============================================

DROP TABLE IF EXISTS image_block_references;
DROP TABLE IF EXISTS image_blocks;

