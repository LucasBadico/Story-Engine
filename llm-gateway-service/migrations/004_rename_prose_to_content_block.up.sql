-- Rename prose_kind to content_kind
ALTER TABLE embedding_chunks RENAME COLUMN prose_kind TO content_kind;

-- Add content_type column
ALTER TABLE embedding_chunks ADD COLUMN content_type VARCHAR(50);

-- Update source_type in embedding_documents from prose_block to content_block
UPDATE embedding_documents SET source_type = 'content_block' WHERE source_type = 'prose_block';

-- Create index for content_type
CREATE INDEX idx_embedding_chunks_content_type ON embedding_chunks(content_type);

