package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.LocationRepository = (*LocationRepository)(nil)

// LocationRepository implements the location repository interface for SQLite
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var parentID sql.NullString
	if l.ParentID != nil {
		parentID = sql.NullString{String: l.ParentID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		l.ID.String(),
		l.TenantID.String(),
		l.WorldID.String(),
		parentID,
		l.Name,
		l.Type,
		l.Description,
		l.HierarchyLevel,
		l.CreatedAt.Format(time.RFC3339),
		l.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a location by ID
func (r *LocationRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Location, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE tenant_id = ? AND id = ?
	`
	var l world.Location
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var parentID sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &parentID, &l.Name, &l.Type, &l.Description, &l.HierarchyLevel, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "location",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Parse UUIDs
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	l.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	l.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	l.WorldID = parsedWorldID

	// Parse timestamps
	l.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	l.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable UUID
	if parentID.Valid {
		if parsedParentID, err := uuid.Parse(parentID.String); err == nil {
			l.ParentID = &parsedParentID
		}
	}

	return &l, nil
}

// ListByWorld lists locations for a world (flat list)
func (r *LocationRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, limit, offset int) ([]*world.Location, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM locations
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY hierarchy_level ASC, name ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String(), limit, offset)
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
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String())
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
		WHERE tenant_id = ? AND parent_id = ?
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), locationID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetAncestors retrieves all ancestors of a location (path to root)
func (r *LocationRepository) GetAncestors(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.Location, error) {
	// Use recursive CTE to get all ancestors (SQLite supports WITH RECURSIVE)
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
			FROM locations
			WHERE tenant_id = ? AND id = ?
			UNION ALL
			SELECT l.id, l.tenant_id, l.world_id, l.parent_id, l.name, l.type, l.description, l.hierarchy_level, l.created_at, l.updated_at
			FROM locations l
			INNER JOIN ancestors a ON l.id = a.parent_id
			WHERE l.tenant_id = ?
		)
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM ancestors
		WHERE id != ?
		ORDER BY hierarchy_level ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), locationID.String(), tenantID.String(), locationID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLocations(rows)
}

// GetDescendants retrieves all descendants of a location (recursive)
func (r *LocationRepository) GetDescendants(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.Location, error) {
	// Use recursive CTE to get all descendants (SQLite supports WITH RECURSIVE)
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
			FROM locations
			WHERE tenant_id = ? AND parent_id = ?
			UNION ALL
			SELECT l.id, l.tenant_id, l.world_id, l.parent_id, l.name, l.type, l.description, l.hierarchy_level, l.created_at, l.updated_at
			FROM locations l
			INNER JOIN descendants d ON l.parent_id = d.id
			WHERE l.tenant_id = ?
		)
		SELECT id, tenant_id, world_id, parent_id, name, type, description, hierarchy_level, created_at, updated_at
		FROM descendants
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), locationID.String(), tenantID.String())
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
		SET name = ?, type = ?, description = ?, parent_id = ?, hierarchy_level = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var parentID sql.NullString
	if l.ParentID != nil {
		parentID = sql.NullString{String: l.ParentID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		l.Name,
		l.Type,
		l.Description,
		parentID,
		l.HierarchyLevel,
		l.UpdatedAt.Format(time.RFC3339),
		l.TenantID.String(),
		l.ID.String(),
	)
	return err
}

// Delete deletes a location
func (r *LocationRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM locations WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByWorld counts locations for a world
func (r *LocationRepository) CountByWorld(ctx context.Context, tenantID, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM locations WHERE tenant_id = ? AND world_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String(), worldID.String()).Scan(&count)
	return count, err
}

func (r *LocationRepository) scanLocations(rows *sql.Rows) ([]*world.Location, error) {
	locations := make([]*world.Location, 0)
	for rows.Next() {
		var l world.Location
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
		var parentID sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &parentID, &l.Name, &l.Type, &l.Description, &l.HierarchyLevel, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		l.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		l.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		l.WorldID = parsedWorldID

		// Parse timestamps
		l.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		l.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable UUID
		if parentID.Valid {
			if parsedParentID, err := uuid.Parse(parentID.String); err == nil {
				l.ParentID = &parsedParentID
			}
		}

		locations = append(locations, &l)
	}

	return locations, rows.Err()
}

