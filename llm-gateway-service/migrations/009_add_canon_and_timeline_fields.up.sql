-- Extend embedding_documents for canon versioning
ALTER TABLE embedding_documents
  ADD COLUMN IF NOT EXISTS valid_from_chapter INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS valid_to_chapter INT,
  ADD COLUMN IF NOT EXISTS is_canonical BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS chapter_number INT,
  ADD COLUMN IF NOT EXISTS story_id UUID;

-- Index for canon retrieval queries
CREATE INDEX IF NOT EXISTS idx_embedding_documents_canon_retrieval
  ON embedding_documents (story_id, is_canonical, valid_from_chapter)
  WHERE is_canonical = true;

-- Extend embedding_chunks for dual-axis timeline
ALTER TABLE embedding_chunks
  ADD COLUMN IF NOT EXISTS chapter_number INT,
  ADD COLUMN IF NOT EXISTS timeline_day INT,
  ADD COLUMN IF NOT EXISTS timeline_confidence VARCHAR(16);
