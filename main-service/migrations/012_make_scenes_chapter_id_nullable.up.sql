-- Make chapter_id nullable in scenes table
ALTER TABLE scenes DROP CONSTRAINT scenes_chapter_fk;
ALTER TABLE scenes DROP CONSTRAINT scenes_chapter_order_unique;

-- Make chapter_id nullable
ALTER TABLE scenes ALTER COLUMN chapter_id DROP NOT NULL;

-- Re-add foreign key constraint (allowing NULL)
ALTER TABLE scenes ADD CONSTRAINT scenes_chapter_fk 
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) 
    ON DELETE CASCADE;

-- Re-add unique constraint (only for non-null chapter_id)
CREATE UNIQUE INDEX scenes_chapter_order_unique ON scenes(chapter_id, order_num) 
    WHERE chapter_id IS NOT NULL;


