DROP INDEX IF EXISTS idx_events_timeline_position;
DROP INDEX IF EXISTS idx_events_hierarchy;
DROP INDEX IF EXISTS idx_events_world_parent;
DROP INDEX IF EXISTS idx_events_parent_id;

ALTER TABLE events DROP COLUMN is_epoch;
ALTER TABLE events DROP COLUMN timeline_position;
ALTER TABLE events DROP COLUMN hierarchy_level;
ALTER TABLE events DROP COLUMN parent_id;

