-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- embedding_documents table
CREATE TABLE embedding_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL,  -- 'story', 'chapter', 'scene', 'beat', 'prose_block'
    source_id UUID NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Ensure one document per source
    UNIQUE(tenant_id, source_type, source_id)
);

-- Indexes for common queries
CREATE INDEX idx_embedding_documents_tenant_id ON embedding_documents(tenant_id);
CREATE INDEX idx_embedding_documents_source ON embedding_documents(source_type, source_id);
CREATE INDEX idx_embedding_documents_created_at ON embedding_documents(created_at DESC);

