CREATE TABLE IF NOT EXISTS character_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    chapter_id UUID REFERENCES chapters(id) ON DELETE SET NULL,
    chapter_number INT NOT NULL,
    emotion TEXT,
    goal TEXT,
    secrets TEXT[],
    clothes TEXT,
    physical_state TEXT,
    timeline_day INT,
    timeline_confidence VARCHAR(16),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_character_states_character ON character_states(character_id, chapter_number DESC);
CREATE INDEX idx_character_states_chapter ON character_states(chapter_id);
CREATE INDEX idx_character_states_tenant ON character_states(tenant_id);
CREATE INDEX idx_character_states_timeline ON character_states(character_id, timeline_day DESC) WHERE timeline_day IS NOT NULL;
