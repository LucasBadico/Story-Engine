package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.InventorySlotRepository = (*InventorySlotRepository)(nil)

// InventorySlotRepository implements the inventory slot repository interface
type InventorySlotRepository struct {
	db *DB
}

// NewInventorySlotRepository creates a new inventory slot repository
func NewInventorySlotRepository(db *DB) *InventorySlotRepository {
	return &InventorySlotRepository{db: db}
}

// Create creates a new inventory slot
func (r *InventorySlotRepository) Create(ctx context.Context, slot *rpg.InventorySlot) error {
	query := `
		INSERT INTO inventory_slots (id, rpg_system_id, name, slot_type)
		VALUES ($1, $2, $3, $4)
	`
	var slotType *string
	if slot.SlotType != nil {
		typeStr := string(*slot.SlotType)
		slotType = &typeStr
	}

	_, err := r.db.Exec(ctx, query, slot.ID, slot.RPGSystemID, slot.Name, slotType)
	return err
}

// GetByID retrieves an inventory slot by ID
func (r *InventorySlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*rpg.InventorySlot, error) {
	query := `
		SELECT id, rpg_system_id, name, slot_type
		FROM inventory_slots
		WHERE id = $1
	`
	var slot rpg.InventorySlot
	var slotType sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(&slot.ID, &slot.RPGSystemID, &slot.Name, &slotType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "inventory_slot",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if slotType.Valid {
		st := rpg.SlotType(slotType.String)
		slot.SlotType = &st
	}

	return &slot, nil
}

// ListBySystem lists slots for an RPG system
func (r *InventorySlotRepository) ListBySystem(ctx context.Context, rpgSystemID uuid.UUID) ([]*rpg.InventorySlot, error) {
	query := `
		SELECT id, rpg_system_id, name, slot_type
		FROM inventory_slots
		WHERE rpg_system_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, rpgSystemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInventorySlots(rows)
}

// Update updates an inventory slot
func (r *InventorySlotRepository) Update(ctx context.Context, slot *rpg.InventorySlot) error {
	query := `
		UPDATE inventory_slots
		SET name = $2, slot_type = $3
		WHERE id = $1
	`
	var slotType *string
	if slot.SlotType != nil {
		typeStr := string(*slot.SlotType)
		slotType = &typeStr
	}

	_, err := r.db.Exec(ctx, query, slot.ID, slot.Name, slotType)
	return err
}

// Delete deletes an inventory slot
func (r *InventorySlotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM inventory_slots WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *InventorySlotRepository) scanInventorySlots(rows pgx.Rows) ([]*rpg.InventorySlot, error) {
	slots := make([]*rpg.InventorySlot, 0)
	for rows.Next() {
		var slot rpg.InventorySlot
		var slotType sql.NullString

		err := rows.Scan(&slot.ID, &slot.RPGSystemID, &slot.Name, &slotType)
		if err != nil {
			return nil, err
		}

		if slotType.Valid {
			st := rpg.SlotType(slotType.String)
			slot.SlotType = &st
		}

		slots = append(slots, &slot)
	}
	return slots, rows.Err()
}

