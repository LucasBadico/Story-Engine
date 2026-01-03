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

var _ repositories.ImageBlockRepository = (*ImageBlockRepository)(nil)

// ImageBlockRepository implements the image block repository interface
type ImageBlockRepository struct {
	db *DB
}

// NewImageBlockRepository creates a new image block repository
func NewImageBlockRepository(db *DB) *ImageBlockRepository {
	return &ImageBlockRepository{db: db}
}

// Create creates a new image block
func (r *ImageBlockRepository) Create(ctx context.Context, ib *story.ImageBlock) error {
	query := `
		INSERT INTO image_blocks (id, tenant_id, chapter_id, order_num, kind, image_url, alt_text, caption, width, height, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		ib.ID, ib.TenantID, ib.ChapterID, ib.OrderNum, string(ib.Kind), ib.ImageURL, ib.AltText, ib.Caption, ib.Width, ib.Height, ib.CreatedAt, ib.UpdatedAt)
	return err
}

// GetByID retrieves an image block by ID
func (r *ImageBlockRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ImageBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, kind, image_url, alt_text, caption, width, height, created_at, updated_at
		FROM image_blocks
		WHERE tenant_id = $1 AND id = $2
	`
	var ib story.ImageBlock
	var chapterID sql.NullString
	var orderNum sql.NullInt32
	var altText sql.NullString
	var caption sql.NullString
	var width sql.NullInt32
	var height sql.NullInt32

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&ib.ID, &ib.TenantID, &chapterID, &orderNum, &ib.Kind, &ib.ImageURL, &altText, &caption, &width, &height, &ib.CreatedAt, &ib.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "image_block",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if chapterID.Valid {
		parsedID, err := uuid.Parse(chapterID.String)
		if err == nil {
			ib.ChapterID = &parsedID
		}
	}
	if orderNum.Valid {
		order := int(orderNum.Int32)
		ib.OrderNum = &order
	}
	if altText.Valid {
		ib.AltText = &altText.String
	}
	if caption.Valid {
		ib.Caption = &caption.String
	}
	if width.Valid {
		w := int(width.Int32)
		ib.Width = &w
	}
	if height.Valid {
		h := int(height.Int32)
		ib.Height = &h
	}

	return &ib, nil
}

// ListByChapter lists image blocks for a chapter
func (r *ImageBlockRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.ImageBlock, error) {
	query := `
		SELECT id, tenant_id, chapter_id, order_num, kind, image_url, alt_text, caption, width, height, created_at, updated_at
		FROM image_blocks
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY COALESCE(order_num, 0) ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanImageBlocks(rows)
}

// Update updates an image block
func (r *ImageBlockRepository) Update(ctx context.Context, ib *story.ImageBlock) error {
	query := `
		UPDATE image_blocks
		SET chapter_id = $2, order_num = $3, kind = $4, image_url = $5, alt_text = $6, caption = $7, width = $8, height = $9, updated_at = $10
		WHERE tenant_id = $11 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, ib.ID, ib.ChapterID, ib.OrderNum, string(ib.Kind), ib.ImageURL, ib.AltText, ib.Caption, ib.Width, ib.Height, ib.UpdatedAt, ib.TenantID)
	return err
}

// Delete deletes an image block
func (r *ImageBlockRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM image_blocks WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "image_block",
			ID:       id.String(),
		}
	}
	return nil
}

// DeleteByChapter deletes all image blocks for a chapter
func (r *ImageBlockRepository) DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error {
	query := `DELETE FROM image_blocks WHERE tenant_id = $1 AND chapter_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, chapterID)
	return err
}

func (r *ImageBlockRepository) scanImageBlocks(rows pgx.Rows) ([]*story.ImageBlock, error) {
	imageBlocks := make([]*story.ImageBlock, 0)
	for rows.Next() {
		var ib story.ImageBlock
		var chapterID sql.NullString
		var orderNum sql.NullInt32
		var altText sql.NullString
		var caption sql.NullString
		var width sql.NullInt32
		var height sql.NullInt32

		err := rows.Scan(
			&ib.ID, &ib.TenantID, &chapterID, &orderNum, &ib.Kind, &ib.ImageURL, &altText, &caption, &width, &height, &ib.CreatedAt, &ib.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if chapterID.Valid {
			parsedID, err := uuid.Parse(chapterID.String)
			if err == nil {
				ib.ChapterID = &parsedID
			}
		}
		if orderNum.Valid {
			order := int(orderNum.Int32)
			ib.OrderNum = &order
		}
		if altText.Valid {
			ib.AltText = &altText.String
		}
		if caption.Valid {
			ib.Caption = &caption.String
		}
		if width.Valid {
			w := int(width.Int32)
			ib.Width = &w
		}
		if height.Valid {
			h := int(height.Int32)
			ib.Height = &h
		}

		imageBlocks = append(imageBlocks, &ib)
	}
	return imageBlocks, rows.Err()
}


