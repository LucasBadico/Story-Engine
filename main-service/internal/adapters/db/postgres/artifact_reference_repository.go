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

var _ repositories.ArtifactReferenceRepository = (*ArtifactReferenceRepository)(nil)

// ArtifactReferenceRepository implements the artifact reference repository interface
type ArtifactReferenceRepository struct {
	db *DB
}

// NewArtifactReferenceRepository creates a new artifact reference repository
func NewArtifactReferenceRepository(db *DB) *ArtifactReferenceRepository {
	return &ArtifactReferenceRepository{db: db}
}

// Create creates a new artifact reference
func (r *ArtifactReferenceRepository) Create(ctx context.Context, ref *world.ArtifactReference) error {
	query := `
		INSERT INTO artifact_references (id, artifact_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		ref.ID, ref.ArtifactID, string(ref.EntityType), ref.EntityID, ref.CreatedAt)
	return err
}

// GetByID retrieves an artifact reference by ID
func (r *ArtifactReferenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.ArtifactReference, error) {
	query := `
		SELECT id, artifact_id, entity_type, entity_id, created_at
		FROM artifact_references
		WHERE id = $1
	`
	var ref world.ArtifactReference
	var entityTypeStr string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&ref.ID, &ref.ArtifactID, &entityTypeStr, &ref.EntityID, &ref.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "artifact_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	ref.EntityType = world.ArtifactReferenceEntityType(entityTypeStr)
	return &ref, nil
}

// ListByArtifact lists artifact references for an artifact
func (r *ArtifactReferenceRepository) ListByArtifact(ctx context.Context, artifactID uuid.UUID) ([]*world.ArtifactReference, error) {
	query := `
		SELECT id, artifact_id, entity_type, entity_id, created_at
		FROM artifact_references
		WHERE artifact_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifactReferences(rows)
}

// ListByEntity lists artifact references for an entity
func (r *ArtifactReferenceRepository) ListByEntity(ctx context.Context, entityType world.ArtifactReferenceEntityType, entityID uuid.UUID) ([]*world.ArtifactReference, error) {
	query := `
		SELECT id, artifact_id, entity_type, entity_id, created_at
		FROM artifact_references
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifactReferences(rows)
}

// Delete deletes an artifact reference
func (r *ArtifactReferenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM artifact_references WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByArtifact deletes all artifact references for an artifact
func (r *ArtifactReferenceRepository) DeleteByArtifact(ctx context.Context, artifactID uuid.UUID) error {
	query := `DELETE FROM artifact_references WHERE artifact_id = $1`
	_, err := r.db.Exec(ctx, query, artifactID)
	return err
}

// DeleteByArtifactAndEntity deletes a specific artifact reference
func (r *ArtifactReferenceRepository) DeleteByArtifactAndEntity(ctx context.Context, artifactID uuid.UUID, entityType world.ArtifactReferenceEntityType, entityID uuid.UUID) error {
	query := `DELETE FROM artifact_references WHERE artifact_id = $1 AND entity_type = $2 AND entity_id = $3`
	_, err := r.db.Exec(ctx, query, artifactID, string(entityType), entityID)
	return err
}

func (r *ArtifactReferenceRepository) scanArtifactReferences(rows pgx.Rows) ([]*world.ArtifactReference, error) {
	references := make([]*world.ArtifactReference, 0)
	for rows.Next() {
		var ref world.ArtifactReference
		var entityTypeStr string

		err := rows.Scan(
			&ref.ID, &ref.ArtifactID, &entityTypeStr, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}

		ref.EntityType = world.ArtifactReferenceEntityType(entityTypeStr)
		references = append(references, &ref)
	}

	return references, rows.Err()
}

