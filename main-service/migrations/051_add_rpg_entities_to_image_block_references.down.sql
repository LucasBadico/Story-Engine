-- Revert RPG entities from image_block_references
ALTER TABLE image_block_references 
DROP CONSTRAINT IF EXISTS image_block_references_entity_type_check;

ALTER TABLE image_block_references 
ADD CONSTRAINT image_block_references_entity_type_check 
CHECK (entity_type IN (
    'scene', 'beat', 'chapter', 'character', 'location', 
    'artifact', 'event', 'world'
));

