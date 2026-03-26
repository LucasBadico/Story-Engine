-- Restore plot_development_entities table
CREATE TABLE IF NOT EXISTS plot_development_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plot_development_id UUID NOT NULL REFERENCES plot_developments(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pde_development ON plot_development_entities(plot_development_id);
CREATE INDEX idx_pde_entity ON plot_development_entities(entity_type, entity_id);

-- Remove new columns
DROP INDEX IF EXISTS idx_plot_developments_event;
ALTER TABLE plot_developments DROP COLUMN IF EXISTS timeline_event_id;
ALTER TABLE plot_developments DROP COLUMN IF EXISTS notes;
