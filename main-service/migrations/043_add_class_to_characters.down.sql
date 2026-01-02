DROP INDEX IF EXISTS idx_characters_current_class;
ALTER TABLE characters DROP COLUMN IF EXISTS class_level;
ALTER TABLE characters DROP COLUMN IF EXISTS current_class_id;


