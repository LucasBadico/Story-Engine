package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.BeatRepository = (*BeatRepository)(nil)

// BeatRepository implements the beat repository interface
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		b.ID, b.TenantID, b.SceneID, b.OrderNum, string(b.Type), b.Intent, b.Outcome, b.CreatedAt, b.UpdatedAt)
	return err
}

// GetByID retrieves a beat by ID
func (r *BeatRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Beat, error) {
	query := `
		SELECT id, tenant_id, scene_id, order_num, type, intent, outcome, created_at, updated_at
		FROM beats
		WHERE tenant_id = $1 AND id = $2
	`
	var b story.Beat
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&b.ID, &b.TenantID, &b.SceneID, &b.OrderNum, &b.Type, &b.Intent, &b.Outcome, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("beat not found")
		}
		return nil, err
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
		WHERE tenant_id = $1 AND scene_id = $2
		ORDER BY order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var beats []*story.Beat
	for rows.Next() {
		var b story.Beat
		err := rows.Scan(&b.ID, &b.TenantID, &b.SceneID, &b.OrderNum, &b.Type, &b.Intent, &b.Outcome, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		beats = append(beats, &b)
	}

	return beats, rows.Err()
}

// Update updates a beat
func (r *BeatRepository) Update(ctx context.Context, b *story.Beat) error {
	query := `
		UPDATE beats
		SET order_num = $2, type = $3, intent = $4, outcome = $5, updated_at = $6
		WHERE tenant_id = $7 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, b.ID, b.OrderNum, string(b.Type), b.Intent, b.Outcome, b.UpdatedAt, b.TenantID)
	return err
}

// Delete deletes a beat
func (r *BeatRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM beats WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByScene deletes all beats for a scene
func (r *BeatRepository) DeleteByScene(ctx context.Context, tenantID, sceneID uuid.UUID) error {
	query := `DELETE FROM beats WHERE tenant_id = $1 AND scene_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, sceneID)
	return err
}

// ListByStory lists all beats for a story
func (r *BeatRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Beat, error) {
	query := `
		SELECT b.id, b.tenant_id, b.scene_id, b.order_num, b.type, b.intent, b.outcome, b.created_at, b.updated_at
		FROM beats b
		JOIN scenes s ON b.scene_id = s.id
		WHERE b.tenant_id = $1 AND s.tenant_id = $1 AND s.story_id = $2
		ORDER BY s.order_num ASC, b.order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var beats []*story.Beat
	for rows.Next() {
		var b story.Beat
		err := rows.Scan(&b.ID, &b.TenantID, &b.SceneID, &b.OrderNum, &b.Type, &b.Intent, &b.Outcome, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		beats = append(beats, &b)
	}

	return beats, rows.Err()
}

