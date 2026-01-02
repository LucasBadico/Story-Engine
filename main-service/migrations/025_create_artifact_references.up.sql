CREATE TABLE IF NOT EXISTS artifact_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(artifact_id, entity_type, entity_id),
    CONSTRAINT artifact_references_entity_type_check CHECK (entity_type IN ('character', 'location'))
);

CREATE INDEX idx_artifact_references_artifact_id ON artifact_references(artifact_id);
CREATE INDEX idx_artifact_references_entity ON artifact_references(entity_type, entity_id);


