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
// tenantID should be obtained from the character
func (r *CharacterTraitRepository) Create(ctx context.Context, ct *world.CharacterTrait) error {
	// Get tenant_id from character
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM characters WHERE id = $1", ct.CharacterID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO character_traits (
			id, tenant_id, character_id, trait_id, trait_name, trait_category, trait_description,
			value, notes, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		ct.ID, tenantID, ct.CharacterID, ct.TraitID, ct.TraitName, ct.TraitCategory, ct.TraitDescription,
		ct.Value, ct.Notes, ct.CreatedAt, ct.UpdatedAt)
	return err
}

// GetByCharacter retrieves all traits for a character
func (r *CharacterTraitRepository) GetByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE tenant_id = $1 AND character_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterTraits(rows)
}

// GetByID retrieves a character-trait by ID
func (r *CharacterTraitRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE tenant_id = $1 AND id = $2
	`
	var ct world.CharacterTrait

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
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
func (r *CharacterTraitRepository) GetByCharacterAndTrait(ctx context.Context, tenantID, characterID, traitID uuid.UUID) (*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE tenant_id = $1 AND character_id = $2 AND trait_id = $3
	`
	var ct world.CharacterTrait

	err := r.db.QueryRow(ctx, query, tenantID, characterID, traitID).Scan(
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
	// Get tenant_id from character
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM characters WHERE id = $1", ct.CharacterID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		UPDATE character_traits
		SET trait_name = $2, trait_category = $3, trait_description = $4,
		    value = $5, notes = $6, updated_at = $7
		WHERE tenant_id = $8 AND id = $1
	`
	_, err := r.db.Exec(ctx, query,
		ct.ID, ct.TraitName, ct.TraitCategory, ct.TraitDescription,
		ct.Value, ct.Notes, ct.UpdatedAt, tenantID)
	return err
}

// Delete deletes a character-trait
func (r *CharacterTraitRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_traits WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByCharacter deletes all traits for a character
func (r *CharacterTraitRepository) DeleteByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) error {
	query := `DELETE FROM character_traits WHERE tenant_id = $1 AND character_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, characterID)
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


