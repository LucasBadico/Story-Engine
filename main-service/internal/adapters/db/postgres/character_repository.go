package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.CharacterRepository = (*CharacterRepository)(nil)

// CharacterRepository implements the character repository interface
type CharacterRepository struct {
	db *DB
}

// NewCharacterRepository creates a new character repository
func NewCharacterRepository(db *DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// Create creates a new character
func (r *CharacterRepository) Create(ctx context.Context, c *world.Character) error {
	query := `
		INSERT INTO characters (id, tenant_id, world_id, archetype_id, current_class_id, class_level, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		c.ID, c.TenantID, c.WorldID, c.ArchetypeID, c.CurrentClassID, c.ClassLevel, c.Name, c.Description, c.CreatedAt, c.UpdatedAt)
	return err
}

// GetByID retrieves a character by ID
func (r *CharacterRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Character, error) {
	query := `
		SELECT id, tenant_id, world_id, archetype_id, current_class_id, class_level, name, description, created_at, updated_at
		FROM characters
		WHERE tenant_id = $1 AND id = $2
	`
	var c world.Character
	var archetypeID sql.NullString
	var currentClassID sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&c.ID, &c.TenantID, &c.WorldID, &archetypeID, &currentClassID, &c.ClassLevel, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if archetypeID.Valid {
		parsedID, err := uuid.Parse(archetypeID.String)
		if err == nil {
			c.ArchetypeID = &parsedID
		}
	}
	if currentClassID.Valid {
		parsedID, err := uuid.Parse(currentClassID.String)
		if err == nil {
			c.CurrentClassID = &parsedID
		}
	}
	return &c, nil
}

// ListByWorld lists characters for a world
func (r *CharacterRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, limit, offset int) ([]*world.Character, error) {
	query := `
		SELECT id, tenant_id, world_id, archetype_id, current_class_id, class_level, name, description, created_at, updated_at
		FROM characters
		WHERE tenant_id = $1 AND world_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, worldID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacters(rows)
}

// Update updates a character
func (r *CharacterRepository) Update(ctx context.Context, c *world.Character) error {
	query := `
		UPDATE characters
		SET name = $2, description = $3, archetype_id = $4, current_class_id = $5, class_level = $6, updated_at = $7
		WHERE tenant_id = $8 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.Name, c.Description, c.ArchetypeID, c.CurrentClassID, c.ClassLevel, c.UpdatedAt, c.TenantID)
	return err
}

// Delete deletes a character
func (r *CharacterRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM characters WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// CountByWorld counts characters for a world
func (r *CharacterRepository) CountByWorld(ctx context.Context, tenantID, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM characters WHERE tenant_id = $1 AND world_id = $2`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID, worldID).Scan(&count)
	return count, err
}

func (r *CharacterRepository) scanCharacters(rows pgx.Rows) ([]*world.Character, error) {
	characters := make([]*world.Character, 0)
	for rows.Next() {
		var c world.Character
		var archetypeID sql.NullString
		var currentClassID sql.NullString

		err := rows.Scan(
			&c.ID, &c.TenantID, &c.WorldID, &archetypeID, &currentClassID, &c.ClassLevel, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if archetypeID.Valid {
			parsedID, err := uuid.Parse(archetypeID.String)
			if err == nil {
				c.ArchetypeID = &parsedID
			}
		}
		if currentClassID.Valid {
			parsedID, err := uuid.Parse(currentClassID.String)
			if err == nil {
				c.CurrentClassID = &parsedID
			}
		}
		characters = append(characters, &c)
	}

	return characters, rows.Err()
}

