CREATE TABLE IF NOT EXISTS event_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    significance TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, location_id)
);

CREATE INDEX idx_event_locations_event_id ON event_locations(event_id);
CREATE INDEX idx_event_locations_location_id ON event_locations(location_id);

