package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ContentBlockRepository = (*ContentBlockRepository)(nil)

// ContentBlockRepository implements the content block repository interface
type ContentBlockRepository struct {
	db *DB
}

// NewContentBlockRepository creates a new content block repository
func NewContentBlockRepository(db *DB) *ContentBlockRepository {
	return &ContentBlockRepository{db: db}
}

// Create creates a new content block
func (r *ContentBlockRepository) Create(ctx context.Context, c *story.ContentBlock) error {
	metadataJSON, err := c.MetadataToJSON()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO content_blocks (id, tenant_id, chapter_id, order_num, type, kind, content, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		c.ID, c.TenantID, c.ChapterID, c.OrderNum, string(c.Type), string(c.Kind), c.Content, metadataJSON, c.CreatedAt, c.UpdatedAt)
	return err
}

// GetByID retrieves a content block by ID
func (r *ContentBlockRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, type, kind, content, metadata, created_at, updated_at
		FROM content_blocks
		WHERE tenant_id = $1 AND id = $2
	`
	var c story.ContentBlock
	var chapterID sql.NullString
	var orderNum sql.NullInt32
	var metadataJSON []byte
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&c.ID, &c.TenantID, &chapterID, &orderNum, &c.Type, &c.Kind, &c.Content, &metadataJSON, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       id.String(),
			}
		}
		return nil, err
	}
	if chapterID.Valid {
		parsedID, err := uuid.Parse(chapterID.String)
		if err == nil {
			c.ChapterID = &parsedID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int32)
		c.OrderNum = &order
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &c.Metadata); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// ListByChapter lists content blocks for a chapter
func (r *ContentBlockRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.ContentBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, type, kind, content, metadata, created_at, updated_at
		FROM content_blocks
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY COALESCE(order_num, 0) ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanContentBlocks(rows)
}

// GetByChapterAndKind retrieves a content block by chapter and kind
func (r *ContentBlockRepository) GetByChapterAndKind(ctx context.Context, tenantID, chapterID uuid.UUID, kind story.ContentKind) (*story.ContentBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, type, kind, content, metadata, created_at, updated_at
		FROM content_blocks
		WHERE tenant_id = $1 AND chapter_id = $2 AND kind = $3
	`
	var c story.ContentBlock
	var chapterIDNull sql.NullString
	var orderNum sql.NullInt32
	var metadataJSON []byte
	err := r.db.QueryRow(ctx, query, tenantID, chapterID, string(kind)).Scan(
		&c.ID, &c.TenantID, &chapterIDNull, &orderNum, &c.Type, &c.Kind, &c.Content, &metadataJSON, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       chapterID.String() + "/" + string(kind),
			}
		}
		return nil, err
	}
	if chapterIDNull.Valid {
		parsedID, err := uuid.Parse(chapterIDNull.String)
		if err == nil {
			c.ChapterID = &parsedID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int32)
		c.OrderNum = &order
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &c.Metadata); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Update updates a content block
func (r *ContentBlockRepository) Update(ctx context.Context, c *story.ContentBlock) error {
	metadataJSON, err := c.MetadataToJSON()
	if err != nil {
		return err
	}

	query := `
		UPDATE content_blocks
		SET order_num = $2, type = $3, kind = $4, content = $5, metadata = $6, updated_at = $7
		WHERE tenant_id = $8 AND id = $1
	`
	_, err = r.db.Exec(ctx, query, c.ID, c.OrderNum, string(c.Type), string(c.Kind), c.Content, metadataJSON, c.UpdatedAt, c.TenantID)
	return err
}

// Delete deletes a content block
func (r *ContentBlockRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM content_blocks WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByChapter deletes all content blocks for a chapter
func (r *ContentBlockRepository) DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error {
	query := `DELETE FROM content_blocks WHERE tenant_id = $1 AND chapter_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, chapterID)
	return err
}

func (r *ContentBlockRepository) scanContentBlocks(rows pgx.Rows) ([]*story.ContentBlock, error) {
	var contentBlocks []*story.ContentBlock
	for rows.Next() {
		var c story.ContentBlock
		var chapterID sql.NullString
		var orderNum sql.NullInt32
		var metadataJSON []byte
		err := rows.Scan(&c.ID, &c.TenantID, &chapterID, &orderNum, &c.Type, &c.Kind, &c.Content, &metadataJSON, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if chapterID.Valid {
			parsedID, err := uuid.Parse(chapterID.String)
			if err == nil {
				c.ChapterID = &parsedID
			}
		}
		if orderNum.Valid {
			order := int(orderNum.Int32)
			c.OrderNum = &order
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &c.Metadata); err != nil {
				return nil, err
			}
		}
		contentBlocks = append(contentBlocks, &c)
	}
	return contentBlocks, rows.Err()
}
