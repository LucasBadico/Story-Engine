-- Add foreign key constraint for pov_character_id in scenes table
ALTER TABLE scenes 
    ADD CONSTRAINT scenes_pov_character_fk 
    FOREIGN KEY (pov_character_id) REFERENCES characters(id) ON DELETE SET NULL;

-- Remove location_id column from scenes table (using scene_references for dynamic relationships)
ALTER TABLE scenes DROP COLUMN IF EXISTS location_id;

