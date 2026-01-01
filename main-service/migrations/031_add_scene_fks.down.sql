-- Remove foreign key constraints for pov_character_id and location_id
ALTER TABLE scenes DROP CONSTRAINT IF EXISTS scenes_location_fk;
ALTER TABLE scenes DROP CONSTRAINT IF EXISTS scenes_pov_character_fk;

