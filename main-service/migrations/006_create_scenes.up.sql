CREATE TABLE IF NOT EXISTS scenes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL,
    chapter_id UUID NOT NULL,
    order_num INT NOT NULL,
    pov_character_id UUID,
    location_id UUID,
    time_ref TEXT,
    goal TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT scenes_story_fk FOREIGN KEY (story_id) REFERENCES stories(id) ON DELETE CASCADE,
    CONSTRAINT scenes_chapter_fk FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    CONSTRAINT scenes_chapter_order_unique UNIQUE (chapter_id, order_num),
    CONSTRAINT scenes_order_positive CHECK (order_num > 0)
);

CREATE INDEX idx_scenes_story_id ON scenes(story_id);
CREATE INDEX idx_scenes_chapter_id ON scenes(chapter_id);
CREATE INDEX idx_scenes_chapter_order ON scenes(chapter_id, order_num);

