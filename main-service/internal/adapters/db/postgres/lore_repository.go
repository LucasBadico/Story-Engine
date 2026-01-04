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

var _ repositories.LoreRepository = (*LoreRepository)(nil)

// LoreRepository implements the lore repository interface
type LoreRepository struct {
	db *DB
}

// NewLoreRepository creates a new lore repository
func NewLoreRepository(db *DB) *LoreRepository {
	return &LoreRepository{db: db}
}

// Create creates a new lore
func (r *LoreRepository) Create(ctx context.Context, l *world.Lore) error {
	// If parent_id is provided, get parent's level to calculate hierarchy_level
	if l.ParentID != nil {
		parent, err := r.GetByID(ctx, l.TenantID, *l.ParentID)
		if err != nil {
			return err
		}
		l.SetHierarchyLevel(parent.HierarchyLevel + 1)
	}

	query := `
		INSERT INTO lores (id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		l.ID, l.TenantID, l.WorldID, l.ParentID, l.Name, l.Category, l.Description, l.Rules, l.Limitations, l.Requirements, l.HierarchyLevel, l.CreatedAt, l.UpdatedAt)
	return err
}

// GetByID retrieves a lore by ID
func (r *LoreRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Lore, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at
		FROM lores
		WHERE tenant_id = $1 AND id = $2
	`
	var l world.Lore
	var parentID *uuid.UUID

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&l.ID, &l.TenantID, &l.WorldID, &parentID, &l.Name, &l.Category, &l.Description, &l.Rules, &l.Limitations, &l.Requirements, &l.HierarchyLevel, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "lore",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	l.ParentID = parentID
	return &l, nil
}

// ListByWorld lists lores for a world
func (r *LoreRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Lore, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at
		FROM lores
		WHERE tenant_id = $1 AND world_id = $2
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLores(rows)
}

// Update updates a lore
func (r *LoreRepository) Update(ctx context.Context, l *world.Lore) error {
	query := `
		UPDATE lores
		SET name = $2, category = $3, description = $4, rules = $5, limitations = $6, requirements = $7, parent_id = $8, hierarchy_level = $9, updated_at = $10
		WHERE tenant_id = $11 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, l.ID, l.Name, l.Category, l.Description, l.Rules, l.Limitations, l.Requirements, l.ParentID, l.HierarchyLevel, l.UpdatedAt, l.TenantID)
	return err
}

// Delete deletes a lore
func (r *LoreRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM lores WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// GetChildren retrieves direct children of a lore
func (r *LoreRepository) GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Lore, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at
		FROM lores
		WHERE tenant_id = $1 AND parent_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLores(rows)
}

func (r *LoreRepository) scanLores(rows pgx.Rows) ([]*world.Lore, error) {
	lores := make([]*world.Lore, 0)
	for rows.Next() {
		var l world.Lore
		var parentID *uuid.UUID

		err := rows.Scan(
			&l.ID, &l.TenantID, &l.WorldID, &parentID, &l.Name, &l.Category, &l.Description, &l.Rules, &l.Limitations, &l.Requirements, &l.HierarchyLevel, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}

		l.ParentID = parentID
		lores = append(lores, &l)
	}

	return lores, rows.Err()
}

