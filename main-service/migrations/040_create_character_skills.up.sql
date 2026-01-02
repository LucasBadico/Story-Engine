CREATE TABLE IF NOT EXISTS character_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES rpg_skills(id) ON DELETE CASCADE,
    
    rank INT NOT NULL DEFAULT 1,
    xp_in_skill INT DEFAULT 0,              -- progressão dentro da skill
    is_active BOOLEAN DEFAULT TRUE,         -- skill equipada/ativa
    
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(character_id, skill_id)
);

CREATE INDEX idx_character_skills_character ON character_skills(character_id);
CREATE INDEX idx_character_skills_skill ON character_skills(skill_id);
CREATE INDEX idx_character_skills_active ON character_skills(character_id, is_active);


