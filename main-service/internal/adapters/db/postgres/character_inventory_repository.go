package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.CharacterInventoryRepository = (*CharacterInventoryRepository)(nil)

// CharacterInventoryRepository implements the character inventory repository interface
type CharacterInventoryRepository struct {
	db *DB
}

// NewCharacterInventoryRepository creates a new character inventory repository
func NewCharacterInventoryRepository(db *DB) *CharacterInventoryRepository {
	return &CharacterInventoryRepository{db: db}
}

// Create creates a new character inventory entry
func (r *CharacterInventoryRepository) Create(ctx context.Context, inventory *rpg.CharacterInventory) error {
	query := `
		INSERT INTO character_inventory (id, tenant_id, character_id, item_id, quantity, slot_id, is_equipped, custom_name, custom_stats, acquired_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		inventory.ID, inventory.TenantID, inventory.CharacterID, inventory.ItemID, inventory.Quantity,
		inventory.SlotID, inventory.IsEquipped, inventory.CustomName, inventory.CustomStats,
		inventory.AcquiredAt)
	return err
}

// GetByID retrieves a character inventory entry by ID
func (r *CharacterInventoryRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.CharacterInventory, error) {
	query := `
		SELECT id, tenant_id, character_id, item_id, quantity, slot_id, is_equipped, custom_name, custom_stats, acquired_at
		FROM character_inventory
		WHERE tenant_id = $1 AND id = $2
	`
	var ci rpg.CharacterInventory
	var slotID sql.NullString
	var customName sql.NullString
	var customStats sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&ci.ID, &ci.TenantID, &ci.CharacterID, &ci.ItemID, &ci.Quantity, &slotID,
		&ci.IsEquipped, &customName, &customStats, &ci.AcquiredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_inventory",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if slotID.Valid {
		parsedID, err := uuid.Parse(slotID.String)
		if err == nil {
			ci.SlotID = &parsedID
		}
	}
	if customName.Valid {
		ci.CustomName = &customName.String
	}
	if customStats.Valid {
		stats := json.RawMessage(customStats.String)
		ci.CustomStats = &stats
	}

	return &ci, nil
}

// GetByCharacterAndItem retrieves a character inventory entry by character, item, and optional slot
func (r *CharacterInventoryRepository) GetByCharacterAndItem(ctx context.Context, tenantID, characterID, itemID uuid.UUID, slotID *uuid.UUID) (*rpg.CharacterInventory, error) {
	var query string
	var args []interface{}

	if slotID != nil {
		query = `
			SELECT id, tenant_id, character_id, item_id, quantity, slot_id, is_equipped, custom_name, custom_stats, acquired_at
			FROM character_inventory
			WHERE tenant_id = $1 AND character_id = $2 AND item_id = $3 AND slot_id = $4
		`
		args = []interface{}{tenantID, characterID, itemID, slotID}
	} else {
		query = `
			SELECT id, tenant_id, character_id, item_id, quantity, slot_id, is_equipped, custom_name, custom_stats, acquired_at
			FROM character_inventory
			WHERE tenant_id = $1 AND character_id = $2 AND item_id = $3 AND slot_id IS NULL
		`
		args = []interface{}{tenantID, characterID, itemID}
	}

	var ci rpg.CharacterInventory
	var slotIDNull sql.NullString
	var customName sql.NullString
	var customStats sql.NullString

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&ci.ID, &ci.TenantID, &ci.CharacterID, &ci.ItemID, &ci.Quantity, &slotIDNull,
		&ci.IsEquipped, &customName, &customStats, &ci.AcquiredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_inventory",
				ID:       characterID.String() + "/" + itemID.String(),
			}
		}
		return nil, err
	}

	if slotIDNull.Valid {
		parsedID, err := uuid.Parse(slotIDNull.String)
		if err == nil {
			ci.SlotID = &parsedID
		}
	}
	if customName.Valid {
		ci.CustomName = &customName.String
	}
	if customStats.Valid {
		stats := json.RawMessage(customStats.String)
		ci.CustomStats = &stats
	}

	return &ci, nil
}

// ListByCharacter lists all inventory entries for a character
func (r *CharacterInventoryRepository) ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*rpg.CharacterInventory, error) {
	query := `
		SELECT id, tenant_id, character_id, item_id, quantity, slot_id, is_equipped, custom_name, custom_stats, acquired_at
		FROM character_inventory
		WHERE tenant_id = $1 AND character_id = $2
		ORDER BY is_equipped DESC, acquired_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterInventory(rows)
}

// ListEquippedByCharacter lists equipped items for a character
func (r *CharacterInventoryRepository) ListEquippedByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*rpg.CharacterInventory, error) {
	query := `
		SELECT id, tenant_id, character_id, item_id, quantity, slot_id, is_equipped, custom_name, custom_stats, acquired_at
		FROM character_inventory
		WHERE tenant_id = $1 AND character_id = $2 AND is_equipped = TRUE
		ORDER BY acquired_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterInventory(rows)
}

// Update updates a character inventory entry
func (r *CharacterInventoryRepository) Update(ctx context.Context, inventory *rpg.CharacterInventory) error {
	query := `
		UPDATE character_inventory
		SET quantity = $2, slot_id = $3, is_equipped = $4, custom_name = $5, custom_stats = $6
		WHERE tenant_id = $7 AND id = $1
	`
	_, err := r.db.Exec(ctx, query,
		inventory.ID, inventory.Quantity, inventory.SlotID, inventory.IsEquipped,
		inventory.CustomName, inventory.CustomStats, inventory.TenantID)
	return err
}

// Delete deletes a character inventory entry
func (r *CharacterInventoryRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_inventory WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByCharacter deletes all inventory entries for a character
func (r *CharacterInventoryRepository) DeleteByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) error {
	query := `DELETE FROM character_inventory WHERE tenant_id = $1 AND character_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, characterID)
	return err
}

func (r *CharacterInventoryRepository) scanCharacterInventory(rows pgx.Rows) ([]*rpg.CharacterInventory, error) {
	items := make([]*rpg.CharacterInventory, 0)
	for rows.Next() {
		var ci rpg.CharacterInventory
		var slotID sql.NullString
		var customName sql.NullString
		var customStats sql.NullString

		err := rows.Scan(
			&ci.ID, &ci.TenantID, &ci.CharacterID, &ci.ItemID, &ci.Quantity, &slotID,
			&ci.IsEquipped, &customName, &customStats, &ci.AcquiredAt)
		if err != nil {
			return nil, err
		}

		if slotID.Valid {
			parsedID, err := uuid.Parse(slotID.String)
			if err == nil {
				ci.SlotID = &parsedID
			}
		}
		if customName.Valid {
			ci.CustomName = &customName.String
		}
		if customStats.Valid {
			stats := json.RawMessage(customStats.String)
			ci.CustomStats = &stats
		}

		items = append(items, &ci)
	}
	return items, rows.Err()
}


