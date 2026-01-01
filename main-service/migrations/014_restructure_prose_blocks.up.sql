-- Remove foreign key constraint and index for scene_id
ALTER TABLE prose_blocks DROP CONSTRAINT IF EXISTS prose_blocks_scene_fk;
DROP INDEX IF EXISTS idx_prose_blocks_scene_id;

-- Add chapter_id column
ALTER TABLE prose_blocks ADD COLUMN chapter_id UUID;

-- Migrate data: try to find chapter_id from scenes
-- Note: This assumes scenes have chapter_id. If scene doesn't have chapter, set to NULL
UPDATE prose_blocks pb
SET chapter_id = s.chapter_id
FROM scenes s
WHERE pb.scene_id = s.id;

-- Make chapter_id NOT NULL after migration
ALTER TABLE prose_blocks ALTER COLUMN chapter_id SET NOT NULL;

-- Add order_num column
ALTER TABLE prose_blocks ADD COLUMN order_num INT;

-- Set initial order_num based on created_at (for existing data)
WITH ordered_blocks AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY chapter_id ORDER BY created_at ASC) as rn
    FROM prose_blocks
)
UPDATE prose_blocks pb
SET order_num = ob.rn
FROM ordered_blocks ob
WHERE pb.id = ob.id;

-- Make order_num NOT NULL and add constraints
ALTER TABLE prose_blocks ALTER COLUMN order_num SET NOT NULL;
ALTER TABLE prose_blocks ADD CONSTRAINT prose_blocks_order_positive CHECK (order_num > 0);

-- Add foreign key for chapter_id
ALTER TABLE prose_blocks ADD CONSTRAINT prose_blocks_chapter_fk FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX idx_prose_blocks_chapter_id ON prose_blocks(chapter_id);
CREATE INDEX idx_prose_blocks_chapter_order ON prose_blocks(chapter_id, order_num);

-- Drop old scene_id column
ALTER TABLE prose_blocks DROP COLUMN scene_id;

