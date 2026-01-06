CREATE TABLE IF NOT EXISTS character_relationships (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    character1_id TEXT NOT NULL,
    character2_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    bidirectional INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT character_relationships_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT character_relationships_character1_fk FOREIGN KEY (character1_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT character_relationships_character2_fk FOREIGN KEY (character2_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT character_relationships_characters_different CHECK (character1_id != character2_id),
    CONSTRAINT character_relationships_unique UNIQUE (character1_id, character2_id)
);

CREATE INDEX IF NOT EXISTS idx_character_relationships_tenant_id ON character_relationships(tenant_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_character1_id ON character_relationships(character1_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_character2_id ON character_relationships(character2_id);

