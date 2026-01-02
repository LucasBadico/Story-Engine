CREATE TABLE IF NOT EXISTS rpg_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rpg_system_id UUID NOT NULL REFERENCES rpg_systems(id) ON DELETE CASCADE,
    
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),                   -- combat, magic, passive, utility
    type VARCHAR(50),                       -- active, passive, spell, ability
    description TEXT,
    
    -- Requisitos
    prerequisites JSONB,                    -- {skills: ["fireball"], stats: {intelligence: 15}}
    max_rank INT DEFAULT 10,
    
    -- Efeitos (para game engine)
    effects_schema JSONB,                   -- {damage: "2d6+INT", cooldown: 3}
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rpg_skills_rpg_system ON rpg_skills(rpg_system_id);
CREATE INDEX idx_rpg_skills_category ON rpg_skills(category);
CREATE INDEX idx_rpg_skills_type ON rpg_skills(type);


