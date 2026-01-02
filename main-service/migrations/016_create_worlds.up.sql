CREATE TABLE IF NOT EXISTS worlds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    genre VARCHAR(50),
    is_implicit BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_worlds_tenant_id ON worlds(tenant_id);
CREATE INDEX idx_worlds_implicit ON worlds(tenant_id, is_implicit);


