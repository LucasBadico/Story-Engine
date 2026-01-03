-- Drop indexes
DROP INDEX IF EXISTS idx_embedding_chunks_related_events_gin;
DROP INDEX IF EXISTS idx_embedding_chunks_related_artifacts_gin;
DROP INDEX IF EXISTS idx_embedding_chunks_related_locations_gin;
DROP INDEX IF EXISTS idx_embedding_chunks_related_characters_gin;
DROP INDEX IF EXISTS idx_embedding_chunks_importance;
DROP INDEX IF EXISTS idx_embedding_chunks_entity_name;
DROP INDEX IF EXISTS idx_embedding_chunks_world_id;

-- Remove World entity metadata columns
ALTER TABLE embedding_chunks
    DROP COLUMN IF EXISTS related_events,
    DROP COLUMN IF EXISTS related_artifacts,
    DROP COLUMN IF EXISTS related_locations,
    DROP COLUMN IF EXISTS related_characters,
    DROP COLUMN IF EXISTS importance,
    DROP COLUMN IF EXISTS event_timeline,
    DROP COLUMN IF EXISTS entity_name,
    DROP COLUMN IF EXISTS world_genre,
    DROP COLUMN IF EXISTS world_name,
    DROP COLUMN IF EXISTS world_id;

