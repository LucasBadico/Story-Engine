CREATE TABLE IF NOT EXISTS character_traits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    trait_id UUID NOT NULL REFERENCES traits(id) ON DELETE CASCADE,
    -- Copied trait data (snapshot)
    trait_name VARCHAR(100) NOT NULL,
    trait_category VARCHAR(50),
    trait_description TEXT,
    -- Character-specific customization
    value VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(character_id, trait_id)
);

CREATE INDEX idx_character_traits_character_id ON character_traits(character_id);
CREATE INDEX idx_character_traits_trait_id ON character_traits(trait_id);


