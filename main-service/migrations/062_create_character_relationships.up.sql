CREATE TABLE IF NOT EXISTS character_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    character1_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    character2_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    relationship_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    bidirectional BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT character_relationships_characters_different CHECK (character1_id != character2_id),
    CONSTRAINT character_relationships_unique UNIQUE (character1_id, character2_id)
);

CREATE INDEX idx_character_relationships_tenant_id ON character_relationships(tenant_id);
CREATE INDEX idx_character_relationships_character1_id ON character_relationships(character1_id);
CREATE INDEX idx_character_relationships_character2_id ON character_relationships(character2_id);

