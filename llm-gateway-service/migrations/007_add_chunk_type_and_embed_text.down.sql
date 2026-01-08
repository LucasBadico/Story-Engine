DROP INDEX IF EXISTS idx_embedding_chunks_type;

ALTER TABLE embedding_chunks
    DROP COLUMN IF EXISTS embed_text,
    DROP COLUMN IF EXISTS type;
