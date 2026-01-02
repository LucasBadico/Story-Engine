CREATE TABLE IF NOT EXISTS event_characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    role VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, character_id)
);

CREATE INDEX idx_event_characters_event_id ON event_characters(event_id);
CREATE INDEX idx_event_characters_character_id ON event_characters(character_id);


