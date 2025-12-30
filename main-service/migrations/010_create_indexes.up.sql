-- Composite indexes for common query patterns
CREATE INDEX idx_stories_tenant_status ON stories(tenant_id, status);
CREATE INDEX idx_stories_tenant_created ON stories(tenant_id, created_at DESC);

-- Version graph navigation
CREATE INDEX idx_stories_root_created ON stories(root_story_id, created_at);

-- Chapter navigation
CREATE INDEX idx_chapters_story_status ON chapters(story_id, status);

-- Full hierarchy queries
CREATE INDEX idx_scenes_story_chapter ON scenes(story_id, chapter_id);

