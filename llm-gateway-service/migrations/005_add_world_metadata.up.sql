-- Add World entity metadata columns to embedding_chunks table
ALTER TABLE embedding_chunks
    ADD COLUMN world_id UUID,
    ADD COLUMN world_name VARCHAR(255),
    ADD COLUMN world_genre VARCHAR(100),
    ADD COLUMN entity_name VARCHAR(255),
    ADD COLUMN event_timeline VARCHAR(255),
    ADD COLUMN importance INTEGER,
    ADD COLUMN related_characters JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN related_locations JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN related_artifacts JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN related_events JSONB DEFAULT '[]'::jsonb;

-- Create indexes for World metadata
CREATE INDEX idx_embedding_chunks_world_id ON embedding_chunks(world_id);
CREATE INDEX idx_embedding_chunks_entity_name ON embedding_chunks(entity_name);
CREATE INDEX idx_embedding_chunks_importance ON embedding_chunks(importance);

-- GIN indexes for JSONB arrays for efficient searching
CREATE INDEX idx_embedding_chunks_related_characters_gin ON embedding_chunks USING GIN (related_characters);
CREATE INDEX idx_embedding_chunks_related_locations_gin ON embedding_chunks USING GIN (related_locations);
CREATE INDEX idx_embedding_chunks_related_artifacts_gin ON embedding_chunks USING GIN (related_artifacts);
CREATE INDEX idx_embedding_chunks_related_events_gin ON embedding_chunks USING GIN (related_events);

