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
		INSERT INTO characters (id, world_id, archetype_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		c.ID, c.WorldID, c.ArchetypeID, c.Name, c.Description, c.CreatedAt, c.UpdatedAt)
	return err
}

// GetByID retrieves a character by ID
func (r *CharacterRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.Character, error) {
	query := `
		SELECT id, world_id, archetype_id, name, description, created_at, updated_at
		FROM characters
		WHERE id = $1
	`
	var c world.Character
	var archetypeID *uuid.UUID

	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.WorldID, &archetypeID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	c.ArchetypeID = archetypeID
	return &c, nil
}

// ListByWorld lists characters for a world
func (r *CharacterRepository) ListByWorld(ctx context.Context, worldID uuid.UUID, limit, offset int) ([]*world.Character, error) {
	query := `
		SELECT id, world_id, archetype_id, name, description, created_at, updated_at
		FROM characters
		WHERE world_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, worldID, limit, offset)
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
		SET name = $2, description = $3, archetype_id = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.Name, c.Description, c.ArchetypeID, c.UpdatedAt)
	return err
}

// Delete deletes a character
func (r *CharacterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM characters WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByWorld counts characters for a world
func (r *CharacterRepository) CountByWorld(ctx context.Context, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM characters WHERE world_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, worldID).Scan(&count)
	return count, err
}

func (r *CharacterRepository) scanCharacters(rows pgx.Rows) ([]*world.Character, error) {
	characters := make([]*world.Character, 0)
	for rows.Next() {
		var c world.Character
		var archetypeID *uuid.UUID

		err := rows.Scan(
			&c.ID, &c.WorldID, &archetypeID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}

		c.ArchetypeID = archetypeID
		characters = append(characters, &c)
	}

	return characters, rows.Err()
}

