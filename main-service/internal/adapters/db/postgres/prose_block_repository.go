package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ProseBlockRepository = (*ProseBlockRepository)(nil)

// ProseBlockRepository implements the prose block repository interface
type ProseBlockRepository struct {
	db *DB
}

// NewProseBlockRepository creates a new prose block repository
func NewProseBlockRepository(db *DB) *ProseBlockRepository {
	return &ProseBlockRepository{db: db}
}

// Create creates a new prose block
func (r *ProseBlockRepository) Create(ctx context.Context, p *story.ProseBlock) error {
	query := `
		INSERT INTO prose_blocks (id, scene_id, kind, content, word_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.SceneID, string(p.Kind), p.Content, p.WordCount, p.CreatedAt, p.UpdatedAt)
	return err
}

// GetByID retrieves a prose block by ID
func (r *ProseBlockRepository) GetByID(ctx context.Context, id uuid.UUID) (*story.ProseBlock, error) {
	query := `
		SELECT id, scene_id, kind, content, word_count, created_at, updated_at
		FROM prose_blocks
		WHERE id = $1
	`
	var p story.ProseBlock
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.SceneID, &p.Kind, &p.Content, &p.WordCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("prose block not found")
		}
		return nil, err
	}
	return &p, nil
}

// ListByScene lists prose blocks for a scene
func (r *ProseBlockRepository) ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*story.ProseBlock, error) {
	query := `
		SELECT id, scene_id, kind, content, word_count, created_at, updated_at
		FROM prose_blocks
		WHERE scene_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proseBlocks []*story.ProseBlock
	for rows.Next() {
		var p story.ProseBlock
		err := rows.Scan(&p.ID, &p.SceneID, &p.Kind, &p.Content, &p.WordCount, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		proseBlocks = append(proseBlocks, &p)
	}

	return proseBlocks, rows.Err()
}

// GetBySceneAndKind retrieves a prose block by scene and kind
func (r *ProseBlockRepository) GetBySceneAndKind(ctx context.Context, sceneID uuid.UUID, kind story.ProseKind) (*story.ProseBlock, error) {
	query := `
		SELECT id, scene_id, kind, content, word_count, created_at, updated_at
		FROM prose_blocks
		WHERE scene_id = $1 AND kind = $2
	`
	var p story.ProseBlock
	err := r.db.QueryRow(ctx, query, sceneID, string(kind)).Scan(
		&p.ID, &p.SceneID, &p.Kind, &p.Content, &p.WordCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("prose block not found")
		}
		return nil, err
	}
	return &p, nil
}

// Update updates a prose block
func (r *ProseBlockRepository) Update(ctx context.Context, p *story.ProseBlock) error {
	query := `
		UPDATE prose_blocks
		SET kind = $2, content = $3, word_count = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, p.ID, string(p.Kind), p.Content, p.WordCount, p.UpdatedAt)
	return err
}

// Delete deletes a prose block
func (r *ProseBlockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prose_blocks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByScene deletes all prose blocks for a scene
func (r *ProseBlockRepository) DeleteByScene(ctx context.Context, sceneID uuid.UUID) error {
	query := `DELETE FROM prose_blocks WHERE scene_id = $1`
	_, err := r.db.Exec(ctx, query, sceneID)
	return err
}

