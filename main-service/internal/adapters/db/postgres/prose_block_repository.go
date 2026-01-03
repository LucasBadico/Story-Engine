package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
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
		INSERT INTO prose_blocks (id, tenant_id, chapter_id, order_num, kind, content, word_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.ChapterID, p.OrderNum, string(p.Kind), p.Content, p.WordCount, p.CreatedAt, p.UpdatedAt)
	return err
}

// GetByID retrieves a prose block by ID
func (r *ProseBlockRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ProseBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, kind, content, word_count, created_at, updated_at
		FROM prose_blocks
		WHERE tenant_id = $1 AND id = $2
	`
	var p story.ProseBlock
	var chapterID sql.NullString
	var orderNum sql.NullInt32
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&p.ID, &p.TenantID, &chapterID, &orderNum, &p.Kind, &p.Content, &p.WordCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       id.String(),
			}
		}
		return nil, err
	}
	if chapterID.Valid {
		parsedID, err := uuid.Parse(chapterID.String)
		if err == nil {
			p.ChapterID = &parsedID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int32)
		p.OrderNum = &order
	}
	return &p, nil
}

// ListByChapter lists prose blocks for a chapter
func (r *ProseBlockRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.ProseBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, kind, content, word_count, created_at, updated_at
		FROM prose_blocks
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY COALESCE(order_num, 0) ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProseBlocks(rows)
}

// GetByChapterAndKind retrieves a prose block by chapter and kind
func (r *ProseBlockRepository) GetByChapterAndKind(ctx context.Context, tenantID, chapterID uuid.UUID, kind story.ProseKind) (*story.ProseBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, kind, content, word_count, created_at, updated_at
		FROM prose_blocks
		WHERE tenant_id = $1 AND chapter_id = $2 AND kind = $3
	`
	var p story.ProseBlock
	var chapterIDNull sql.NullString
	var orderNum sql.NullInt32
	err := r.db.QueryRow(ctx, query, tenantID, chapterID, string(kind)).Scan(
		&p.ID, &p.TenantID, &chapterIDNull, &orderNum, &p.Kind, &p.Content, &p.WordCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       chapterID.String() + "/" + string(kind),
			}
		}
		return nil, err
	}
	if chapterIDNull.Valid {
		parsedID, err := uuid.Parse(chapterIDNull.String)
		if err == nil {
			p.ChapterID = &parsedID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int32)
		p.OrderNum = &order
	}
	return &p, nil
}

// Update updates a prose block
func (r *ProseBlockRepository) Update(ctx context.Context, p *story.ProseBlock) error {
	query := `
		UPDATE prose_blocks
		SET order_num = $2, kind = $3, content = $4, word_count = $5, updated_at = $6
		WHERE tenant_id = $7 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, p.ID, p.OrderNum, string(p.Kind), p.Content, p.WordCount, p.UpdatedAt, p.TenantID)
	return err
}

// Delete deletes a prose block
func (r *ProseBlockRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM prose_blocks WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByChapter deletes all prose blocks for a chapter
func (r *ProseBlockRepository) DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error {
	query := `DELETE FROM prose_blocks WHERE tenant_id = $1 AND chapter_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, chapterID)
	return err
}

func (r *ProseBlockRepository) scanProseBlocks(rows pgx.Rows) ([]*story.ProseBlock, error) {
	var proseBlocks []*story.ProseBlock
	for rows.Next() {
		var p story.ProseBlock
		var chapterID sql.NullString
		var orderNum sql.NullInt32
		err := rows.Scan(&p.ID, &p.TenantID, &chapterID, &orderNum, &p.Kind, &p.Content, &p.WordCount, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if chapterID.Valid {
			parsedID, err := uuid.Parse(chapterID.String)
			if err == nil {
				p.ChapterID = &parsedID
			}
		}
		if orderNum.Valid {
			order := int(orderNum.Int32)
			p.OrderNum = &order
		}
		proseBlocks = append(proseBlocks, &p)
	}
	return proseBlocks, rows.Err()
}

