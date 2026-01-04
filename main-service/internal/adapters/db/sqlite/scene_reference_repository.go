package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.SceneReferenceRepository = (*SceneReferenceRepository)(nil)

// SceneReferenceRepository implements the scene reference repository interface for SQLite
type SceneReferenceRepository struct {
	db *DB
}

// NewSceneReferenceRepository creates a new scene reference repository
func NewSceneReferenceRepository(db *DB) *SceneReferenceRepository {
	return &SceneReferenceRepository{db: db}
}

// Create creates a new scene reference
func (r *SceneReferenceRepository) Create(ctx context.Context, ref *story.SceneReference) error {
	// Get tenant_id from scene
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM scenes WHERE id = ?", ref.SceneID.String()).Scan(&tenantIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       ref.SceneID.String(),
			}
		}
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO scene_references (id, tenant_id, scene_id, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(ctx, query,
		ref.ID.String(),
		tenantID.String(),
		ref.SceneID.String(),
		string(ref.EntityType),
		ref.EntityID.String(),
		ref.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a scene reference by ID
func (r *SceneReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.SceneReference, error) {
	query := `
		SELECT id, scene_id, entity_type, entity_id, created_at
		FROM scene_references
		WHERE tenant_id = ? AND id = ?
	`
	var ref story.SceneReference
	var idStr, sceneIDStr, entityIDStr, createdAtStr string
	var entityTypeStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &sceneIDStr, &entityTypeStr, &entityIDStr, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "scene_reference",
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

	parsedSceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		return nil, err
	}
	ref.SceneID = parsedSceneID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	ref.EntityID = parsedEntityID

	// Parse entity type
	ref.EntityType = story.SceneReferenceEntityType(entityTypeStr)

	// Parse timestamp
	ref.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// ListByScene lists scene references for a scene
func (r *SceneReferenceRepository) ListByScene(ctx context.Context, tenantID, sceneID uuid.UUID) ([]*story.SceneReference, error) {
	query := `
		SELECT id, scene_id, entity_type, entity_id, created_at
		FROM scene_references
		WHERE tenant_id = ? AND scene_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), sceneID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSceneReferences(rows)
}

// ListByEntity lists scene references for an entity
func (r *SceneReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.SceneReferenceEntityType, entityID uuid.UUID) ([]*story.SceneReference, error) {
	query := `
		SELECT id, scene_id, entity_type, entity_id, created_at
		FROM scene_references
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), string(entityType), entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSceneReferences(rows)
}

// Delete deletes a scene reference
func (r *SceneReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM scene_references WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByScene deletes all scene references for a scene
func (r *SceneReferenceRepository) DeleteByScene(ctx context.Context, tenantID, sceneID uuid.UUID) error {
	query := `DELETE FROM scene_references WHERE tenant_id = ? AND scene_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), sceneID.String())
	return err
}

// DeleteBySceneAndEntity deletes a specific scene reference
func (r *SceneReferenceRepository) DeleteBySceneAndEntity(ctx context.Context, tenantID, sceneID uuid.UUID, entityType story.SceneReferenceEntityType, entityID uuid.UUID) error {
	query := `DELETE FROM scene_references WHERE tenant_id = ? AND scene_id = ? AND entity_type = ? AND entity_id = ?`
	result, err := r.db.Exec(ctx, query, tenantID.String(), sceneID.String(), string(entityType), entityID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "scene_reference",
			ID:       sceneID.String() + "/" + string(entityType) + "/" + entityID.String(),
		}
	}
	return nil
}

func (r *SceneReferenceRepository) scanSceneReferences(rows *sql.Rows) ([]*story.SceneReference, error) {
	references := make([]*story.SceneReference, 0)
	for rows.Next() {
		var ref story.SceneReference
		var idStr, sceneIDStr, entityIDStr, createdAtStr string
		var entityTypeStr string

		err := rows.Scan(
			&idStr, &sceneIDStr, &entityTypeStr, &entityIDStr, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ref.ID = parsedID

		parsedSceneID, err := uuid.Parse(sceneIDStr)
		if err != nil {
			return nil, err
		}
		ref.SceneID = parsedSceneID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		ref.EntityID = parsedEntityID

		// Parse entity type
		ref.EntityType = story.SceneReferenceEntityType(entityTypeStr)

		// Parse timestamp
		ref.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		references = append(references, &ref)
	}

	return references, rows.Err()
}

