-- Remove unique constraints on order_num to allow duplicates
-- This allows more flexibility in organizing items

-- Remove unique constraint from scenes
DROP INDEX IF EXISTS scenes_chapter_order_unique;

-- Remove unique constraint from beats
ALTER TABLE beats DROP CONSTRAINT IF EXISTS beats_scene_order_unique;


