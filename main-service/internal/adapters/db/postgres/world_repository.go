package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.WorldRepository = (*WorldRepository)(nil)

// WorldRepository implements the world repository interface
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
		INSERT INTO worlds (id, tenant_id, name, description, genre, is_implicit, rpg_system_id, time_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	var timeConfigJSON []byte
	if w.TimeConfig != nil {
		var err error
		timeConfigJSON, err = json.Marshal(w.TimeConfig)
		if err != nil {
			return err
		}
	}
	_, err := r.db.Exec(ctx, query,
		w.ID, w.TenantID, w.Name, w.Description, w.Genre, w.IsImplicit, w.RPGSystemID, timeConfigJSON, w.CreatedAt, w.UpdatedAt)
	return err
}

// GetByID retrieves a world by ID
func (r *WorldRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.World, error) {
	query := `
		SELECT id, tenant_id, name, description, genre, is_implicit, rpg_system_id, time_config, created_at, updated_at
		FROM worlds
		WHERE tenant_id = $1 AND id = $2
	`
	var w world.World
	var rpgSystemID sql.NullString
	var timeConfigJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&w.ID, &w.TenantID, &w.Name, &w.Description, &w.Genre, &w.IsImplicit, &rpgSystemID, &timeConfigJSON, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "world",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if rpgSystemID.Valid {
		parsedID, err := uuid.Parse(rpgSystemID.String)
		if err == nil {
			w.RPGSystemID = &parsedID
		}
	}

	if len(timeConfigJSON) > 0 {
		var timeConfig world.TimeConfig
		if err := json.Unmarshal(timeConfigJSON, &timeConfig); err == nil {
			w.TimeConfig = &timeConfig
		}
	}

	return &w, nil
}

// ListByTenant lists worlds for a tenant
func (r *WorldRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.World, error) {
	query := `
		SELECT id, tenant_id, name, description, genre, is_implicit, rpg_system_id, time_config, created_at, updated_at
		FROM worlds
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
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
		SET name = $2, description = $3, genre = $4, is_implicit = $5, rpg_system_id = $6, time_config = $7, updated_at = $8
		WHERE tenant_id = $9 AND id = $1
	`
	var timeConfigJSON []byte
	if w.TimeConfig != nil {
		var err error
		timeConfigJSON, err = json.Marshal(w.TimeConfig)
		if err != nil {
			return err
		}
	}
	_, err := r.db.Exec(ctx, query, w.ID, w.Name, w.Description, w.Genre, w.IsImplicit, w.RPGSystemID, timeConfigJSON, w.UpdatedAt, w.TenantID)
	return err
}

// Delete deletes a world
func (r *WorldRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM worlds WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// CountByTenant counts worlds for a tenant
func (r *WorldRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM worlds WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

func (r *WorldRepository) scanWorlds(rows pgx.Rows) ([]*world.World, error) {
	worlds := make([]*world.World, 0)
	for rows.Next() {
		var w world.World
		var rpgSystemID sql.NullString
		var timeConfigJSON []byte

		err := rows.Scan(
			&w.ID, &w.TenantID, &w.Name, &w.Description, &w.Genre, &w.IsImplicit, &rpgSystemID, &timeConfigJSON, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if rpgSystemID.Valid {
			parsedID, err := uuid.Parse(rpgSystemID.String)
			if err == nil {
				w.RPGSystemID = &parsedID
			}
		}

		if len(timeConfigJSON) > 0 {
			var timeConfig world.TimeConfig
			if err := json.Unmarshal(timeConfigJSON, &timeConfig); err == nil {
				w.TimeConfig = &timeConfig
			}
		}

		worlds = append(worlds, &w)
	}

	return worlds, rows.Err()
}

