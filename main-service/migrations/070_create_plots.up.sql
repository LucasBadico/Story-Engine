CREATE TABLE IF NOT EXISTS plots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'secondary',
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plots_story ON plots(story_id);
CREATE INDEX idx_plots_tenant ON plots(tenant_id);
CREATE INDEX idx_plots_status ON plots(story_id, status);

CREATE TABLE IF NOT EXISTS plot_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plot_id UUID NOT NULL REFERENCES plots(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plot_entities_plot ON plot_entities(plot_id);
CREATE INDEX idx_plot_entities_entity ON plot_entities(entity_type, entity_id);

CREATE TABLE IF NOT EXISTS plot_developments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plot_id UUID NOT NULL REFERENCES plots(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    timeline_day INT,
    timeline_confidence VARCHAR(16),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plot_developments_plot ON plot_developments(plot_id, chapter_number);
CREATE INDEX idx_plot_developments_chapter ON plot_developments(chapter_number);
CREATE INDEX idx_plot_developments_tenant ON plot_developments(tenant_id);

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
