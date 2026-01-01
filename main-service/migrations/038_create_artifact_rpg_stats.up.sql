CREATE TABLE IF NOT EXISTS artifact_rpg_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    
    stats JSONB NOT NULL,
    
    is_active BOOLEAN DEFAULT TRUE,
    version INT NOT NULL DEFAULT 1,
    reason TEXT,
    timeline VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_artifact_rpg_stats_artifact ON artifact_rpg_stats(artifact_id);
CREATE UNIQUE INDEX idx_artifact_rpg_stats_active ON artifact_rpg_stats(artifact_id) WHERE is_active = TRUE;
CREATE INDEX idx_artifact_rpg_stats_event ON artifact_rpg_stats(event_id);

