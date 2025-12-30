CREATE TABLE IF NOT EXISTS stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    version_number INT NOT NULL DEFAULT 1,
    root_story_id UUID NOT NULL,
    previous_story_id UUID,
    created_by_user_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT stories_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT stories_root_fk FOREIGN KEY (root_story_id) REFERENCES stories(id) ON DELETE CASCADE,
    CONSTRAINT stories_previous_fk FOREIGN KEY (previous_story_id) REFERENCES stories(id) ON DELETE SET NULL,
    CONSTRAINT stories_created_by_fk FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT stories_status_check CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE INDEX idx_stories_tenant_id ON stories(tenant_id);
CREATE INDEX idx_stories_root_story_id ON stories(root_story_id);
CREATE INDEX idx_stories_previous_story_id ON stories(previous_story_id);
CREATE INDEX idx_stories_status ON stories(status);
CREATE INDEX idx_stories_created_at ON stories(created_at DESC);

