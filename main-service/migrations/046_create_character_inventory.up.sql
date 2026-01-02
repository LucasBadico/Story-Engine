CREATE TABLE IF NOT EXISTS character_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    
    quantity INT DEFAULT 1,
    slot_id UUID REFERENCES inventory_slots(id) ON DELETE SET NULL,  -- null = backpack/bolsa
    is_equipped BOOLEAN DEFAULT FALSE,
    
    -- Customização por instância (item único do character)
    custom_name VARCHAR(100),               -- "Minha Espada Favorita"
    custom_stats JSONB,                     -- override de stats para este item específico
    
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(character_id, item_id, slot_id)
);

CREATE INDEX idx_character_inventory_character ON character_inventory(character_id);
CREATE INDEX idx_character_inventory_item ON character_inventory(item_id);
CREATE INDEX idx_character_inventory_slot ON character_inventory(slot_id);
CREATE INDEX idx_character_inventory_equipped ON character_inventory(character_id, is_equipped);


