ALTER TABLE events ADD COLUMN parent_id TEXT REFERENCES events(id) ON DELETE SET NULL;
ALTER TABLE events ADD COLUMN hierarchy_level INTEGER NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN timeline_position REAL NOT NULL DEFAULT 0.0;
ALTER TABLE events ADD COLUMN is_epoch INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_events_parent_id ON events(parent_id);
CREATE INDEX idx_events_world_parent ON events(world_id, parent_id);
CREATE INDEX idx_events_hierarchy ON events(world_id, hierarchy_level);
CREATE INDEX idx_events_timeline_position ON events(world_id, timeline_position);

-- SQLite não suporta partial unique index, validar na aplicação

