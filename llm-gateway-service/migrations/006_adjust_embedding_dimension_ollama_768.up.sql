-- Adjust embedding vector dimension for Ollama (768)
DROP INDEX IF EXISTS idx_embedding_chunks_embedding;

ALTER TABLE embedding_chunks
    ALTER COLUMN embedding TYPE vector(768)
    USING embedding::vector(768);

CREATE INDEX idx_embedding_chunks_embedding
    ON embedding_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
