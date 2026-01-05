DROP INDEX IF EXISTS idx_events_timeline_position;
DROP INDEX IF EXISTS idx_events_hierarchy;
DROP INDEX IF EXISTS idx_events_world_parent;
DROP INDEX IF EXISTS idx_events_parent_id;
DROP INDEX IF EXISTS idx_events_epoch_unique;

ALTER TABLE events
DROP COLUMN IF EXISTS is_epoch,
DROP COLUMN IF EXISTS timeline_position,
DROP COLUMN IF EXISTS hierarchy_level,
DROP COLUMN IF EXISTS parent_id;

