-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- embedding_documents table
-- Note: tenant_id references tenants table from main-service
-- If tenants table doesn't exist yet, the foreign key will be added later
CREATE TABLE embedding_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    source_type VARCHAR(50) NOT NULL,  -- 'story', 'chapter', 'scene', 'beat', 'prose_block'
    source_id UUID NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Ensure one document per source
    UNIQUE(tenant_id, source_type, source_id)
);

-- Add foreign key constraint if tenants table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants') THEN
        ALTER TABLE embedding_documents 
        ADD CONSTRAINT embedding_documents_tenant_id_fk 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Indexes for common queries
CREATE INDEX idx_embedding_documents_tenant_id ON embedding_documents(tenant_id);
CREATE INDEX idx_embedding_documents_source ON embedding_documents(source_type, source_id);
CREATE INDEX idx_embedding_documents_created_at ON embedding_documents(created_at DESC);

