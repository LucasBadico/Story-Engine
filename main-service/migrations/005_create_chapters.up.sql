CREATE TABLE IF NOT EXISTS chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL,
    number INT NOT NULL,
    title VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chapters_story_fk FOREIGN KEY (story_id) REFERENCES stories(id) ON DELETE CASCADE,
    CONSTRAINT chapters_story_number_unique UNIQUE (story_id, number),
    CONSTRAINT chapters_status_check CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT chapters_number_positive CHECK (number > 0)
);

CREATE INDEX idx_chapters_story_id ON chapters(story_id);
CREATE INDEX idx_chapters_story_number ON chapters(story_id, number);

