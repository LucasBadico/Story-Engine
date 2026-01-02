CREATE TABLE IF NOT EXISTS archetype_traits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archetype_id UUID NOT NULL REFERENCES archetypes(id) ON DELETE CASCADE,
    trait_id UUID NOT NULL REFERENCES traits(id) ON DELETE CASCADE,
    default_value VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(archetype_id, trait_id)
);

CREATE INDEX idx_archetype_traits_archetype_id ON archetype_traits(archetype_id);
CREATE INDEX idx_archetype_traits_trait_id ON archetype_traits(trait_id);


