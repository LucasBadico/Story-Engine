CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rpg_system_id UUID NOT NULL REFERENCES rpg_systems(id) ON DELETE CASCADE,
    artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL,  -- opcional: link ao lore/narrativa
    
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),                   -- weapon, armor, consumable, quest, misc
    description TEXT,
    
    -- Mecânicas de inventário
    slots_required INT DEFAULT 1,           -- quantos slots ocupa
    weight DECIMAL(10,2),                   -- peso em kg
    size VARCHAR(20),                       -- tiny, small, medium, large, huge
    max_stack INT DEFAULT 1,                -- 1 = não stackable, 99 = consumíveis
    
    -- Restrições de equipamento
    equip_slots JSONB,                      -- ["weapon", "off_hand"] - onde pode equipar
    requirements JSONB,                     -- {level: 5, class: "warrior", stats: {strength: 15}}
    
    -- Stats do item (para game engine)
    item_stats JSONB,                       -- {damage: "2d6", armor: 5, effects: [...]}
    
    is_template BOOLEAN DEFAULT FALSE,      -- true = template para criar cópias
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_items_rpg_system ON inventory_items(rpg_system_id);
CREATE INDEX idx_inventory_items_artifact ON inventory_items(artifact_id);
CREATE INDEX idx_inventory_items_category ON inventory_items(category);
CREATE INDEX idx_inventory_items_template ON inventory_items(is_template);


