CREATE TABLE IF NOT EXISTS rpg_class_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id UUID NOT NULL REFERENCES rpg_classes(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES rpg_skills(id) ON DELETE CASCADE,
    unlock_level INT DEFAULT 1,             -- nível da classe para desbloquear
    
    UNIQUE(class_id, skill_id)
);

CREATE INDEX idx_rpg_class_skills_class ON rpg_class_skills(class_id);
CREATE INDEX idx_rpg_class_skills_skill ON rpg_class_skills(skill_id);


