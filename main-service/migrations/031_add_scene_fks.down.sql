-- Remove foreign key constraint for pov_character_id
ALTER TABLE scenes DROP CONSTRAINT IF EXISTS scenes_pov_character_fk;

-- Restore location_id column (if needed for rollback)
ALTER TABLE scenes ADD COLUMN IF NOT EXISTS location_id UUID;

