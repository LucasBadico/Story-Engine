-- Revert chapter_id to NOT NULL
DROP INDEX IF EXISTS scenes_chapter_order_unique;

ALTER TABLE scenes DROP CONSTRAINT scenes_chapter_fk;

-- Remove NULL values before making NOT NULL
UPDATE scenes SET chapter_id = (SELECT id FROM chapters LIMIT 1) WHERE chapter_id IS NULL;

ALTER TABLE scenes ALTER COLUMN chapter_id SET NOT NULL;

ALTER TABLE scenes ADD CONSTRAINT scenes_chapter_fk 
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) 
    ON DELETE CASCADE;

ALTER TABLE scenes ADD CONSTRAINT scenes_chapter_order_unique UNIQUE (chapter_id, order_num);


