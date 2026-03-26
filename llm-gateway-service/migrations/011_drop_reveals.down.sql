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
