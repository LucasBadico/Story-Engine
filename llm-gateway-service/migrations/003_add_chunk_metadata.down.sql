-- Drop indexes
DROP INDEX IF EXISTS idx_embedding_chunks_characters_gin;
DROP INDEX IF EXISTS idx_embedding_chunks_prose_kind;
DROP INDEX IF EXISTS idx_embedding_chunks_location_id;
DROP INDEX IF EXISTS idx_embedding_chunks_beat_type;
DROP INDEX IF EXISTS idx_embedding_chunks_beat_id;
DROP INDEX IF EXISTS idx_embedding_chunks_scene_id;

-- Drop metadata columns
ALTER TABLE embedding_chunks
    DROP COLUMN IF EXISTS prose_kind,
    DROP COLUMN IF EXISTS pov_character,
    DROP COLUMN IF EXISTS timeline,
    DROP COLUMN IF EXISTS location_name,
    DROP COLUMN IF EXISTS location_id,
    DROP COLUMN IF EXISTS characters,
    DROP COLUMN IF EXISTS beat_intent,
    DROP COLUMN IF EXISTS beat_type,
    DROP COLUMN IF EXISTS beat_id,
    DROP COLUMN IF EXISTS scene_id;


