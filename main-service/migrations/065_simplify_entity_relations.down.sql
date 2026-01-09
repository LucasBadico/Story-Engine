-- Restore columns for LLM pipeline
ALTER TABLE entity_relations ADD COLUMN IF NOT EXISTS source_ref VARCHAR(100);
ALTER TABLE entity_relations ADD COLUMN IF NOT EXISTS target_ref VARCHAR(100);
ALTER TABLE entity_relations ADD COLUMN IF NOT EXISTS confidence REAL;
ALTER TABLE entity_relations ADD COLUMN IF NOT EXISTS evidence_spans JSONB;
ALTER TABLE entity_relations ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'confirmed';

-- Update source_ref and target_ref with existing IDs
UPDATE entity_relations SET source_ref = source_id::text WHERE source_ref IS NULL;
UPDATE entity_relations SET target_ref = target_id::text WHERE target_ref IS NULL;

-- Make source_ref NOT NULL after populating
ALTER TABLE entity_relations ALTER COLUMN source_ref SET NOT NULL;
ALTER TABLE entity_relations ALTER COLUMN target_ref SET NOT NULL;

-- Allow source_id and target_id to be NULL again
ALTER TABLE entity_relations ALTER COLUMN source_id DROP NOT NULL;
ALTER TABLE entity_relations ALTER COLUMN target_id DROP NOT NULL;

-- Restore status constraint
ALTER TABLE entity_relations ADD CONSTRAINT entity_relations_status_check CHECK (status IN ('draft', 'confirmed', 'orphan'));

-- Restore indexes
DROP INDEX IF EXISTS idx_entity_relations_source;
DROP INDEX IF EXISTS idx_entity_relations_target;
CREATE INDEX idx_entity_relations_source ON entity_relations(source_type, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX idx_entity_relations_target ON entity_relations(target_type, target_id) WHERE target_id IS NOT NULL;
CREATE INDEX idx_entity_relations_source_ref ON entity_relations(source_ref);
CREATE INDEX idx_entity_relations_target_ref ON entity_relations(target_ref);
CREATE INDEX idx_entity_relations_status ON entity_relations(status);
CREATE INDEX idx_entity_relations_temp_refs ON entity_relations(status) WHERE source_ref LIKE 'tmp:%' OR target_ref LIKE 'tmp:%';

