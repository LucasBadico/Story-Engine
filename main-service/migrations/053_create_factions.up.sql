CREATE TABLE IF NOT EXISTS factions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES factions(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50),
    description TEXT,
    beliefs TEXT,
    structure TEXT,
    symbols TEXT,
    hierarchy_level INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_factions_tenant_id ON factions(tenant_id);
CREATE INDEX idx_factions_world_id ON factions(world_id);
CREATE INDEX idx_factions_parent_id ON factions(parent_id);
CREATE INDEX idx_factions_world_parent ON factions(world_id, parent_id);

