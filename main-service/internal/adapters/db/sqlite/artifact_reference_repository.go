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

var _ repositories.ArtifactReferenceRepository = (*ArtifactReferenceRepository)(nil)

// ArtifactReferenceRepository implements the artifact reference repository interface for SQLite
type ArtifactReferenceRepository struct {
	db *DB
}

// NewArtifactReferenceRepository creates a new artifact reference repository
func NewArtifactReferenceRepository(db *DB) *ArtifactReferenceRepository {
	return &ArtifactReferenceRepository{db: db}
}

// Create creates a new artifact reference
func (r *ArtifactReferenceRepository) Create(ctx context.Context, ref *world.ArtifactReference) error {
	// Get tenant_id from artifact
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM artifacts WHERE id = ?", ref.ArtifactID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO artifact_references (id, tenant_id, artifact_id, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(ctx, query,
		ref.ID.String(),
		tenantID.String(),
		ref.ArtifactID.String(),
		string(ref.EntityType),
		ref.EntityID.String(),
		ref.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an artifact reference by ID
func (r *ArtifactReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.ArtifactReference, error) {
	query := `
		SELECT id, artifact_id, entity_type, entity_id, created_at
		FROM artifact_references
		WHERE tenant_id = ? AND id = ?
	`
	var ref world.ArtifactReference
	var idStr, artifactIDStr, entityIDStr, createdAtStr string
	var entityTypeStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &artifactIDStr, &entityTypeStr, &entityIDStr, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "artifact_reference",
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
	ref.ID = parsedID

	parsedArtifactID, err := uuid.Parse(artifactIDStr)
	if err != nil {
		return nil, err
	}
	ref.ArtifactID = parsedArtifactID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	ref.EntityID = parsedEntityID

	// Parse entity type
	ref.EntityType = world.ArtifactReferenceEntityType(entityTypeStr)

	// Parse timestamp
	ref.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// ListByArtifact lists artifact references for an artifact
func (r *ArtifactReferenceRepository) ListByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) ([]*world.ArtifactReference, error) {
	query := `
		SELECT id, artifact_id, entity_type, entity_id, created_at
		FROM artifact_references
		WHERE tenant_id = ? AND artifact_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), artifactID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifactReferences(rows)
}

// ListByEntity lists artifact references for an entity
func (r *ArtifactReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType world.ArtifactReferenceEntityType, entityID uuid.UUID) ([]*world.ArtifactReference, error) {
	query := `
		SELECT id, artifact_id, entity_type, entity_id, created_at
		FROM artifact_references
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), string(entityType), entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifactReferences(rows)
}

// Delete deletes an artifact reference
func (r *ArtifactReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM artifact_references WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByArtifact deletes all artifact references for an artifact
func (r *ArtifactReferenceRepository) DeleteByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) error {
	query := `DELETE FROM artifact_references WHERE tenant_id = ? AND artifact_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), artifactID.String())
	return err
}

// DeleteByArtifactAndEntity deletes a specific artifact reference
func (r *ArtifactReferenceRepository) DeleteByArtifactAndEntity(ctx context.Context, tenantID, artifactID uuid.UUID, entityType world.ArtifactReferenceEntityType, entityID uuid.UUID) error {
	query := `DELETE FROM artifact_references WHERE tenant_id = ? AND artifact_id = ? AND entity_type = ? AND entity_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), artifactID.String(), string(entityType), entityID.String())
	return err
}

func (r *ArtifactReferenceRepository) scanArtifactReferences(rows *sql.Rows) ([]*world.ArtifactReference, error) {
	references := make([]*world.ArtifactReference, 0)
	for rows.Next() {
		var ref world.ArtifactReference
		var idStr, artifactIDStr, entityIDStr, createdAtStr string
		var entityTypeStr string

		err := rows.Scan(
			&idStr, &artifactIDStr, &entityTypeStr, &entityIDStr, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ref.ID = parsedID

		parsedArtifactID, err := uuid.Parse(artifactIDStr)
		if err != nil {
			return nil, err
		}
		ref.ArtifactID = parsedArtifactID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		ref.EntityID = parsedEntityID

		// Parse entity type
		ref.EntityType = world.ArtifactReferenceEntityType(entityTypeStr)

		// Parse timestamp
		ref.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		references = append(references, &ref)
	}

	return references, rows.Err()
}

