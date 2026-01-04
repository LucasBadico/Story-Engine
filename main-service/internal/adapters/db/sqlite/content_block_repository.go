package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ContentBlockRepository = (*ContentBlockRepository)(nil)

// ContentBlockRepository implements the content block repository interface for SQLite
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var chapterID sql.NullString
	if c.ChapterID != nil {
		chapterID = sql.NullString{String: c.ChapterID.String(), Valid: true}
	}

	var orderNum sql.NullInt64
	if c.OrderNum != nil {
		orderNum = sql.NullInt64{Int64: int64(*c.OrderNum), Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		c.ID.String(),
		c.TenantID.String(),
		chapterID,
		orderNum,
		string(c.Type),
		string(c.Kind),
		c.Content,
		string(metadataJSON),
		c.CreatedAt.Format(time.RFC3339),
		c.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a content block by ID
func (r *ContentBlockRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, type, kind, content, metadata, created_at, updated_at
		FROM content_blocks
		WHERE tenant_id = ? AND id = ?
	`
	var c story.ContentBlock
	var idStr, tenantIDStr, createdAtStr, updatedAtStr, metadataStr string
	var chapterID sql.NullString
	var orderNum sql.NullInt64

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &chapterID, &orderNum, &c.Type, &c.Kind, &c.Content, &metadataStr, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_block",
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
	c.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	c.TenantID = parsedTenantID

	// Parse timestamps
	c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable fields
	if chapterID.Valid {
		parsedChapterID, err := uuid.Parse(chapterID.String)
		if err == nil {
			c.ChapterID = &parsedChapterID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int64)
		c.OrderNum = &order
	}
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &c.Metadata); err != nil {
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
		WHERE tenant_id = ? AND chapter_id = ?
		ORDER BY COALESCE(order_num, 0) ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), chapterID.String())
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
		WHERE tenant_id = ? AND chapter_id = ? AND kind = ?
	`
	var c story.ContentBlock
	var idStr, tenantIDStr, createdAtStr, updatedAtStr, metadataStr string
	var chapterIDNull sql.NullString
	var orderNum sql.NullInt64

	err := r.db.QueryRow(ctx, query, tenantID.String(), chapterID.String(), string(kind)).Scan(
		&idStr, &tenantIDStr, &chapterIDNull, &orderNum, &c.Type, &c.Kind, &c.Content, &metadataStr, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       chapterID.String() + "/" + string(kind),
			}
		}
		return nil, err
	}

	// Parse UUIDs
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	c.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	c.TenantID = parsedTenantID

	// Parse timestamps
	c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable fields
	if chapterIDNull.Valid {
		parsedID, err := uuid.Parse(chapterIDNull.String)
		if err == nil {
			c.ChapterID = &parsedID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int64)
		c.OrderNum = &order
	}
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &c.Metadata); err != nil {
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
		SET order_num = ?, type = ?, kind = ?, content = ?, metadata = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var orderNum sql.NullInt64
	if c.OrderNum != nil {
		orderNum = sql.NullInt64{Int64: int64(*c.OrderNum), Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		orderNum,
		string(c.Type),
		string(c.Kind),
		c.Content,
		string(metadataJSON),
		c.UpdatedAt.Format(time.RFC3339),
		c.TenantID.String(),
		c.ID.String(),
	)
	return err
}

// Delete deletes a content block
func (r *ContentBlockRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM content_blocks WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByChapter deletes all content blocks for a chapter
func (r *ContentBlockRepository) DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error {
	query := `DELETE FROM content_blocks WHERE tenant_id = ? AND chapter_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), chapterID.String())
	return err
}

func (r *ContentBlockRepository) scanContentBlocks(rows *sql.Rows) ([]*story.ContentBlock, error) {
	var contentBlocks []*story.ContentBlock
	for rows.Next() {
		var c story.ContentBlock
		var idStr, tenantIDStr, createdAtStr, updatedAtStr, metadataStr string
		var chapterID sql.NullString
		var orderNum sql.NullInt64

		err := rows.Scan(&idStr, &tenantIDStr, &chapterID, &orderNum, &c.Type, &c.Kind, &c.Content, &metadataStr, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		c.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		c.TenantID = parsedTenantID

		// Parse timestamps
		c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable fields
		if chapterID.Valid {
			parsedID, err := uuid.Parse(chapterID.String)
			if err == nil {
				c.ChapterID = &parsedID
			}
		}
		if orderNum.Valid {
			order := int(orderNum.Int64)
			c.OrderNum = &order
		}
		if metadataStr != "" {
			if err := json.Unmarshal([]byte(metadataStr), &c.Metadata); err != nil {
				return nil, err
			}
		}

		contentBlocks = append(contentBlocks, &c)
	}
	return contentBlocks, rows.Err()
}

