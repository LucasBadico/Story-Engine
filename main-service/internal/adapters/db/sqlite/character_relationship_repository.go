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

var _ repositories.CharacterRelationshipRepository = (*CharacterRelationshipRepository)(nil)

// CharacterRelationshipRepository implements the character-relationship repository interface for SQLite
type CharacterRelationshipRepository struct {
	db *DB
}

// NewCharacterRelationshipRepository creates a new character-relationship repository
func NewCharacterRelationshipRepository(db *DB) *CharacterRelationshipRepository {
	return &CharacterRelationshipRepository{db: db}
}

// Create creates a new character-relationship
func (r *CharacterRelationshipRepository) Create(ctx context.Context, cr *world.CharacterRelationship) error {
	query := `
		INSERT INTO character_relationships (
			id, tenant_id, character1_id, character2_id, relationship_type,
			description, bidirectional, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	bidirectionalInt := 0
	if cr.Bidirectional {
		bidirectionalInt = 1
	}
	_, err := r.db.Exec(ctx, query,
		cr.ID.String(),
		cr.TenantID.String(),
		cr.Character1ID.String(),
		cr.Character2ID.String(),
		cr.RelationshipType,
		cr.Description,
		bidirectionalInt,
		cr.CreatedAt.Format(time.RFC3339),
		cr.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a character-relationship by ID
func (r *CharacterRelationshipRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.CharacterRelationship, error) {
	query := `
		SELECT id, tenant_id, character1_id, character2_id, relationship_type,
		       description, bidirectional, created_at, updated_at
		FROM character_relationships
		WHERE tenant_id = ? AND id = ?
	`
	var cr world.CharacterRelationship
	var idStr, tenantIDStr, character1IDStr, character2IDStr, createdAtStr, updatedAtStr string
	var bidirectionalInt int

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &character1IDStr, &character2IDStr, &cr.RelationshipType,
		&cr.Description, &bidirectionalInt, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_relationship",
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
	cr.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	cr.TenantID = parsedTenantID

	parsedCharacter1ID, err := uuid.Parse(character1IDStr)
	if err != nil {
		return nil, err
	}
	cr.Character1ID = parsedCharacter1ID

	parsedCharacter2ID, err := uuid.Parse(character2IDStr)
	if err != nil {
		return nil, err
	}
	cr.Character2ID = parsedCharacter2ID

	cr.Bidirectional = bidirectionalInt != 0

	// Parse timestamps
	cr.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	cr.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	return &cr, nil
}

// ListByCharacter retrieves all relationships for a character (where character is character1_id or character2_id)
func (r *CharacterRelationshipRepository) ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.CharacterRelationship, error) {
	query := `
		SELECT id, tenant_id, character1_id, character2_id, relationship_type,
		       description, bidirectional, created_at, updated_at
		FROM character_relationships
		WHERE tenant_id = ? AND (character1_id = ? OR character2_id = ?)
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), characterID.String(), characterID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterRelationships(rows)
}

// Update updates a character-relationship
func (r *CharacterRelationshipRepository) Update(ctx context.Context, cr *world.CharacterRelationship) error {
	query := `
		UPDATE character_relationships
		SET relationship_type = ?, description = ?, bidirectional = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	bidirectionalInt := 0
	if cr.Bidirectional {
		bidirectionalInt = 1
	}
	_, err := r.db.Exec(ctx, query,
		cr.RelationshipType,
		cr.Description,
		bidirectionalInt,
		cr.UpdatedAt.Format(time.RFC3339),
		cr.TenantID.String(),
		cr.ID.String(),
	)
	return err
}

// Delete deletes a character-relationship
func (r *CharacterRelationshipRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_relationships WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

func (r *CharacterRelationshipRepository) scanCharacterRelationships(rows *sql.Rows) ([]*world.CharacterRelationship, error) {
	characterRelationships := make([]*world.CharacterRelationship, 0)
	for rows.Next() {
		var cr world.CharacterRelationship
		var idStr, tenantIDStr, character1IDStr, character2IDStr, createdAtStr, updatedAtStr string
		var bidirectionalInt int

		err := rows.Scan(
			&idStr, &tenantIDStr, &character1IDStr, &character2IDStr, &cr.RelationshipType,
			&cr.Description, &bidirectionalInt, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		cr.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		cr.TenantID = parsedTenantID

		parsedCharacter1ID, err := uuid.Parse(character1IDStr)
		if err != nil {
			return nil, err
		}
		cr.Character1ID = parsedCharacter1ID

		parsedCharacter2ID, err := uuid.Parse(character2IDStr)
		if err != nil {
			return nil, err
		}
		cr.Character2ID = parsedCharacter2ID

		cr.Bidirectional = bidirectionalInt != 0

		// Parse timestamps
		cr.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		cr.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		characterRelationships = append(characterRelationships, &cr)
	}

	return characterRelationships, rows.Err()
}

