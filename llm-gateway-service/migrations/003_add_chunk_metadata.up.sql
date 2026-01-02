-- Add metadata columns to embedding_chunks table
ALTER TABLE embedding_chunks
    ADD COLUMN scene_id UUID,
    ADD COLUMN beat_id UUID,
    ADD COLUMN beat_type VARCHAR(50),
    ADD COLUMN beat_intent TEXT,
    ADD COLUMN characters JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN location_id UUID,
    ADD COLUMN location_name VARCHAR(255),
    ADD COLUMN timeline VARCHAR(255),
    ADD COLUMN pov_character VARCHAR(255),
    ADD COLUMN prose_kind VARCHAR(50);

-- Create indexes for common filters
CREATE INDEX idx_embedding_chunks_scene_id ON embedding_chunks(scene_id);
CREATE INDEX idx_embedding_chunks_beat_id ON embedding_chunks(beat_id);
CREATE INDEX idx_embedding_chunks_beat_type ON embedding_chunks(beat_type);
CREATE INDEX idx_embedding_chunks_location_id ON embedding_chunks(location_id);
CREATE INDEX idx_embedding_chunks_prose_kind ON embedding_chunks(prose_kind);

-- GIN index for JSONB characters array for efficient searching
CREATE INDEX idx_embedding_chunks_characters_gin ON embedding_chunks USING GIN (characters);


