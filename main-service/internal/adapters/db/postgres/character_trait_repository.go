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

var _ repositories.CharacterTraitRepository = (*CharacterTraitRepository)(nil)

// CharacterTraitRepository implements the character-trait repository interface
type CharacterTraitRepository struct {
	db *DB
}

// NewCharacterTraitRepository creates a new character-trait repository
func NewCharacterTraitRepository(db *DB) *CharacterTraitRepository {
	return &CharacterTraitRepository{db: db}
}

// Create creates a new character-trait relationship
func (r *CharacterTraitRepository) Create(ctx context.Context, ct *world.CharacterTrait) error {
	query := `
		INSERT INTO character_traits (
			id, character_id, trait_id, trait_name, trait_category, trait_description,
			value, notes, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		ct.ID, ct.CharacterID, ct.TraitID, ct.TraitName, ct.TraitCategory, ct.TraitDescription,
		ct.Value, ct.Notes, ct.CreatedAt, ct.UpdatedAt)
	return err
}

// GetByCharacter retrieves all traits for a character
func (r *CharacterTraitRepository) GetByCharacter(ctx context.Context, characterID uuid.UUID) ([]*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE character_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterTraits(rows)
}

// GetByID retrieves a character-trait by ID
func (r *CharacterTraitRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE id = $1
	`
	var ct world.CharacterTrait

	err := r.db.QueryRow(ctx, query, id).Scan(
		&ct.ID, &ct.CharacterID, &ct.TraitID, &ct.TraitName, &ct.TraitCategory, &ct.TraitDescription,
		&ct.Value, &ct.Notes, &ct.CreatedAt, &ct.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_trait",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &ct, nil
}

// GetByCharacterAndTrait retrieves a character-trait by character and trait IDs
func (r *CharacterTraitRepository) GetByCharacterAndTrait(ctx context.Context, characterID, traitID uuid.UUID) (*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE character_id = $1 AND trait_id = $2
	`
	var ct world.CharacterTrait

	err := r.db.QueryRow(ctx, query, characterID, traitID).Scan(
		&ct.ID, &ct.CharacterID, &ct.TraitID, &ct.TraitName, &ct.TraitCategory, &ct.TraitDescription,
		&ct.Value, &ct.Notes, &ct.CreatedAt, &ct.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_trait",
				ID:       characterID.String() + "/" + traitID.String(),
			}
		}
		return nil, err
	}

	return &ct, nil
}

// Update updates a character-trait
func (r *CharacterTraitRepository) Update(ctx context.Context, ct *world.CharacterTrait) error {
	query := `
		UPDATE character_traits
		SET trait_name = $2, trait_category = $3, trait_description = $4,
		    value = $5, notes = $6, updated_at = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		ct.ID, ct.TraitName, ct.TraitCategory, ct.TraitDescription,
		ct.Value, ct.Notes, ct.UpdatedAt)
	return err
}

// Delete deletes a character-trait
func (r *CharacterTraitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM character_traits WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByCharacter deletes all traits for a character
func (r *CharacterTraitRepository) DeleteByCharacter(ctx context.Context, characterID uuid.UUID) error {
	query := `DELETE FROM character_traits WHERE character_id = $1`
	_, err := r.db.Exec(ctx, query, characterID)
	return err
}

func (r *CharacterTraitRepository) scanCharacterTraits(rows pgx.Rows) ([]*world.CharacterTrait, error) {
	characterTraits := make([]*world.CharacterTrait, 0)
	for rows.Next() {
		var ct world.CharacterTrait

		err := rows.Scan(
			&ct.ID, &ct.CharacterID, &ct.TraitID, &ct.TraitName, &ct.TraitCategory, &ct.TraitDescription,
			&ct.Value, &ct.Notes, &ct.CreatedAt, &ct.UpdatedAt)
		if err != nil {
			return nil, err
		}

		characterTraits = append(characterTraits, &ct)
	}

	return characterTraits, rows.Err()
}

