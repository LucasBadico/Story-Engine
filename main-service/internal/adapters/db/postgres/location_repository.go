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

var _ repositories.LocationRepository = (*LocationRepository)(nil)

// LocationRepository implements the location repository interface
type LocationRepository struct {
	db *DB
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *DB) *LocationRepository {
	return &LocationRepository{db: db}
}

// Create creates a new location
func (r *LocationRepository) Create(ctx context.Context, l *world.Location) error {
	// If parent_id is provided, get parent's level to calculate hierarchy_level
	if l.ParentID != nil {
		parent, err := r.GetByID(ctx, l.TenantID, *l.ParentID)
		if err != nil {
			return err
		}
		l.SetHierarchyLevel(parent.HierarchyLevel + 1)
	}

	query := `
		INSERT INTO locations (id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		l.ID, l.TenantID, l.WorldID, l.ParentID, l.Name, l.Type, l.Description, l.HierarchyLevel, l.CreatedAt, l.UpdatedAt)
	return err
}

// GetByID retrieves a location by ID
func (r *LocationRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Location, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1 AND id = $2
	`
	var l world.Location
	var parentID *uuid.UUID

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&l.ID, &l.TenantID, &l.WorldID, &parentID, &l.Name, &l.Type, &l.Description, &l.HierarchyLevel, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "location",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	l.ParentID = parentID
	return &l, nil
}

// ListByWorld lists locations for a world (flat list)
func (r *LocationRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, limit, offset int) ([]*world.Location, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1 AND world_id = $2
		ORDER BY hierarchy_level ASC, name ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, worldID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// ListByWorldTree lists locations for a world in tree structure (all locations, ordered by hierarchy)
func (r *LocationRepository) ListByWorldTree(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Location, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1 AND world_id = $2
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetChildren retrieves direct children of a location
func (r *LocationRepository) GetChildren(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.Location, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1 AND parent_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetAncestors retrieves all ancestors of a location (path to root)
func (r *LocationRepository) GetAncestors(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.Location, error) {
	// Use recursive CTE to get all ancestors
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
			FROM locations
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT l.id, l.tenant_id, l.world_id, l.parent_id, l.name, l.type, l.description, l.hierarchy_level, l.created_at, l.updated_at
			FROM locations l
			INNER JOIN ancestors a ON l.id = a.parent_id
			WHERE l.tenant_id = $1
		)
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM ancestors
		WHERE id != $2
		ORDER BY hierarchy_level ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetDescendants retrieves all descendants of a location (recursive)
func (r *LocationRepository) GetDescendants(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.Location, error) {
	// Use recursive CTE to get all descendants
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
			FROM locations
			WHERE tenant_id = $1 AND parent_id = $2
			UNION ALL
			SELECT l.id, l.tenant_id, l.world_id, l.parent_id, l.name, l.type, l.description, l.hierarchy_level, l.created_at, l.updated_at
			FROM locations l
			INNER JOIN descendants d ON l.parent_id = d.id
			WHERE l.tenant_id = $1
		)
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM descendants
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// Update updates a location
func (r *LocationRepository) Update(ctx context.Context, l *world.Location) error {
	query := `
		UPDATE locations
		SET name = $2, type = $3, description = $4, parent_id = $5, hierarchy_level = $6, updated_at = $7
		WHERE tenant_id = $8 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, l.ID, l.Name, l.Type, l.Description, l.ParentID, l.HierarchyLevel, l.UpdatedAt, l.TenantID)
	return err
}

// Delete deletes a location
func (r *LocationRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM locations WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// CountByWorld counts locations for a world
func (r *LocationRepository) CountByWorld(ctx context.Context, tenantID, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM locations WHERE tenant_id = $1 AND world_id = $2`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID, worldID).Scan(&count)
	return count, err
}

func (r *LocationRepository) scanLocations(rows pgx.Rows) ([]*world.Location, error) {
	locations := make([]*world.Location, 0)
	for rows.Next() {
		var l world.Location
		var parentID *uuid.UUID

		err := rows.Scan(
			&l.ID, &l.TenantID, &l.WorldID, &parentID, &l.Name, &l.Type, &l.Description, &l.HierarchyLevel, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}

		l.ParentID = parentID
		locations = append(locations, &l)
	}

	return locations, rows.Err()
}


