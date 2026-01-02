CREATE TABLE IF NOT EXISTS rpg_systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Schemas dinâmicos
    base_stats_schema JSONB NOT NULL,      -- força, destreza, etc
    derived_stats_schema JSONB,             -- HP, mana (calculados)
    progression_schema JSONB,               -- XP, níveis, ranks
    
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_rpg_systems_tenant_name ON rpg_systems(tenant_id, name) WHERE tenant_id IS NOT NULL;
CREATE UNIQUE INDEX idx_rpg_systems_builtin_name ON rpg_systems(name) WHERE is_builtin = TRUE;
CREATE INDEX idx_rpg_systems_tenant_id ON rpg_systems(tenant_id);


