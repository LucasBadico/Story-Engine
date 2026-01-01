ALTER TABLE characters ADD COLUMN current_class_id UUID REFERENCES rpg_classes(id) ON DELETE SET NULL;
ALTER TABLE characters ADD COLUMN class_level INT DEFAULT 1;

CREATE INDEX idx_characters_current_class ON characters(current_class_id);

