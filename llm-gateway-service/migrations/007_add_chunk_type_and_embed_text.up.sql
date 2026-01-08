-- Add chunk type metadata and embed text to embedding_chunks table
ALTER TABLE embedding_chunks
    ADD COLUMN type VARCHAR(32) NOT NULL DEFAULT 'raw',
    ADD COLUMN embed_text TEXT;

-- Index for chunk type
CREATE INDEX idx_embedding_chunks_type ON embedding_chunks(type);
