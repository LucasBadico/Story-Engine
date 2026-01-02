-- Add tenant_id to all tables that don't have it yet
-- This migration adds tenant_id column and populates it based on parent relationships

-- ============================================
-- World Building Tables (via world_id)
-- ============================================

-- Characters
ALTER TABLE characters ADD COLUMN tenant_id UUID;
UPDATE characters c SET tenant_id = w.tenant_id FROM worlds w WHERE c.world_id = w.id;
ALTER TABLE characters ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE characters ADD CONSTRAINT characters_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_characters_tenant_id ON characters(tenant_id);

-- Locations
ALTER TABLE locations ADD COLUMN tenant_id UUID;
UPDATE locations l SET tenant_id = w.tenant_id FROM worlds w WHERE l.world_id = w.id;
ALTER TABLE locations ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE locations ADD CONSTRAINT locations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_locations_tenant_id ON locations(tenant_id);

-- Artifacts
ALTER TABLE artifacts ADD COLUMN tenant_id UUID;
UPDATE artifacts a SET tenant_id = w.tenant_id FROM worlds w WHERE a.world_id = w.id;
ALTER TABLE artifacts ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE artifacts ADD CONSTRAINT artifacts_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_artifacts_tenant_id ON artifacts(tenant_id);

-- Events
ALTER TABLE events ADD COLUMN tenant_id UUID;
UPDATE events e SET tenant_id = w.tenant_id FROM worlds w WHERE e.world_id = w.id;
ALTER TABLE events ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE events ADD CONSTRAINT events_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_events_tenant_id ON events(tenant_id);

-- Artifact References
ALTER TABLE artifact_references ADD COLUMN tenant_id UUID;
UPDATE artifact_references ar SET tenant_id = a.tenant_id FROM artifacts a WHERE ar.artifact_id = a.id;
ALTER TABLE artifact_references ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE artifact_references ADD CONSTRAINT artifact_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_artifact_references_tenant_id ON artifact_references(tenant_id);

-- Event Characters
ALTER TABLE event_characters ADD COLUMN tenant_id UUID;
UPDATE event_characters ec SET tenant_id = e.tenant_id FROM events e WHERE ec.event_id = e.id;
ALTER TABLE event_characters ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE event_characters ADD CONSTRAINT event_characters_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_event_characters_tenant_id ON event_characters(tenant_id);

-- Event Locations
ALTER TABLE event_locations ADD COLUMN tenant_id UUID;
UPDATE event_locations el SET tenant_id = e.tenant_id FROM events e WHERE el.event_id = e.id;
ALTER TABLE event_locations ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE event_locations ADD CONSTRAINT event_locations_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_event_locations_tenant_id ON event_locations(tenant_id);

-- Event Artifacts
ALTER TABLE event_artifacts ADD COLUMN tenant_id UUID;
UPDATE event_artifacts ea SET tenant_id = e.tenant_id FROM events e WHERE ea.event_id = e.id;
ALTER TABLE event_artifacts ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE event_artifacts ADD CONSTRAINT event_artifacts_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_event_artifacts_tenant_id ON event_artifacts(tenant_id);

-- Character Traits
ALTER TABLE character_traits ADD COLUMN tenant_id UUID;
UPDATE character_traits ct SET tenant_id = c.tenant_id FROM characters c WHERE ct.character_id = c.id;
ALTER TABLE character_traits ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE character_traits ADD CONSTRAINT character_traits_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_character_traits_tenant_id ON character_traits(tenant_id);

-- ============================================
-- Story Tables (via story_id or chapter_id)
-- ============================================

-- Chapters
ALTER TABLE chapters ADD COLUMN tenant_id UUID;
UPDATE chapters ch SET tenant_id = s.tenant_id FROM stories s WHERE ch.story_id = s.id;
ALTER TABLE chapters ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE chapters ADD CONSTRAINT chapters_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_chapters_tenant_id ON chapters(tenant_id);

-- Scenes
ALTER TABLE scenes ADD COLUMN tenant_id UUID;
UPDATE scenes sc SET tenant_id = s.tenant_id FROM stories s WHERE sc.story_id = s.id;
ALTER TABLE scenes ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE scenes ADD CONSTRAINT scenes_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_scenes_tenant_id ON scenes(tenant_id);

-- Beats
ALTER TABLE beats ADD COLUMN tenant_id UUID;
UPDATE beats b SET tenant_id = s.tenant_id FROM scenes sc JOIN stories s ON sc.story_id = s.id WHERE b.scene_id = sc.id;
ALTER TABLE beats ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE beats ADD CONSTRAINT beats_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_beats_tenant_id ON beats(tenant_id);

-- Prose Blocks
ALTER TABLE prose_blocks ADD COLUMN tenant_id UUID;
UPDATE prose_blocks pb SET tenant_id = s.tenant_id FROM scenes sc JOIN stories s ON sc.story_id = s.id WHERE pb.scene_id = sc.id;
ALTER TABLE prose_blocks ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE prose_blocks ADD CONSTRAINT prose_blocks_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_prose_blocks_tenant_id ON prose_blocks(tenant_id);

-- Image Blocks
ALTER TABLE image_blocks ADD COLUMN tenant_id UUID;
UPDATE image_blocks ib SET tenant_id = s.tenant_id FROM chapters ch JOIN stories s ON ch.story_id = s.id WHERE ib.chapter_id = ch.id;
ALTER TABLE image_blocks ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE image_blocks ADD CONSTRAINT image_blocks_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_image_blocks_tenant_id ON image_blocks(tenant_id);

-- Scene References
ALTER TABLE scene_references ADD COLUMN tenant_id UUID;
UPDATE scene_references sr SET tenant_id = s.tenant_id FROM scenes sc JOIN stories s ON sc.story_id = s.id WHERE sr.scene_id = sc.id;
ALTER TABLE scene_references ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE scene_references ADD CONSTRAINT scene_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_scene_references_tenant_id ON scene_references(tenant_id);

-- Prose Block References
ALTER TABLE prose_block_references ADD COLUMN tenant_id UUID;
UPDATE prose_block_references pbr SET tenant_id = s.tenant_id FROM prose_blocks pb JOIN scenes sc ON pb.scene_id = sc.id JOIN stories s ON sc.story_id = s.id WHERE pbr.prose_block_id = pb.id;
ALTER TABLE prose_block_references ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE prose_block_references ADD CONSTRAINT prose_block_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_prose_block_references_tenant_id ON prose_block_references(tenant_id);

-- Image Block References
ALTER TABLE image_block_references ADD COLUMN tenant_id UUID;
UPDATE image_block_references ibr SET tenant_id = s.tenant_id FROM image_blocks ib JOIN chapters ch ON ib.chapter_id = ch.id JOIN stories s ON ch.story_id = s.id WHERE ibr.image_block_id = ib.id;
ALTER TABLE image_block_references ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE image_block_references ADD CONSTRAINT image_block_references_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_image_block_references_tenant_id ON image_block_references(tenant_id);

-- ============================================
-- RPG Tables (via rpg_system_id or character_id)
-- ============================================

-- RPG Skills
ALTER TABLE rpg_skills ADD COLUMN tenant_id UUID;
UPDATE rpg_skills rs SET tenant_id = rgs.tenant_id FROM rpg_systems rgs WHERE rs.rpg_system_id = rgs.id;
ALTER TABLE rpg_skills ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE rpg_skills ADD CONSTRAINT rpg_skills_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_rpg_skills_tenant_id ON rpg_skills(tenant_id);

-- RPG Classes
ALTER TABLE rpg_classes ADD COLUMN tenant_id UUID;
UPDATE rpg_classes rc SET tenant_id = rgs.tenant_id FROM rpg_systems rgs WHERE rc.rpg_system_id = rgs.id;
ALTER TABLE rpg_classes ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE rpg_classes ADD CONSTRAINT rpg_classes_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_rpg_classes_tenant_id ON rpg_classes(tenant_id);

-- RPG Class Skills
ALTER TABLE rpg_class_skills ADD COLUMN tenant_id UUID;
UPDATE rpg_class_skills rcs SET tenant_id = rc.tenant_id FROM rpg_classes rc WHERE rcs.class_id = rc.id;
ALTER TABLE rpg_class_skills ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE rpg_class_skills ADD CONSTRAINT rpg_class_skills_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_rpg_class_skills_tenant_id ON rpg_class_skills(tenant_id);

-- Inventory Slots
ALTER TABLE inventory_slots ADD COLUMN tenant_id UUID;
UPDATE inventory_slots ins SET tenant_id = rgs.tenant_id FROM rpg_systems rgs WHERE ins.rpg_system_id = rgs.id;
ALTER TABLE inventory_slots ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE inventory_slots ADD CONSTRAINT inventory_slots_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_inventory_slots_tenant_id ON inventory_slots(tenant_id);

-- Inventory Items
ALTER TABLE inventory_items ADD COLUMN tenant_id UUID;
UPDATE inventory_items ini SET tenant_id = rgs.tenant_id FROM rpg_systems rgs WHERE ini.rpg_system_id = rgs.id;
ALTER TABLE inventory_items ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE inventory_items ADD CONSTRAINT inventory_items_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_inventory_items_tenant_id ON inventory_items(tenant_id);

-- Character Skills
ALTER TABLE character_skills ADD COLUMN tenant_id UUID;
UPDATE character_skills cs SET tenant_id = c.tenant_id FROM characters c WHERE cs.character_id = c.id;
ALTER TABLE character_skills ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE character_skills ADD CONSTRAINT character_skills_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_character_skills_tenant_id ON character_skills(tenant_id);

-- Character RPG Stats
ALTER TABLE character_rpg_stats ADD COLUMN tenant_id UUID;
UPDATE character_rpg_stats crs SET tenant_id = c.tenant_id FROM characters c WHERE crs.character_id = c.id;
ALTER TABLE character_rpg_stats ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE character_rpg_stats ADD CONSTRAINT character_rpg_stats_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_character_rpg_stats_tenant_id ON character_rpg_stats(tenant_id);

-- Artifact RPG Stats
ALTER TABLE artifact_rpg_stats ADD COLUMN tenant_id UUID;
UPDATE artifact_rpg_stats ars SET tenant_id = a.tenant_id FROM artifacts a WHERE ars.artifact_id = a.id;
ALTER TABLE artifact_rpg_stats ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE artifact_rpg_stats ADD CONSTRAINT artifact_rpg_stats_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_artifact_rpg_stats_tenant_id ON artifact_rpg_stats(tenant_id);

-- Character Inventory
ALTER TABLE character_inventory ADD COLUMN tenant_id UUID;
UPDATE character_inventory ci SET tenant_id = c.tenant_id FROM characters c WHERE ci.character_id = c.id;
ALTER TABLE character_inventory ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE character_inventory ADD CONSTRAINT character_inventory_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_character_inventory_tenant_id ON character_inventory(tenant_id);

