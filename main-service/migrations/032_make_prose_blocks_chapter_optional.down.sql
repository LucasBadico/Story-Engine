-- Revert chapter_id to NOT NULL
ALTER TABLE prose_blocks ALTER COLUMN chapter_id SET NOT NULL;

-- Revert entity_type constraint
ALTER TABLE prose_block_references 
    DROP CONSTRAINT IF EXISTS prose_block_references_entity_type_check;

ALTER TABLE prose_block_references 
    ADD CONSTRAINT prose_block_references_entity_type_check 
    CHECK (entity_type IN ('scene', 'beat', 'character', 'location', 'trait'));

