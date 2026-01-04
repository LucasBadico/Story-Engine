CREATE TABLE IF NOT EXISTS lores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES lores(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    description TEXT,
    rules TEXT,
    limitations TEXT,
    requirements TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lores_tenant_id ON lores(tenant_id);
CREATE INDEX idx_lores_world_id ON lores(world_id);
CREATE INDEX idx_lores_parent_id ON lores(parent_id);
CREATE INDEX idx_lores_world_parent ON lores(world_id, parent_id);

