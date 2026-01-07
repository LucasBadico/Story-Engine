-- Revert embedding vector dimension to OpenAI default (1536)
DROP INDEX IF EXISTS idx_embedding_chunks_embedding;

ALTER TABLE embedding_chunks
    ALTER COLUMN embedding TYPE vector(1536)
    USING embedding::vector(1536);

CREATE INDEX idx_embedding_chunks_embedding
    ON embedding_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
