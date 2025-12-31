-- Revert stories_root_fk constraint to non-deferrable
ALTER TABLE stories DROP CONSTRAINT stories_root_fk;
ALTER TABLE stories ADD CONSTRAINT stories_root_fk 
    FOREIGN KEY (root_story_id) REFERENCES stories(id) 
    ON DELETE CASCADE;

