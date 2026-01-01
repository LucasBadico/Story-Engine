CREATE TABLE IF NOT EXISTS character_rpg_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    
    base_stats JSONB NOT NULL,
    derived_stats JSONB,
    progression JSONB,                      -- {level: 5, xp: 1200, rank: "foundation"}
    
    is_active BOOLEAN DEFAULT TRUE,
    version INT NOT NULL DEFAULT 1,
    reason TEXT,
    timeline VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_character_rpg_stats_character ON character_rpg_stats(character_id);
CREATE UNIQUE INDEX idx_character_rpg_stats_active ON character_rpg_stats(character_id) WHERE is_active = TRUE;
CREATE INDEX idx_character_rpg_stats_event ON character_rpg_stats(event_id);

