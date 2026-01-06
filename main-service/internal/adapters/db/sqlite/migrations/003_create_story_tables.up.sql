-- Create stories table (world_id FK references worlds created in migration 002)
CREATE TABLE IF NOT EXISTS stories (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    world_id TEXT,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    version_number INTEGER NOT NULL DEFAULT 1,
    root_story_id TEXT NOT NULL,
    previous_story_id TEXT,
    created_by_user_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT stories_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT stories_root_fk FOREIGN KEY (root_story_id) REFERENCES stories(id) ON DELETE CASCADE,
    CONSTRAINT stories_previous_fk FOREIGN KEY (previous_story_id) REFERENCES stories(id) ON DELETE SET NULL,
    CONSTRAINT stories_world_fk FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE SET NULL,
    CONSTRAINT stories_status_check CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_stories_tenant_id ON stories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stories_root_story_id ON stories(root_story_id);
CREATE INDEX IF NOT EXISTS idx_stories_previous_story_id ON stories(previous_story_id);
CREATE INDEX IF NOT EXISTS idx_stories_status ON stories(status);
CREATE INDEX IF NOT EXISTS idx_stories_created_at ON stories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stories_world_id ON stories(world_id);

-- Create chapters table
CREATE TABLE IF NOT EXISTS chapters (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    story_id TEXT NOT NULL,
    number INTEGER NOT NULL,
    title TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT chapters_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT chapters_story_fk FOREIGN KEY (story_id) REFERENCES stories(id) ON DELETE CASCADE,
    CONSTRAINT chapters_story_number_unique UNIQUE (story_id, number),
    CONSTRAINT chapters_status_check CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT chapters_number_positive CHECK (number > 0)
);

CREATE INDEX IF NOT EXISTS idx_chapters_story_id ON chapters(story_id);
CREATE INDEX IF NOT EXISTS idx_chapters_tenant_id ON chapters(tenant_id);
CREATE INDEX IF NOT EXISTS idx_chapters_story_number ON chapters(story_id, number);

-- Create scenes table (chapter_id is nullable, pov_character_id references characters from migration 002)
CREATE TABLE IF NOT EXISTS scenes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    story_id TEXT NOT NULL,
    chapter_id TEXT,
    order_num INTEGER NOT NULL,
    pov_character_id TEXT,
    time_ref TEXT,
    goal TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT scenes_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT scenes_story_fk FOREIGN KEY (story_id) REFERENCES stories(id) ON DELETE CASCADE,
    CONSTRAINT scenes_chapter_fk FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    CONSTRAINT scenes_pov_character_fk FOREIGN KEY (pov_character_id) REFERENCES characters(id) ON DELETE SET NULL,
    CONSTRAINT scenes_order_positive CHECK (order_num > 0)
);

CREATE INDEX IF NOT EXISTS idx_scenes_story_id ON scenes(story_id);
CREATE INDEX IF NOT EXISTS idx_scenes_tenant_id ON scenes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scenes_chapter_id ON scenes(chapter_id);
CREATE INDEX IF NOT EXISTS idx_scenes_chapter_order ON scenes(chapter_id, order_num) WHERE chapter_id IS NOT NULL;

-- Create beats table
CREATE TABLE IF NOT EXISTS beats (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    scene_id TEXT NOT NULL,
    order_num INTEGER NOT NULL,
    type TEXT NOT NULL,
    intent TEXT,
    outcome TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT beats_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT beats_scene_fk FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE,
    CONSTRAINT beats_scene_order_unique UNIQUE (scene_id, order_num),
    CONSTRAINT beats_type_check CHECK (type IN ('setup', 'turn', 'reveal', 'conflict', 'climax', 'resolution', 'hook', 'transition')),
    CONSTRAINT beats_order_positive CHECK (order_num > 0)
);

CREATE INDEX IF NOT EXISTS idx_beats_scene_id ON beats(scene_id);
CREATE INDEX IF NOT EXISTS idx_beats_tenant_id ON beats(tenant_id);
CREATE INDEX IF NOT EXISTS idx_beats_scene_order ON beats(scene_id, order_num);
CREATE INDEX IF NOT EXISTS idx_beats_type ON beats(type);

-- Create content_blocks table (chapter_id is nullable)
CREATE TABLE IF NOT EXISTS content_blocks (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    chapter_id TEXT,
    kind TEXT NOT NULL DEFAULT 'final',
    content TEXT NOT NULL,
    word_count INTEGER NOT NULL DEFAULT 0,
    type TEXT NOT NULL DEFAULT 'text',
    metadata TEXT NOT NULL DEFAULT '{}',
    order_num INTEGER,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT content_blocks_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT content_blocks_chapter_fk FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    CONSTRAINT content_blocks_kind_check CHECK (kind IN ('final', 'alt_a', 'alt_b', 'cleaned', 'localized', 'draft')),
    CONSTRAINT content_blocks_type_check CHECK (type IN ('text', 'image', 'video', 'audio', 'embed', 'link')),
    CONSTRAINT content_blocks_word_count_positive CHECK (word_count >= 0),
    CONSTRAINT content_blocks_order_positive CHECK (order_num > 0 OR order_num IS NULL)
);

CREATE INDEX IF NOT EXISTS idx_content_blocks_chapter_id ON content_blocks(chapter_id);
CREATE INDEX IF NOT EXISTS idx_content_blocks_tenant_id ON content_blocks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_content_blocks_chapter_order ON content_blocks(chapter_id, order_num) WHERE chapter_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_content_blocks_kind ON content_blocks(kind);
CREATE INDEX IF NOT EXISTS idx_content_blocks_type ON content_blocks(type);

-- Create content_block_references table
CREATE TABLE IF NOT EXISTS content_block_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    content_block_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CONSTRAINT content_block_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT content_block_references_content_block_fk FOREIGN KEY (content_block_id) REFERENCES content_blocks(id) ON DELETE CASCADE,
    CONSTRAINT content_block_references_entity_type_check CHECK (entity_type IN (
        'scene', 'beat', 'chapter', 'character', 'location', 
        'artifact', 'event', 'world',
        'faction', 'lore', 'faction_reference', 'lore_reference'
    )),
    CONSTRAINT content_block_references_unique UNIQUE (content_block_id, entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_content_block_references_content_block_id ON content_block_references(content_block_id);
CREATE INDEX IF NOT EXISTS idx_content_block_references_tenant_id ON content_block_references(tenant_id);
CREATE INDEX IF NOT EXISTS idx_content_block_references_entity ON content_block_references(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_content_block_references_entity_id ON content_block_references(entity_id);

