-- Remove columns that were for LLM pipeline only
-- source_id and target_id are now required (entities must exist before creating relations)

-- Drop indexes that depend on columns we're removing
DROP INDEX IF EXISTS idx_entity_relations_source_ref;
DROP INDEX IF EXISTS idx_entity_relations_target_ref;
DROP INDEX IF EXISTS idx_entity_relations_temp_refs;
DROP INDEX IF EXISTS idx_entity_relations_status;

-- Remove the status constraint before dropping the column
ALTER TABLE entity_relations DROP CONSTRAINT IF EXISTS entity_relations_status_check;

-- Drop columns
ALTER TABLE entity_relations DROP COLUMN IF EXISTS source_ref;
ALTER TABLE entity_relations DROP COLUMN IF EXISTS target_ref;
ALTER TABLE entity_relations DROP COLUMN IF EXISTS confidence;
ALTER TABLE entity_relations DROP COLUMN IF EXISTS evidence_spans;
ALTER TABLE entity_relations DROP COLUMN IF EXISTS status;

-- Make source_id and target_id NOT NULL (must have valid entities)
-- First update any NULLs to prevent migration failure (shouldn't happen in practice)
DELETE FROM entity_relations WHERE source_id IS NULL OR target_id IS NULL;

-- Now add NOT NULL constraint
ALTER TABLE entity_relations ALTER COLUMN source_id SET NOT NULL;
ALTER TABLE entity_relations ALTER COLUMN target_id SET NOT NULL;

-- Update indexes for the new non-nullable columns
DROP INDEX IF EXISTS idx_entity_relations_source;
DROP INDEX IF EXISTS idx_entity_relations_target;
CREATE INDEX idx_entity_relations_source ON entity_relations(source_type, source_id);
CREATE INDEX idx_entity_relations_target ON entity_relations(target_type, target_id);

