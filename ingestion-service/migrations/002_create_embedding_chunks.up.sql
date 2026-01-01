-- embedding_chunks table
CREATE TABLE embedding_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES embedding_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),  -- OpenAI ada-002 dimension (can be adjusted for other models)
    token_count INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Ensure unique chunk index per document
    UNIQUE(document_id, chunk_index)
);

-- Index for vector similarity search (IVFFlat index for pgvector)
-- Note: IVFFlat requires data to exist, so we'll create it after initial data is loaded
-- For now, create a basic index
CREATE INDEX idx_embedding_chunks_document_id ON embedding_chunks(document_id);
CREATE INDEX idx_embedding_chunks_embedding ON embedding_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Index for filtering by document
CREATE INDEX idx_embedding_chunks_document_chunk ON embedding_chunks(document_id, chunk_index);

