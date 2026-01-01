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
		parent, err := r.GetByID(ctx, *l.ParentID)
		if err != nil {
			return err
		}
		l.SetHierarchyLevel(parent.HierarchyLevel + 1)
	}

	query := `
		INSERT INTO locations (id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		l.ID, l.WorldID, l.ParentID, l.Name, l.Type, l.Description, l.HierarchyLevel, l.CreatedAt, l.UpdatedAt)
	return err
}

// GetByID retrieves a location by ID
func (r *LocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.Location, error) {
	query := `
		SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE id = $1
	`
	var l world.Location
	var parentID *uuid.UUID

	err := r.db.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.WorldID, &parentID, &l.Name, &l.Type, &l.Description, &l.HierarchyLevel, &l.CreatedAt, &l.UpdatedAt)
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
func (r *LocationRepository) ListByWorld(ctx context.Context, worldID uuid.UUID, limit, offset int) ([]*world.Location, error) {
	query := `
		SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE world_id = $1
		ORDER BY hierarchy_level ASC, name ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, worldID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// ListByWorldTree lists locations for a world in tree structure (all locations, ordered by hierarchy)
func (r *LocationRepository) ListByWorldTree(ctx context.Context, worldID uuid.UUID) ([]*world.Location, error) {
	query := `
		SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE world_id = $1
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetChildren retrieves direct children of a location
func (r *LocationRepository) GetChildren(ctx context.Context, locationID uuid.UUID) ([]*world.Location, error) {
	query := `
		SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE parent_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetAncestors retrieves all ancestors of a location (path to root)
func (r *LocationRepository) GetAncestors(ctx context.Context, locationID uuid.UUID) ([]*world.Location, error) {
	// Use recursive CTE to get all ancestors
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
			FROM locations
			WHERE id = $1
			UNION ALL
			SELECT l.id, l.world_id, l.parent_id, l.name, l.type, l.description, l.hierarchy_level, l.created_at, l.updated_at
			FROM locations l
			INNER JOIN ancestors a ON l.id = a.parent_id
		)
		SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM ancestors
		WHERE id != $1
		ORDER BY hierarchy_level ASC
	`
	rows, err := r.db.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetDescendants retrieves all descendants of a location (recursive)
func (r *LocationRepository) GetDescendants(ctx context.Context, locationID uuid.UUID) ([]*world.Location, error) {
	// Use recursive CTE to get all descendants
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
			FROM locations
			WHERE parent_id = $1
			UNION ALL
			SELECT l.id, l.world_id, l.parent_id, l.name, l.type, l.description, l.hierarchy_level, l.created_at, l.updated_at
			FROM locations l
			INNER JOIN descendants d ON l.parent_id = d.id
		)
		SELECT id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM descendants
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, locationID)
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
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, l.ID, l.Name, l.Type, l.Description, l.ParentID, l.HierarchyLevel, l.UpdatedAt)
	return err
}

// Delete deletes a location
func (r *LocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM locations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByWorld counts locations for a world
func (r *LocationRepository) CountByWorld(ctx context.Context, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM locations WHERE world_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, worldID).Scan(&count)
	return count, err
}

func (r *LocationRepository) scanLocations(rows pgx.Rows) ([]*world.Location, error) {
	locations := make([]*world.Location, 0)
	for rows.Next() {
		var l world.Location
		var parentID *uuid.UUID

		err := rows.Scan(
			&l.ID, &l.WorldID, &parentID, &l.Name, &l.Type, &l.Description, &l.HierarchyLevel, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}

		l.ParentID = parentID
		locations = append(locations, &l)
	}

	return locations, rows.Err()
}

