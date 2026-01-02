-- Re-add unique constraints (if needed for rollback)

-- Re-add unique constraint for scenes (only for non-null chapter_id)
CREATE UNIQUE INDEX scenes_chapter_order_unique ON scenes(chapter_id, order_num) 
    WHERE chapter_id IS NOT NULL;

-- Re-add unique constraint for beats
ALTER TABLE beats ADD CONSTRAINT beats_scene_order_unique UNIQUE (scene_id, order_num);


