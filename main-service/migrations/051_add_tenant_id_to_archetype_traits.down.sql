-- Remove tenant_id column and index from archetype_traits table

DROP INDEX IF EXISTS idx_archetype_traits_tenant_id;
ALTER TABLE archetype_traits DROP CONSTRAINT IF EXISTS archetype_traits_tenant_fk;
ALTER TABLE archetype_traits DROP COLUMN IF EXISTS tenant_id;

