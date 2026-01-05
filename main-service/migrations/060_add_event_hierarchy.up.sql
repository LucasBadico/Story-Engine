ALTER TABLE events
ADD COLUMN parent_id UUID REFERENCES events(id) ON DELETE SET NULL,
ADD COLUMN hierarchy_level INT NOT NULL DEFAULT 0,
ADD COLUMN timeline_position DOUBLE PRECISION NOT NULL DEFAULT 0.0,
ADD COLUMN is_epoch BOOLEAN NOT NULL DEFAULT FALSE;

-- Garantir apenas um epoch por mundo
CREATE UNIQUE INDEX idx_events_epoch_unique ON events(world_id) WHERE is_epoch = TRUE;

CREATE INDEX idx_events_parent_id ON events(parent_id);
CREATE INDEX idx_events_world_parent ON events(world_id, parent_id);
CREATE INDEX idx_events_hierarchy ON events(world_id, hierarchy_level);
CREATE INDEX idx_events_timeline_position ON events(world_id, timeline_position);

