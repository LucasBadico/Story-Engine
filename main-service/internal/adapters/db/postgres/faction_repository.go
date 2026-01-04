package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.FactionRepository = (*FactionRepository)(nil)

// FactionRepository implements the faction repository interface
type FactionRepository struct {
	db *DB
}

// NewFactionRepository creates a new faction repository
func NewFactionRepository(db *DB) *FactionRepository {
	return &FactionRepository{db: db}
}

// Create creates a new faction
func (r *FactionRepository) Create(ctx context.Context, f *world.Faction) error {
	// If parent_id is provided, get parent's level to calculate hierarchy_level
	if f.ParentID != nil {
		parent, err := r.GetByID(ctx, f.TenantID, *f.ParentID)
		if err != nil {
			return err
		}
		f.SetHierarchyLevel(parent.HierarchyLevel + 1)
	}

	query := `
		INSERT INTO factions (id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		f.ID, f.TenantID, f.WorldID, f.ParentID, f.Name, f.Type, f.Description, f.Beliefs, f.Structure, f.Symbols, f.HierarchyLevel, f.CreatedAt, f.UpdatedAt)
	return err
}

// GetByID retrieves a faction by ID
func (r *FactionRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Faction, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at
		FROM factions
		WHERE tenant_id = $1 AND id = $2
	`
	var f world.Faction
	var parentID *uuid.UUID

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&f.ID, &f.TenantID, &f.WorldID, &parentID, &f.Name, &f.Type, &f.Description, &f.Beliefs, &f.Structure, &f.Symbols, &f.HierarchyLevel, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "faction",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	f.ParentID = parentID
	return &f, nil
}

// ListByWorld lists factions for a world
func (r *FactionRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Faction, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at
		FROM factions
		WHERE tenant_id = $1 AND world_id = $2
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactions(rows)
}

// Update updates a faction
func (r *FactionRepository) Update(ctx context.Context, f *world.Faction) error {
	query := `
		UPDATE factions
		SET name = $2, type = $3, description = $4, beliefs = $5, structure = $6, symbols = $7, parent_id = $8, hierarchy_level = $9, updated_at = $10
		WHERE tenant_id = $11 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, f.ID, f.Name, f.Type, f.Description, f.Beliefs, f.Structure, f.Symbols, f.ParentID, f.HierarchyLevel, f.UpdatedAt, f.TenantID)
	return err
}

// Delete deletes a faction
func (r *FactionRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM factions WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// GetChildren retrieves direct children of a faction
func (r *FactionRepository) GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Faction, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at
		FROM factions
		WHERE tenant_id = $1 AND parent_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactions(rows)
}

func (r *FactionRepository) scanFactions(rows pgx.Rows) ([]*world.Faction, error) {
	factions := make([]*world.Faction, 0)
	for rows.Next() {
		var f world.Faction
		var parentID *uuid.UUID

		err := rows.Scan(
			&f.ID, &f.TenantID, &f.WorldID, &parentID, &f.Name, &f.Type, &f.Description, &f.Beliefs, &f.Structure, &f.Symbols, &f.HierarchyLevel, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}

		f.ParentID = parentID
		factions = append(factions, &f)
	}

	return factions, rows.Err()
}

