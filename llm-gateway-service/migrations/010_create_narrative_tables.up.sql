-- Concepts: entity identity (character, location, artefact, faction, event)
CREATE TABLE IF NOT EXISTS concepts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  story_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(50) NOT NULL,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, story_id, type, name)
);

CREATE INDEX IF NOT EXISTS idx_concepts_story ON concepts (story_id, type);
CREATE INDEX IF NOT EXISTS idx_concepts_tenant ON concepts (tenant_id);

-- Concept links: relationship graph between concepts
CREATE TABLE IF NOT EXISTS concept_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  source_concept_id UUID NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
  target_concept_id UUID NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
  relation_type VARCHAR(100) NOT NULL,
  summary TEXT,
  metadata JSONB DEFAULT '{}',
  valid_from_chapter INT NOT NULL DEFAULT 0,
  valid_to_chapter INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_concept_links_source ON concept_links (source_concept_id);
CREATE INDEX IF NOT EXISTS idx_concept_links_target ON concept_links (target_concept_id);
CREATE INDEX IF NOT EXISTS idx_concept_links_tenant ON concept_links (tenant_id);

-- State snapshots: per-chapter entity state (hot layer)
CREATE TABLE IF NOT EXISTS state_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  story_id UUID NOT NULL,
  entity_id UUID NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  chapter_number INT NOT NULL,
  timeline_day INT,
  timeline_confidence VARCHAR(16),
  state_data JSONB NOT NULL DEFAULT '{}',
  scene_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_snapshots_entity ON state_snapshots (story_id, entity_id, chapter_number DESC);
CREATE INDEX IF NOT EXISTS idx_state_snapshots_story ON state_snapshots (story_id, chapter_number DESC);
CREATE INDEX IF NOT EXISTS idx_state_snapshots_timeline ON state_snapshots (story_id, timeline_day DESC) WHERE timeline_day IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_state_snapshots_tenant ON state_snapshots (tenant_id);

-- Events: atomic narrative actions in the timeline
CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  story_id UUID NOT NULL,
  chapter_number INT NOT NULL,
  timeline_day INT,
  timeline_confidence VARCHAR(16),
  scene_id UUID,
  description TEXT NOT NULL,
  entities_involved JSONB DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_story ON events (story_id, chapter_number);
CREATE INDEX IF NOT EXISTS idx_events_timeline ON events (story_id, timeline_day) WHERE timeline_day IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_events_tenant ON events (tenant_id);

-- Reveals: what the reader learns per chapter
CREATE TABLE IF NOT EXISTS reveals (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  story_id UUID NOT NULL,
  chapter_number INT NOT NULL,
  fact TEXT NOT NULL,
  entity_ids JSONB DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reveals_story ON reveals (story_id, chapter_number);
CREATE INDEX IF NOT EXISTS idx_reveals_tenant ON reveals (tenant_id);
