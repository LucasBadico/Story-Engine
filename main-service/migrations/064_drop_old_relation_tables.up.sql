-- Drop old relation tables - they are now replaced by entity_relations
-- This is a greenfield migration (no backfill needed)

-- Drop indexes first
DROP INDEX IF EXISTS idx_character_relationships_character2_id;
DROP INDEX IF EXISTS idx_character_relationships_character1_id;
DROP INDEX IF EXISTS idx_character_relationships_tenant_id;
DROP TABLE IF EXISTS character_relationships;

DROP INDEX IF EXISTS idx_event_references_entity;
DROP INDEX IF EXISTS idx_event_references_event_id;
DROP INDEX IF EXISTS idx_event_references_tenant_id;
DROP TABLE IF EXISTS event_references;

DROP INDEX IF EXISTS idx_lore_references_entity;
DROP INDEX IF EXISTS idx_lore_references_lore_id;
DROP INDEX IF EXISTS idx_lore_references_tenant_id;
DROP TABLE IF EXISTS lore_references;

DROP INDEX IF EXISTS idx_faction_references_entity;
DROP INDEX IF EXISTS idx_faction_references_faction_id;
DROP INDEX IF EXISTS idx_faction_references_tenant_id;
DROP TABLE IF EXISTS faction_references;

DROP INDEX IF EXISTS idx_artifact_references_entity;
DROP INDEX IF EXISTS idx_artifact_references_artifact_id;
DROP INDEX IF EXISTS idx_artifact_references_tenant_id;
DROP TABLE IF EXISTS artifact_references;

DROP INDEX IF EXISTS idx_scene_references_entity;
DROP INDEX IF EXISTS idx_scene_references_scene_id;
DROP INDEX IF EXISTS idx_scene_references_tenant_id;
DROP TABLE IF EXISTS scene_references;

-- Note: content_block_references will be migrated in Phase 6 when updating endpoints

