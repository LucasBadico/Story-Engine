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

var _ repositories.RPGSystemRepository = (*RPGSystemRepository)(nil)

// RPGSystemRepository implements the RPG system repository interface
type RPGSystemRepository struct {
	db *DB
}

// NewRPGSystemRepository creates a new RPG system repository
func NewRPGSystemRepository(db *DB) *RPGSystemRepository {
	return &RPGSystemRepository{db: db}
}

// Create creates a new RPG system
func (r *RPGSystemRepository) Create(ctx context.Context, system *rpg.RPGSystem) error {
	query := `
		INSERT INTO rpg_systems (id, tenant_id, name, description, base_stats_schema, derived_stats_schema, progression_schema, is_builtin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		system.ID, system.TenantID, system.Name, system.Description,
		system.BaseStatsSchema, system.DerivedStatsSchema, system.ProgressionSchema,
		system.IsBuiltin, system.CreatedAt, system.UpdatedAt)
	return err
}

// GetByID retrieves an RPG system by ID
func (r *RPGSystemRepository) GetByID(ctx context.Context, id uuid.UUID) (*rpg.RPGSystem, error) {
	query := `
		SELECT id, tenant_id, name, description, base_stats_schema, derived_stats_schema, progression_schema, is_builtin, created_at, updated_at
		FROM rpg_systems
		WHERE id = $1
	`
	var system rpg.RPGSystem
	var tenantID sql.NullString
	var description sql.NullString
	var derivedStatsSchema sql.NullString
	var progressionSchema sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&system.ID, &tenantID, &system.Name, &description,
		&system.BaseStatsSchema, &derivedStatsSchema, &progressionSchema,
		&system.IsBuiltin, &system.CreatedAt, &system.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "rpg_system",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if tenantID.Valid {
		parsedID, err := uuid.Parse(tenantID.String)
		if err == nil {
			system.TenantID = &parsedID
		}
	}
	if description.Valid {
		system.Description = &description.String
	}
	if derivedStatsSchema.Valid {
		derived := json.RawMessage(derivedStatsSchema.String)
		system.DerivedStatsSchema = &derived
	}
	if progressionSchema.Valid {
		progression := json.RawMessage(progressionSchema.String)
		system.ProgressionSchema = &progression
	}

	return &system, nil
}

// List lists RPG systems
func (r *RPGSystemRepository) List(ctx context.Context, tenantID *uuid.UUID) ([]*rpg.RPGSystem, error) {
	var query string
	var args []interface{}

	if tenantID == nil {
		// List builtin systems only
		query = `
			SELECT id, tenant_id, name, description, base_stats_schema, derived_stats_schema, progression_schema, is_builtin, created_at, updated_at
			FROM rpg_systems
			WHERE is_builtin = TRUE
			ORDER BY name ASC
		`
		args = []interface{}{}
	} else {
		// List builtin + tenant custom systems
		query = `
			SELECT id, tenant_id, name, description, base_stats_schema, derived_stats_schema, progression_schema, is_builtin, created_at, updated_at
			FROM rpg_systems
			WHERE is_builtin = TRUE OR tenant_id = $1
			ORDER BY is_builtin DESC, name ASC
		`
		args = []interface{}{*tenantID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRPGSystems(rows)
}

// Update updates an RPG system
func (r *RPGSystemRepository) Update(ctx context.Context, system *rpg.RPGSystem) error {
	query := `
		UPDATE rpg_systems
		SET name = $2, description = $3, base_stats_schema = $4, derived_stats_schema = $5, progression_schema = $6, updated_at = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		system.ID, system.Name, system.Description,
		system.BaseStatsSchema, system.DerivedStatsSchema, system.ProgressionSchema,
		system.UpdatedAt)
	return err
}

// Delete deletes an RPG system
func (r *RPGSystemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM rpg_systems WHERE id = $1 AND is_builtin = FALSE`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByTenant counts RPG systems for a tenant
func (r *RPGSystemRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM rpg_systems WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

func (r *RPGSystemRepository) scanRPGSystems(rows pgx.Rows) ([]*rpg.RPGSystem, error) {
	systems := make([]*rpg.RPGSystem, 0)
	for rows.Next() {
		var system rpg.RPGSystem
		var tenantID sql.NullString
		var description sql.NullString
		var derivedStatsSchema sql.NullString
		var progressionSchema sql.NullString

		err := rows.Scan(
			&system.ID, &tenantID, &system.Name, &description,
			&system.BaseStatsSchema, &derivedStatsSchema, &progressionSchema,
			&system.IsBuiltin, &system.CreatedAt, &system.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if tenantID.Valid {
			parsedID, err := uuid.Parse(tenantID.String)
			if err == nil {
				system.TenantID = &parsedID
			}
		}
		if description.Valid {
			system.Description = &description.String
		}
		if derivedStatsSchema.Valid {
			derived := json.RawMessage(derivedStatsSchema.String)
			system.DerivedStatsSchema = &derived
		}
		if progressionSchema.Valid {
			progression := json.RawMessage(progressionSchema.String)
			system.ProgressionSchema = &progression
		}

		systems = append(systems, &system)
	}
	return systems, rows.Err()
}

