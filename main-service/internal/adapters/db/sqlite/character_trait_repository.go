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

var _ repositories.CharacterTraitRepository = (*CharacterTraitRepository)(nil)

// CharacterTraitRepository implements the character-trait repository interface for SQLite
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
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM characters WHERE id = ?", ct.CharacterID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO character_traits (
			id, tenant_id, character_id, trait_id, trait_name, trait_category, trait_description,
			value, notes, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(ctx, query,
		ct.ID.String(),
		tenantID.String(),
		ct.CharacterID.String(),
		ct.TraitID.String(),
		ct.TraitName,
		ct.TraitCategory,
		ct.TraitDescription,
		ct.Value,
		ct.Notes,
		ct.CreatedAt.Format(time.RFC3339),
		ct.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByCharacter retrieves all traits for a character
func (r *CharacterTraitRepository) GetByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.CharacterTrait, error) {
	query := `
		SELECT id, character_id, trait_id, trait_name, trait_category, trait_description,
		       value, notes, created_at, updated_at
		FROM character_traits
		WHERE tenant_id = ? AND character_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), characterID.String())
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
		WHERE tenant_id = ? AND id = ?
	`
	var ct world.CharacterTrait
	var idStr, characterIDStr, traitIDStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &characterIDStr, &traitIDStr, &ct.TraitName, &ct.TraitCategory, &ct.TraitDescription,
		&ct.Value, &ct.Notes, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_trait",
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
	ct.ID = parsedID

	parsedCharacterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		return nil, err
	}
	ct.CharacterID = parsedCharacterID

	parsedTraitID, err := uuid.Parse(traitIDStr)
	if err != nil {
		return nil, err
	}
	ct.TraitID = parsedTraitID

	// Parse timestamps
	ct.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	ct.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
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
		WHERE tenant_id = ? AND character_id = ? AND trait_id = ?
	`
	var ct world.CharacterTrait
	var idStr, characterIDStr, traitIDStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), characterID.String(), traitID.String()).Scan(
		&idStr, &characterIDStr, &traitIDStr, &ct.TraitName, &ct.TraitCategory, &ct.TraitDescription,
		&ct.Value, &ct.Notes, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_trait",
				ID:       characterID.String() + "/" + traitID.String(),
			}
		}
		return nil, err
	}

	// Parse UUIDs
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	ct.ID = parsedID

	parsedCharacterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		return nil, err
	}
	ct.CharacterID = parsedCharacterID

	parsedTraitID, err := uuid.Parse(traitIDStr)
	if err != nil {
		return nil, err
	}
	ct.TraitID = parsedTraitID

	// Parse timestamps
	ct.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	ct.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	return &ct, nil
}

// Update updates a character-trait
func (r *CharacterTraitRepository) Update(ctx context.Context, ct *world.CharacterTrait) error {
	// Get tenant_id from character
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM characters WHERE id = ?", ct.CharacterID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		UPDATE character_traits
		SET trait_name = ?, trait_category = ?, trait_description = ?,
		    value = ?, notes = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	_, err = r.db.Exec(ctx, query,
		ct.TraitName,
		ct.TraitCategory,
		ct.TraitDescription,
		ct.Value,
		ct.Notes,
		ct.UpdatedAt.Format(time.RFC3339),
		tenantID.String(),
		ct.ID.String(),
	)
	return err
}

// Delete deletes a character-trait
func (r *CharacterTraitRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_traits WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByCharacter deletes all traits for a character
func (r *CharacterTraitRepository) DeleteByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) error {
	query := `DELETE FROM character_traits WHERE tenant_id = ? AND character_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), characterID.String())
	return err
}

func (r *CharacterTraitRepository) scanCharacterTraits(rows *sql.Rows) ([]*world.CharacterTrait, error) {
	characterTraits := make([]*world.CharacterTrait, 0)
	for rows.Next() {
		var ct world.CharacterTrait
		var idStr, characterIDStr, traitIDStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&idStr, &characterIDStr, &traitIDStr, &ct.TraitName, &ct.TraitCategory, &ct.TraitDescription,
			&ct.Value, &ct.Notes, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ct.ID = parsedID

		parsedCharacterID, err := uuid.Parse(characterIDStr)
		if err != nil {
			return nil, err
		}
		ct.CharacterID = parsedCharacterID

		parsedTraitID, err := uuid.Parse(traitIDStr)
		if err != nil {
			return nil, err
		}
		ct.TraitID = parsedTraitID

		// Parse timestamps
		ct.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		ct.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		characterTraits = append(characterTraits, &ct)
	}

	return characterTraits, rows.Err()
}

