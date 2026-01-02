CREATE TABLE IF NOT EXISTS inventory_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rpg_system_id UUID NOT NULL REFERENCES rpg_systems(id) ON DELETE CASCADE,
    
    name VARCHAR(50) NOT NULL,              -- head, chest, weapon, ring1, backpack
    slot_type VARCHAR(50),                  -- equipment, consumable, quest
    
    UNIQUE(rpg_system_id, name)
);

CREATE INDEX idx_inventory_slots_rpg_system ON inventory_slots(rpg_system_id);
CREATE INDEX idx_inventory_slots_type ON inventory_slots(slot_type);

