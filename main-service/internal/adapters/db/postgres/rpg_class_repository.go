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

var _ repositories.RPGClassRepository = (*RPGClassRepository)(nil)

// RPGClassRepository implements the RPG class repository interface
type RPGClassRepository struct {
	db *DB
}

// NewRPGClassRepository creates a new RPG class repository
func NewRPGClassRepository(db *DB) *RPGClassRepository {
	return &RPGClassRepository{db: db}
}

// Create creates a new RPG class
func (r *RPGClassRepository) Create(ctx context.Context, class *rpg.RPGClass) error {
	query := `
		INSERT INTO rpg_classes (id, tenant_id, rpg_system_id, parent_class_id, name, tier, description, requirements, stat_bonuses, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		class.ID, class.TenantID, class.RPGSystemID, class.ParentClassID, class.Name, class.Tier,
		class.Description, class.Requirements, class.StatBonuses,
		class.CreatedAt, class.UpdatedAt)
	return err
}

// GetByID retrieves an RPG class by ID
func (r *RPGClassRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.RPGClass, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, parent_class_id, name, tier, description, requirements, stat_bonuses, created_at, updated_at
		FROM rpg_classes
		WHERE tenant_id = $1 AND id = $2
	`
	var class rpg.RPGClass
	var parentClassID sql.NullString
	var description sql.NullString
	var requirements sql.NullString
	var statBonuses sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&class.ID, &class.TenantID, &class.RPGSystemID, &parentClassID, &class.Name, &class.Tier,
		&description, &requirements, &statBonuses,
		&class.CreatedAt, &class.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "rpg_class",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if parentClassID.Valid {
		parsedID, err := uuid.Parse(parentClassID.String)
		if err == nil {
			class.ParentClassID = &parsedID
		}
	}
	if description.Valid {
		class.Description = &description.String
	}
	if requirements.Valid {
		req := json.RawMessage(requirements.String)
		class.Requirements = &req
	}
	if statBonuses.Valid {
		bonuses := json.RawMessage(statBonuses.String)
		class.StatBonuses = &bonuses
	}

	return &class, nil
}

// ListBySystem lists classes for an RPG system
func (r *RPGClassRepository) ListBySystem(ctx context.Context, tenantID, rpgSystemID uuid.UUID) ([]*rpg.RPGClass, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, parent_class_id, name, tier, description, requirements, stat_bonuses, created_at, updated_at
		FROM rpg_classes
		WHERE tenant_id = $1 AND rpg_system_id = $2
		ORDER BY tier ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, rpgSystemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRPGClasses(rows)
}

// ListByParent lists classes that evolve from a parent class
func (r *RPGClassRepository) ListByParent(ctx context.Context, tenantID, parentClassID uuid.UUID) ([]*rpg.RPGClass, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, parent_class_id, name, tier, description, requirements, stat_bonuses, created_at, updated_at
		FROM rpg_classes
		WHERE tenant_id = $1 AND parent_class_id = $2
		ORDER BY tier ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, parentClassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRPGClasses(rows)
}

// Update updates an RPG class
func (r *RPGClassRepository) Update(ctx context.Context, class *rpg.RPGClass) error {
	query := `
		UPDATE rpg_classes
		SET parent_class_id = $2, name = $3, tier = $4, description = $5, requirements = $6, stat_bonuses = $7, updated_at = $8
		WHERE tenant_id = $9 AND id = $1
	`
	_, err := r.db.Exec(ctx, query,
		class.ID, class.ParentClassID, class.Name, class.Tier,
		class.Description, class.Requirements, class.StatBonuses,
		class.UpdatedAt, class.TenantID)
	return err
}

// Delete deletes an RPG class
func (r *RPGClassRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM rpg_classes WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "rpg_class",
			ID:       id.String(),
		}
	}
	return nil
}

func (r *RPGClassRepository) scanRPGClasses(rows pgx.Rows) ([]*rpg.RPGClass, error) {
	classes := make([]*rpg.RPGClass, 0)
	for rows.Next() {
		var class rpg.RPGClass
		var parentClassID sql.NullString
		var description sql.NullString
		var requirements sql.NullString
		var statBonuses sql.NullString

		err := rows.Scan(
			&class.ID, &class.TenantID, &class.RPGSystemID, &parentClassID, &class.Name, &class.Tier,
			&description, &requirements, &statBonuses,
			&class.CreatedAt, &class.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if parentClassID.Valid {
			parsedID, err := uuid.Parse(parentClassID.String)
			if err == nil {
				class.ParentClassID = &parsedID
			}
		}
		if description.Valid {
			class.Description = &description.String
		}
		if requirements.Valid {
			req := json.RawMessage(requirements.String)
			class.Requirements = &req
		}
		if statBonuses.Valid {
			bonuses := json.RawMessage(statBonuses.String)
			class.StatBonuses = &bonuses
		}

		classes = append(classes, &class)
	}
	return classes, rows.Err()
}


