-- Add tenant_id to archetype_traits table
-- This migration adds tenant_id column and populates it based on archetype relationship

ALTER TABLE archetype_traits ADD COLUMN tenant_id UUID;
UPDATE archetype_traits at SET tenant_id = a.tenant_id FROM archetypes a WHERE at.archetype_id = a.id;
ALTER TABLE archetype_traits ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE archetype_traits ADD CONSTRAINT archetype_traits_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_archetype_traits_tenant_id ON archetype_traits(tenant_id);

