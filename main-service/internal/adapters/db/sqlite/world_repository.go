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

var _ repositories.WorldRepository = (*WorldRepository)(nil)

// WorldRepository implements the world repository interface for SQLite
type WorldRepository struct {
	db *DB
}

// NewWorldRepository creates a new world repository
func NewWorldRepository(db *DB) *WorldRepository {
	return &WorldRepository{db: db}
}

// Create creates a new world
func (r *WorldRepository) Create(ctx context.Context, w *world.World) error {
	query := `
		INSERT INTO worlds (id, tenant_id, name, description, genre, is_implicit, rpg_system_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var rpgSystemID sql.NullString
	if w.RPGSystemID != nil {
		rpgSystemID = sql.NullString{String: w.RPGSystemID.String(), Valid: true}
	}

	isImplicit := 0
	if w.IsImplicit {
		isImplicit = 1
	}

	_, err := r.db.Exec(ctx, query,
		w.ID.String(),
		w.TenantID.String(),
		w.Name,
		w.Description,
		w.Genre,
		isImplicit,
		rpgSystemID,
		w.CreatedAt.Format(time.RFC3339),
		w.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a world by ID
func (r *WorldRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.World, error) {
	query := `
		SELECT id, tenant_id, name, description, genre, is_implicit, rpg_system_id, created_at, updated_at
		FROM worlds
		WHERE tenant_id = ? AND id = ?
	`
	var w world.World
	var idStr, tenantIDStr, createdAtStr, updatedAtStr string
	var rpgSystemID sql.NullString
	var isImplicitInt int

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &w.Name, &w.Description, &w.Genre, &isImplicitInt, &rpgSystemID, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "world",
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
	w.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	w.TenantID = parsedTenantID

	// Parse timestamps
	w.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	w.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse boolean
	w.IsImplicit = isImplicitInt != 0

	// Parse nullable UUID
	if rpgSystemID.Valid {
		if parsedRPGSystemID, err := uuid.Parse(rpgSystemID.String); err == nil {
			w.RPGSystemID = &parsedRPGSystemID
		}
	}

	return &w, nil
}

// ListByTenant lists worlds for a tenant
func (r *WorldRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.World, error) {
	query := `
		SELECT id, tenant_id, name, description, genre, is_implicit, rpg_system_id, created_at, updated_at
		FROM worlds
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanWorlds(rows)
}

// Update updates a world
func (r *WorldRepository) Update(ctx context.Context, w *world.World) error {
	query := `
		UPDATE worlds
		SET name = ?, description = ?, genre = ?, is_implicit = ?, rpg_system_id = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var rpgSystemID sql.NullString
	if w.RPGSystemID != nil {
		rpgSystemID = sql.NullString{String: w.RPGSystemID.String(), Valid: true}
	}

	isImplicit := 0
	if w.IsImplicit {
		isImplicit = 1
	}

	_, err := r.db.Exec(ctx, query,
		w.Name,
		w.Description,
		w.Genre,
		isImplicit,
		rpgSystemID,
		w.UpdatedAt.Format(time.RFC3339),
		w.TenantID.String(),
		w.ID.String(),
	)
	return err
}

// Delete deletes a world
func (r *WorldRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM worlds WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByTenant counts worlds for a tenant
func (r *WorldRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM worlds WHERE tenant_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(&count)
	return count, err
}

func (r *WorldRepository) scanWorlds(rows *sql.Rows) ([]*world.World, error) {
	worlds := make([]*world.World, 0)
	for rows.Next() {
		var w world.World
		var idStr, tenantIDStr, createdAtStr, updatedAtStr string
		var rpgSystemID sql.NullString
		var isImplicitInt int

		err := rows.Scan(
			&idStr, &tenantIDStr, &w.Name, &w.Description, &w.Genre, &isImplicitInt, &rpgSystemID, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		w.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		w.TenantID = parsedTenantID

		// Parse timestamps
		w.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		w.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse boolean
		w.IsImplicit = isImplicitInt != 0

		// Parse nullable UUID
		if rpgSystemID.Valid {
			if parsedRPGSystemID, err := uuid.Parse(rpgSystemID.String); err == nil {
				w.RPGSystemID = &parsedRPGSystemID
			}
		}

		worlds = append(worlds, &w)
	}

	return worlds, rows.Err()
}

