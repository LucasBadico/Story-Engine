-- Drop content_type index
DROP INDEX IF EXISTS idx_embedding_chunks_content_type;

-- Revert source_type in embedding_documents from content_block to prose_block
UPDATE embedding_documents SET source_type = 'prose_block' WHERE source_type = 'content_block';

-- Remove content_type column
ALTER TABLE embedding_chunks DROP COLUMN IF EXISTS content_type;

-- Rename content_kind back to prose_kind
ALTER TABLE embedding_chunks RENAME COLUMN content_kind TO prose_kind;

