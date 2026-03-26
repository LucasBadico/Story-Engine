ALTER TABLE embedding_chunks
  DROP COLUMN IF EXISTS timeline_confidence,
  DROP COLUMN IF EXISTS timeline_day,
  DROP COLUMN IF EXISTS chapter_number;

DROP INDEX IF EXISTS idx_embedding_documents_canon_retrieval;

ALTER TABLE embedding_documents
  DROP COLUMN IF EXISTS story_id,
  DROP COLUMN IF EXISTS chapter_number,
  DROP COLUMN IF EXISTS is_canonical,
  DROP COLUMN IF EXISTS valid_to_chapter,
  DROP COLUMN IF EXISTS valid_from_chapter;
