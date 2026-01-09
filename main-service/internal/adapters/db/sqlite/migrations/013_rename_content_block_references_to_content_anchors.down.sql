-- Recreate content_block_references from content_anchors
ALTER TABLE content_anchors RENAME TO tmp_content_anchors;

CREATE TABLE IF NOT EXISTS content_block_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    content_block_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CONSTRAINT content_block_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT content_block_references_content_block_fk FOREIGN KEY (content_block_id) REFERENCES content_blocks(id) ON DELETE CASCADE,
    CONSTRAINT content_block_references_entity_type_check CHECK (entity_type IN (
        'scene', 'beat', 'chapter', 'character', 'location', 'artifact',
        'event', 'world', 'rpg_system', 'rpg_skill', 'rpg_class',
        'inventory_item', 'faction', 'lore', 'faction_reference', 'lore_reference'
    )),
    CONSTRAINT content_block_references_unique UNIQUE (content_block_id, entity_type, entity_id)
);

INSERT INTO content_block_references (id, tenant_id, content_block_id, entity_type, entity_id, created_at)
SELECT id, tenant_id, content_block_id, entity_type, entity_id, created_at
FROM tmp_content_anchors;

DROP TABLE tmp_content_anchors;

DROP INDEX IF EXISTS idx_content_anchors_content_block_id;
DROP INDEX IF EXISTS idx_content_anchors_tenant_id;
DROP INDEX IF EXISTS idx_content_anchors_entity;
DROP INDEX IF EXISTS idx_content_anchors_entity_id;

CREATE INDEX IF NOT EXISTS idx_content_block_references_content_block_id ON content_block_references(content_block_id);
CREATE INDEX IF NOT EXISTS idx_content_block_references_tenant_id ON content_block_references(tenant_id);
CREATE INDEX IF NOT EXISTS idx_content_block_references_entity ON content_block_references(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_content_block_references_entity_id ON content_block_references(entity_id);


