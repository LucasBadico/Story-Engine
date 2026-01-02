-- Remove tenant_id columns and indexes from all tables

-- World Building Tables
DROP INDEX IF EXISTS idx_characters_tenant_id;
ALTER TABLE characters DROP CONSTRAINT IF EXISTS characters_tenant_fk;
ALTER TABLE characters DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_locations_tenant_id;
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_tenant_fk;
ALTER TABLE locations DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_artifacts_tenant_id;
ALTER TABLE artifacts DROP CONSTRAINT IF EXISTS artifacts_tenant_fk;
ALTER TABLE artifacts DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_events_tenant_id;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_tenant_fk;
ALTER TABLE events DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_artifact_references_tenant_id;
ALTER TABLE artifact_references DROP CONSTRAINT IF EXISTS artifact_references_tenant_fk;
ALTER TABLE artifact_references DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_event_characters_tenant_id;
ALTER TABLE event_characters DROP CONSTRAINT IF EXISTS event_characters_tenant_fk;
ALTER TABLE event_characters DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_event_locations_tenant_id;
ALTER TABLE event_locations DROP CONSTRAINT IF EXISTS event_locations_tenant_fk;
ALTER TABLE event_locations DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_event_artifacts_tenant_id;
ALTER TABLE event_artifacts DROP CONSTRAINT IF EXISTS event_artifacts_tenant_fk;
ALTER TABLE event_artifacts DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_character_traits_tenant_id;
ALTER TABLE character_traits DROP CONSTRAINT IF EXISTS character_traits_tenant_fk;
ALTER TABLE character_traits DROP COLUMN IF EXISTS tenant_id;

-- Story Tables
DROP INDEX IF EXISTS idx_chapters_tenant_id;
ALTER TABLE chapters DROP CONSTRAINT IF EXISTS chapters_tenant_fk;
ALTER TABLE chapters DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_scenes_tenant_id;
ALTER TABLE scenes DROP CONSTRAINT IF EXISTS scenes_tenant_fk;
ALTER TABLE scenes DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_beats_tenant_id;
ALTER TABLE beats DROP CONSTRAINT IF EXISTS beats_tenant_fk;
ALTER TABLE beats DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_prose_blocks_tenant_id;
ALTER TABLE prose_blocks DROP CONSTRAINT IF EXISTS prose_blocks_tenant_fk;
ALTER TABLE prose_blocks DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_image_blocks_tenant_id;
ALTER TABLE image_blocks DROP CONSTRAINT IF EXISTS image_blocks_tenant_fk;
ALTER TABLE image_blocks DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_scene_references_tenant_id;
ALTER TABLE scene_references DROP CONSTRAINT IF EXISTS scene_references_tenant_fk;
ALTER TABLE scene_references DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_prose_block_references_tenant_id;
ALTER TABLE prose_block_references DROP CONSTRAINT IF EXISTS prose_block_references_tenant_fk;
ALTER TABLE prose_block_references DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_image_block_references_tenant_id;
ALTER TABLE image_block_references DROP CONSTRAINT IF EXISTS image_block_references_tenant_fk;
ALTER TABLE image_block_references DROP COLUMN IF EXISTS tenant_id;

-- RPG Tables
DROP INDEX IF EXISTS idx_rpg_skills_tenant_id;
ALTER TABLE rpg_skills DROP CONSTRAINT IF EXISTS rpg_skills_tenant_fk;
ALTER TABLE rpg_skills DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_rpg_classes_tenant_id;
ALTER TABLE rpg_classes DROP CONSTRAINT IF EXISTS rpg_classes_tenant_fk;
ALTER TABLE rpg_classes DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_rpg_class_skills_tenant_id;
ALTER TABLE rpg_class_skills DROP CONSTRAINT IF EXISTS rpg_class_skills_tenant_fk;
ALTER TABLE rpg_class_skills DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_inventory_slots_tenant_id;
ALTER TABLE inventory_slots DROP CONSTRAINT IF EXISTS inventory_slots_tenant_fk;
ALTER TABLE inventory_slots DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_inventory_items_tenant_id;
ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_tenant_fk;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_character_skills_tenant_id;
ALTER TABLE character_skills DROP CONSTRAINT IF EXISTS character_skills_tenant_fk;
ALTER TABLE character_skills DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_character_rpg_stats_tenant_id;
ALTER TABLE character_rpg_stats DROP CONSTRAINT IF EXISTS character_rpg_stats_tenant_fk;
ALTER TABLE character_rpg_stats DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_artifact_rpg_stats_tenant_id;
ALTER TABLE artifact_rpg_stats DROP CONSTRAINT IF EXISTS artifact_rpg_stats_tenant_fk;
ALTER TABLE artifact_rpg_stats DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_character_inventory_tenant_id;
ALTER TABLE character_inventory DROP CONSTRAINT IF EXISTS character_inventory_tenant_fk;
ALTER TABLE character_inventory DROP COLUMN IF EXISTS tenant_id;

