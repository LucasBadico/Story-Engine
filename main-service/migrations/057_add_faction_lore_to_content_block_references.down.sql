-- Revert faction and lore entities from content_block_references
ALTER TABLE content_block_references 
DROP CONSTRAINT IF EXISTS content_block_references_entity_type_check;

ALTER TABLE content_block_references 
ADD CONSTRAINT content_block_references_entity_type_check 
CHECK (entity_type IN (
    'scene', 'beat', 'chapter', 'character', 'location', 
    'artifact', 'event', 'world',
    'rpg_system', 'rpg_skill', 'rpg_class', 'inventory_item'
));

