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

var _ repositories.BeatRepository = (*BeatRepository)(nil)

// BeatRepository implements the beat repository interface for SQLite
type BeatRepository struct {
	db *DB
}

// NewBeatRepository creates a new beat repository
func NewBeatRepository(db *DB) *BeatRepository {
	return &BeatRepository{db: db}
}

// Create creates a new beat
func (r *BeatRepository) Create(ctx context.Context, b *story.Beat) error {
	query := `
		INSERT INTO beats (id, tenant_id, scene_id, order_num, type, intent, outcome, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var intent sql.NullString
	if b.Intent != "" {
		intent = sql.NullString{String: b.Intent, Valid: true}
	}

	var outcome sql.NullString
	if b.Outcome != "" {
		outcome = sql.NullString{String: b.Outcome, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		b.ID.String(),
		b.TenantID.String(),
		b.SceneID.String(),
		b.OrderNum,
		string(b.Type),
		intent,
		outcome,
		b.CreatedAt.Format(time.RFC3339),
		b.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a beat by ID
func (r *BeatRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Beat, error) {
	query := `
		SELECT id, tenant_id, scene_id, order_num, type, intent, outcome, created_at, updated_at
		FROM beats
		WHERE tenant_id = ? AND id = ?
	`
	var b story.Beat
	var idStr, tenantIDStr, sceneIDStr, createdAtStr, updatedAtStr string
	var intent, outcome sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &sceneIDStr, &b.OrderNum, &b.Type, &intent, &outcome, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "beat",
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
	b.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	b.TenantID = parsedTenantID

	parsedSceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		return nil, err
	}
	b.SceneID = parsedSceneID

	// Parse timestamps
	b.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	b.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable strings
	if intent.Valid {
		b.Intent = intent.String
	}
	if outcome.Valid {
		b.Outcome = outcome.String
	}

	return &b, nil
}

// ListByScene lists beats for a scene
func (r *BeatRepository) ListByScene(ctx context.Context, tenantID, sceneID uuid.UUID) ([]*story.Beat, error) {
	return r.ListBySceneOrdered(ctx, tenantID, sceneID)
}

// ListBySceneOrdered lists beats for a scene ordered by order_num
func (r *BeatRepository) ListBySceneOrdered(ctx context.Context, tenantID, sceneID uuid.UUID) ([]*story.Beat, error) {
	query := `
		SELECT id, tenant_id, scene_id, order_num, type, intent, outcome, created_at, updated_at
		FROM beats
		WHERE tenant_id = ? AND scene_id = ?
		ORDER BY order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), sceneID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBeats(rows)
}

// Update updates a beat
func (r *BeatRepository) Update(ctx context.Context, b *story.Beat) error {
	query := `
		UPDATE beats
		SET order_num = ?, type = ?, intent = ?, outcome = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var intent sql.NullString
	if b.Intent != "" {
		intent = sql.NullString{String: b.Intent, Valid: true}
	}

	var outcome sql.NullString
	if b.Outcome != "" {
		outcome = sql.NullString{String: b.Outcome, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		b.OrderNum,
		string(b.Type),
		intent,
		outcome,
		b.UpdatedAt.Format(time.RFC3339),
		b.TenantID.String(),
		b.ID.String(),
	)
	return err
}

// Delete deletes a beat
func (r *BeatRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM beats WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByScene deletes all beats for a scene
func (r *BeatRepository) DeleteByScene(ctx context.Context, tenantID, sceneID uuid.UUID) error {
	query := `DELETE FROM beats WHERE tenant_id = ? AND scene_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), sceneID.String())
	return err
}

// ListByStory lists all beats for a story
func (r *BeatRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Beat, error) {
	query := `
		SELECT b.id, b.tenant_id, b.scene_id, b.order_num, b.type, b.intent, b.outcome, b.created_at, b.updated_at
		FROM beats b
		JOIN scenes s ON b.scene_id = s.id
		WHERE b.tenant_id = ? AND s.tenant_id = ? AND s.story_id = ?
		ORDER BY s.order_num ASC, b.order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), tenantID.String(), storyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBeats(rows)
}

func (r *BeatRepository) scanBeats(rows *sql.Rows) ([]*story.Beat, error) {
	var beats []*story.Beat
	for rows.Next() {
		var b story.Beat
		var idStr, tenantIDStr, sceneIDStr, createdAtStr, updatedAtStr string
		var intent, outcome sql.NullString

		err := rows.Scan(&idStr, &tenantIDStr, &sceneIDStr, &b.OrderNum, &b.Type, &intent, &outcome, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		b.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		b.TenantID = parsedTenantID

		parsedSceneID, err := uuid.Parse(sceneIDStr)
		if err != nil {
			return nil, err
		}
		b.SceneID = parsedSceneID

		// Parse timestamps
		b.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		b.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable strings
		if intent.Valid {
			b.Intent = intent.String
		}
		if outcome.Valid {
			b.Outcome = outcome.String
		}

		beats = append(beats, &b)
	}

	return beats, rows.Err()
}

