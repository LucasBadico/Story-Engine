-- Add scene_id column back
ALTER TABLE prose_blocks ADD COLUMN scene_id UUID;

-- Drop chapter_id foreign key and indexes
ALTER TABLE prose_blocks DROP CONSTRAINT IF EXISTS prose_blocks_chapter_fk;
DROP INDEX IF EXISTS idx_prose_blocks_chapter_id;
DROP INDEX IF EXISTS idx_prose_blocks_chapter_order;

-- Drop order_num column
ALTER TABLE prose_blocks DROP COLUMN IF EXISTS order_num;

-- Make scene_id NOT NULL
ALTER TABLE prose_blocks ALTER COLUMN scene_id SET NOT NULL;

-- Add back foreign key for scene_id
ALTER TABLE prose_blocks ADD CONSTRAINT prose_blocks_scene_fk FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE;

-- Create index for scene_id
CREATE INDEX idx_prose_blocks_scene_id ON prose_blocks(scene_id);

-- Drop chapter_id column
ALTER TABLE prose_blocks DROP COLUMN chapter_id;


