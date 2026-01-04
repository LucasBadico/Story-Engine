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

var _ repositories.LoreRepository = (*LoreRepository)(nil)

// LoreRepository implements the lore repository interface for SQLite
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var parentID sql.NullString
	if l.ParentID != nil {
		parentID = sql.NullString{String: l.ParentID.String(), Valid: true}
	}

	var category sql.NullString
	if l.Category != nil {
		category = sql.NullString{String: *l.Category, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		l.ID.String(),
		l.TenantID.String(),
		l.WorldID.String(),
		parentID,
		l.Name,
		category,
		l.Description,
		l.Rules,
		l.Limitations,
		l.Requirements,
		l.HierarchyLevel,
		l.CreatedAt.Format(time.RFC3339),
		l.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a lore by ID
func (r *LoreRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Lore, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at
		FROM lores
		WHERE tenant_id = ? AND id = ?
	`
	var l world.Lore
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var parentID, category sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &parentID, &l.Name, &category, &l.Description, &l.Rules, &l.Limitations, &l.Requirements, &l.HierarchyLevel, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "lore",
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

	// Parse nullable string
	if category.Valid {
		l.Category = &category.String
	}

	return &l, nil
}

// ListByWorld lists lores for a world
func (r *LoreRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Lore, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at
		FROM lores
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String())
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
		SET name = ?, category = ?, description = ?, rules = ?, limitations = ?, requirements = ?, parent_id = ?, hierarchy_level = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var parentID sql.NullString
	if l.ParentID != nil {
		parentID = sql.NullString{String: l.ParentID.String(), Valid: true}
	}

	var category sql.NullString
	if l.Category != nil {
		category = sql.NullString{String: *l.Category, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		l.Name,
		category,
		l.Description,
		l.Rules,
		l.Limitations,
		l.Requirements,
		parentID,
		l.HierarchyLevel,
		l.UpdatedAt.Format(time.RFC3339),
		l.TenantID.String(),
		l.ID.String(),
	)
	return err
}

// Delete deletes a lore
func (r *LoreRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM lores WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// GetChildren retrieves direct children of a lore
func (r *LoreRepository) GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Lore, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, category, description, rules, limitations, requirements, hierarchy_level, created_at, updated_at
		FROM lores
		WHERE tenant_id = ? AND parent_id = ?
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), parentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLores(rows)
}

func (r *LoreRepository) scanLores(rows *sql.Rows) ([]*world.Lore, error) {
	lores := make([]*world.Lore, 0)
	for rows.Next() {
		var l world.Lore
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
		var parentID, category sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &parentID, &l.Name, &category, &l.Description, &l.Rules, &l.Limitations, &l.Requirements, &l.HierarchyLevel, &createdAtStr, &updatedAtStr)
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

		// Parse nullable string
		if category.Valid {
			l.Category = &category.String
		}

		lores = append(lores, &l)
	}

	return lores, rows.Err()
}

