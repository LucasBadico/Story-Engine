CREATE TABLE IF NOT EXISTS timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,
    timeline_day INT,
    timeline_confidence VARCHAR(16),
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_timeline_events_story ON timeline_events(story_id, chapter_number);
CREATE INDEX idx_timeline_events_chapter ON timeline_events(chapter_id);
CREATE INDEX idx_timeline_events_tenant ON timeline_events(tenant_id);
CREATE INDEX idx_timeline_events_day ON timeline_events(story_id, timeline_day) WHERE timeline_day IS NOT NULL;

CREATE TABLE IF NOT EXISTS timeline_event_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timeline_event_id UUID NOT NULL REFERENCES timeline_events(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_te_entities_event ON timeline_event_entities(timeline_event_id);
CREATE INDEX idx_te_entities_entity ON timeline_event_entities(entity_type, entity_id);
