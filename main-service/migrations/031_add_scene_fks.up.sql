-- Add foreign key constraints for pov_character_id and location_id in scenes table
ALTER TABLE scenes 
    ADD CONSTRAINT scenes_pov_character_fk 
    FOREIGN KEY (pov_character_id) REFERENCES characters(id) ON DELETE SET NULL;

ALTER TABLE scenes 
    ADD CONSTRAINT scenes_location_fk 
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL;

