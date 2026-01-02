CREATE TABLE IF NOT EXISTS event_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    role VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, artifact_id)
);

CREATE INDEX idx_event_artifacts_event_id ON event_artifacts(event_id);
CREATE INDEX idx_event_artifacts_artifact_id ON event_artifacts(artifact_id);


