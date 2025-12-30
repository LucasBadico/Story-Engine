CREATE TABLE IF NOT EXISTS beats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id UUID NOT NULL,
    order_num INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    intent TEXT,
    outcome TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT beats_scene_fk FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE,
    CONSTRAINT beats_scene_order_unique UNIQUE (scene_id, order_num),
    CONSTRAINT beats_type_check CHECK (type IN ('setup', 'turn', 'reveal', 'conflict', 'climax', 'resolution', 'hook', 'transition')),
    CONSTRAINT beats_order_positive CHECK (order_num > 0)
);

CREATE INDEX idx_beats_scene_id ON beats(scene_id);
CREATE INDEX idx_beats_scene_order ON beats(scene_id, order_num);
CREATE INDEX idx_beats_type ON beats(type);

