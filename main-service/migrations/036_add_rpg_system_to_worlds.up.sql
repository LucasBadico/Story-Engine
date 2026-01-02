ALTER TABLE worlds ADD COLUMN rpg_system_id UUID REFERENCES rpg_systems(id) ON DELETE SET NULL;

CREATE INDEX idx_worlds_rpg_system_id ON worlds(rpg_system_id);


