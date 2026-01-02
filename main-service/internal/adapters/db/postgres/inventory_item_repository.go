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

var _ repositories.InventoryItemRepository = (*InventoryItemRepository)(nil)

// InventoryItemRepository implements the inventory item repository interface
type InventoryItemRepository struct {
	db *DB
}

// NewInventoryItemRepository creates a new inventory item repository
func NewInventoryItemRepository(db *DB) *InventoryItemRepository {
	return &InventoryItemRepository{db: db}
}

// Create creates a new inventory item
func (r *InventoryItemRepository) Create(ctx context.Context, item *rpg.InventoryItem) error {
	query := `
		INSERT INTO inventory_items (id, tenant_id, rpg_system_id, artifact_id, name, category, description, slots_required, weight, size, max_stack, equip_slots, requirements, item_stats, is_template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	var category *string
	var size *string
	if item.Category != nil {
		categoryStr := string(*item.Category)
		category = &categoryStr
	}
	if item.Size != nil {
		sizeStr := string(*item.Size)
		size = &sizeStr
	}

	_, err := r.db.Exec(ctx, query,
		item.ID, item.TenantID, item.RPGSystemID, item.ArtifactID, item.Name, category, item.Description,
		item.SlotsRequired, item.Weight, size, item.MaxStack,
		item.EquipSlots, item.Requirements, item.ItemStats,
		item.IsTemplate, item.CreatedAt, item.UpdatedAt)
	return err
}

// GetByID retrieves an inventory item by ID
func (r *InventoryItemRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.InventoryItem, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, artifact_id, name, category, description, slots_required, weight, size, max_stack, equip_slots, requirements, item_stats, is_template, created_at, updated_at
		FROM inventory_items
		WHERE tenant_id = $1 AND id = $2
	`
	var item rpg.InventoryItem
	var artifactID sql.NullString
	var category sql.NullString
	var description sql.NullString
	var weight sql.NullFloat64
	var size sql.NullString
	var equipSlots sql.NullString
	var requirements sql.NullString
	var itemStats sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&item.ID, &item.TenantID, &item.RPGSystemID, &artifactID, &item.Name, &category, &description,
		&item.SlotsRequired, &weight, &size, &item.MaxStack,
		&equipSlots, &requirements, &itemStats,
		&item.IsTemplate, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "inventory_item",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if artifactID.Valid {
		parsedID, err := uuid.Parse(artifactID.String)
		if err == nil {
			item.ArtifactID = &parsedID
		}
	}
	if category.Valid {
		cat := rpg.ItemCategory(category.String)
		item.Category = &cat
	}
	if description.Valid {
		item.Description = &description.String
	}
	if weight.Valid {
		item.Weight = &weight.Float64
	}
	if size.Valid {
		s := rpg.ItemSize(size.String)
		item.Size = &s
	}
	if equipSlots.Valid {
		slots := json.RawMessage(equipSlots.String)
		item.EquipSlots = &slots
	}
	if requirements.Valid {
		req := json.RawMessage(requirements.String)
		item.Requirements = &req
	}
	if itemStats.Valid {
		stats := json.RawMessage(itemStats.String)
		item.ItemStats = &stats
	}

	return &item, nil
}

// ListBySystem lists items for an RPG system
func (r *InventoryItemRepository) ListBySystem(ctx context.Context, tenantID, rpgSystemID uuid.UUID) ([]*rpg.InventoryItem, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, artifact_id, name, category, description, slots_required, weight, size, max_stack, equip_slots, requirements, item_stats, is_template, created_at, updated_at
		FROM inventory_items
		WHERE tenant_id = $1 AND rpg_system_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, rpgSystemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInventoryItems(rows)
}

// ListByArtifact lists items linked to an artifact
func (r *InventoryItemRepository) ListByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) ([]*rpg.InventoryItem, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, artifact_id, name, category, description, slots_required, weight, size, max_stack, equip_slots, requirements, item_stats, is_template, created_at, updated_at
		FROM inventory_items
		WHERE tenant_id = $1 AND artifact_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInventoryItems(rows)
}

// Update updates an inventory item
func (r *InventoryItemRepository) Update(ctx context.Context, item *rpg.InventoryItem) error {
	query := `
		UPDATE inventory_items
		SET artifact_id = $2, name = $3, category = $4, description = $5, slots_required = $6, weight = $7, size = $8, max_stack = $9, equip_slots = $10, requirements = $11, item_stats = $12, is_template = $13, updated_at = $14
		WHERE tenant_id = $15 AND id = $1
	`
	var category *string
	var size *string
	if item.Category != nil {
		categoryStr := string(*item.Category)
		category = &categoryStr
	}
	if item.Size != nil {
		sizeStr := string(*item.Size)
		size = &sizeStr
	}

	_, err := r.db.Exec(ctx, query,
		item.ID, item.ArtifactID, item.Name, category, item.Description,
		item.SlotsRequired, item.Weight, size, item.MaxStack,
		item.EquipSlots, item.Requirements, item.ItemStats,
		item.IsTemplate, item.UpdatedAt, item.TenantID)
	return err
}

// Delete deletes an inventory item
func (r *InventoryItemRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM inventory_items WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *InventoryItemRepository) scanInventoryItems(rows pgx.Rows) ([]*rpg.InventoryItem, error) {
	items := make([]*rpg.InventoryItem, 0)
	for rows.Next() {
		var item rpg.InventoryItem
		var artifactID sql.NullString
		var category sql.NullString
		var description sql.NullString
		var weight sql.NullFloat64
		var size sql.NullString
		var equipSlots sql.NullString
		var requirements sql.NullString
		var itemStats sql.NullString

		err := rows.Scan(
			&item.ID, &item.TenantID, &item.RPGSystemID, &artifactID, &item.Name, &category, &description,
			&item.SlotsRequired, &weight, &size, &item.MaxStack,
			&equipSlots, &requirements, &itemStats,
			&item.IsTemplate, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if artifactID.Valid {
			parsedID, err := uuid.Parse(artifactID.String)
			if err == nil {
				item.ArtifactID = &parsedID
			}
		}
		if category.Valid {
			cat := rpg.ItemCategory(category.String)
			item.Category = &cat
		}
		if description.Valid {
			item.Description = &description.String
		}
		if weight.Valid {
			item.Weight = &weight.Float64
		}
		if size.Valid {
			s := rpg.ItemSize(size.String)
			item.Size = &s
		}
		if equipSlots.Valid {
			slots := json.RawMessage(equipSlots.String)
			item.EquipSlots = &slots
		}
		if requirements.Valid {
			req := json.RawMessage(requirements.String)
			item.Requirements = &req
		}
		if itemStats.Valid {
			stats := json.RawMessage(itemStats.String)
			item.ItemStats = &stats
		}

		items = append(items, &item)
	}
	return items, rows.Err()
}


