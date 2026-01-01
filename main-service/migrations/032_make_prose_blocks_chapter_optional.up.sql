-- Make chapter_id optional in prose_blocks (can be related to world entities via references)
ALTER TABLE prose_blocks ALTER COLUMN chapter_id DROP NOT NULL;

-- Update prose_block_references to include new world entities
ALTER TABLE prose_block_references 
    DROP CONSTRAINT IF EXISTS prose_block_references_entity_type_check;

ALTER TABLE prose_block_references 
    ADD CONSTRAINT prose_block_references_entity_type_check 
    CHECK (entity_type IN ('scene', 'beat', 'chapter', 'character', 'location', 'artifact', 'event', 'world'));

