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

var _ repositories.CharacterRelationshipRepository = (*CharacterRelationshipRepository)(nil)

// CharacterRelationshipRepository implements the character-relationship repository interface
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		cr.ID, cr.TenantID, cr.Character1ID, cr.Character2ID, cr.RelationshipType,
		cr.Description, cr.Bidirectional, cr.CreatedAt, cr.UpdatedAt)
	return err
}

// GetByID retrieves a character-relationship by ID
func (r *CharacterRelationshipRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.CharacterRelationship, error) {
	query := `
		SELECT id, tenant_id, character1_id, character2_id, relationship_type,
		       description, bidirectional, created_at, updated_at
		FROM character_relationships
		WHERE tenant_id = $1 AND id = $2
	`
	var cr world.CharacterRelationship

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&cr.ID, &cr.TenantID, &cr.Character1ID, &cr.Character2ID, &cr.RelationshipType,
		&cr.Description, &cr.Bidirectional, &cr.CreatedAt, &cr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_relationship",
				ID:       id.String(),
			}
		}
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
		WHERE tenant_id = $1 AND (character1_id = $2 OR character2_id = $2)
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
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
		SET relationship_type = $2, description = $3, bidirectional = $4, updated_at = $5
		WHERE tenant_id = $6 AND id = $1
	`
	_, err := r.db.Exec(ctx, query,
		cr.ID, cr.RelationshipType, cr.Description, cr.Bidirectional, cr.UpdatedAt, cr.TenantID)
	return err
}

// Delete deletes a character-relationship
func (r *CharacterRelationshipRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_relationships WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *CharacterRelationshipRepository) scanCharacterRelationships(rows pgx.Rows) ([]*world.CharacterRelationship, error) {
	characterRelationships := make([]*world.CharacterRelationship, 0)
	for rows.Next() {
		var cr world.CharacterRelationship

		err := rows.Scan(
			&cr.ID, &cr.TenantID, &cr.Character1ID, &cr.Character2ID, &cr.RelationshipType,
			&cr.Description, &cr.Bidirectional, &cr.CreatedAt, &cr.UpdatedAt)
		if err != nil {
			return nil, err
		}

		characterRelationships = append(characterRelationships, &cr)
	}

	return characterRelationships, rows.Err()
}

