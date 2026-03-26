-- Replace plot_development_entities with timeline_event link + free-text notes field.
-- Entity context comes through the linked timeline_event.entities_involved.

-- Add new columns
ALTER TABLE plot_developments ADD COLUMN notes TEXT;
ALTER TABLE plot_developments ADD COLUMN timeline_event_id UUID REFERENCES timeline_events(id) ON DELETE SET NULL;

CREATE INDEX idx_plot_developments_event ON plot_developments(timeline_event_id) WHERE timeline_event_id IS NOT NULL;

-- Drop the entity join table
DROP TABLE IF EXISTS plot_development_entities;
