CREATE TABLE IF NOT EXISTS rpg_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rpg_system_id UUID NOT NULL REFERENCES rpg_systems(id) ON DELETE CASCADE,
    parent_class_id UUID REFERENCES rpg_classes(id) ON DELETE SET NULL,  -- evolução
    
    name VARCHAR(100) NOT NULL,
    tier INT DEFAULT 1,                     -- 1=base, 2=advanced, 3=master
    description TEXT,
    
    -- Requisitos para desbloquear
    requirements JSONB,                     -- {level: 30, stats: {strength: 20}, class: "warrior"}
    
    -- Bônus da classe
    stat_bonuses JSONB,                     -- {strength: +2, hp_per_level: +10}
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rpg_classes_rpg_system ON rpg_classes(rpg_system_id);
CREATE INDEX idx_rpg_classes_parent ON rpg_classes(parent_class_id);
CREATE INDEX idx_rpg_classes_tier ON rpg_classes(tier);

