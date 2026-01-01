package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.SceneReferenceRepository = (*SceneReferenceRepository)(nil)

// SceneReferenceRepository implements the scene reference repository interface
type SceneReferenceRepository struct {
	db *DB
}

// NewSceneReferenceRepository creates a new scene reference repository
func NewSceneReferenceRepository(db *DB) *SceneReferenceRepository {
	return &SceneReferenceRepository{db: db}
}

// Create creates a new scene reference
func (r *SceneReferenceRepository) Create(ctx context.Context, ref *story.SceneReference) error {
	query := `
		INSERT INTO scene_references (id, scene_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		ref.ID, ref.SceneID, string(ref.EntityType), ref.EntityID, ref.CreatedAt)
	return err
}

// GetByID retrieves a scene reference by ID
func (r *SceneReferenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*story.SceneReference, error) {
	query := `
		SELECT id, scene_id, entity_type, entity_id, created_at
		FROM scene_references
		WHERE id = $1
	`
	var ref story.SceneReference
	var entityTypeStr string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&ref.ID, &ref.SceneID, &entityTypeStr, &ref.EntityID, &ref.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "scene_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	ref.EntityType = story.SceneReferenceEntityType(entityTypeStr)
	return &ref, nil
}

// ListByScene lists scene references for a scene
func (r *SceneReferenceRepository) ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*story.SceneReference, error) {
	query := `
		SELECT id, scene_id, entity_type, entity_id, created_at
		FROM scene_references
		WHERE scene_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSceneReferences(rows)
}

// ListByEntity lists scene references for an entity
func (r *SceneReferenceRepository) ListByEntity(ctx context.Context, entityType story.SceneReferenceEntityType, entityID uuid.UUID) ([]*story.SceneReference, error) {
	query := `
		SELECT id, scene_id, entity_type, entity_id, created_at
		FROM scene_references
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSceneReferences(rows)
}

// Delete deletes a scene reference
func (r *SceneReferenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM scene_references WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByScene deletes all scene references for a scene
func (r *SceneReferenceRepository) DeleteByScene(ctx context.Context, sceneID uuid.UUID) error {
	query := `DELETE FROM scene_references WHERE scene_id = $1`
	_, err := r.db.Exec(ctx, query, sceneID)
	return err
}

// DeleteBySceneAndEntity deletes a specific scene reference
func (r *SceneReferenceRepository) DeleteBySceneAndEntity(ctx context.Context, sceneID uuid.UUID, entityType story.SceneReferenceEntityType, entityID uuid.UUID) error {
	query := `DELETE FROM scene_references WHERE scene_id = $1 AND entity_type = $2 AND entity_id = $3`
	_, err := r.db.Exec(ctx, query, sceneID, string(entityType), entityID)
	return err
}

func (r *SceneReferenceRepository) scanSceneReferences(rows pgx.Rows) ([]*story.SceneReference, error) {
	references := make([]*story.SceneReference, 0)
	for rows.Next() {
		var ref story.SceneReference
		var entityTypeStr string

		err := rows.Scan(
			&ref.ID, &ref.SceneID, &entityTypeStr, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}

		ref.EntityType = story.SceneReferenceEntityType(entityTypeStr)
		references = append(references, &ref)
	}

	return references, rows.Err()
}

