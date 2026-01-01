ALTER TABLE stories ADD COLUMN world_id UUID REFERENCES worlds(id) ON DELETE SET NULL;

CREATE INDEX idx_stories_world_id ON stories(world_id);

