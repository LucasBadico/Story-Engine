package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

var _ repositories.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository implements the document repository interface
type DocumentRepository struct {
	db *DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new document
func (r *DocumentRepository) Create(ctx context.Context, doc *memory.Document) error {
	metadataJSON, err := doc.MetadataJSON()
	if err != nil {
		return err
	}
	query := `
		INSERT INTO embedding_documents (id, tenant_id, source_type, source_id, title, content, metadata,
			valid_from_chapter, valid_to_chapter, is_canonical, chapter_number, story_id,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err = r.db.Exec(ctx, query,
		doc.ID, doc.TenantID, string(doc.SourceType), doc.SourceID, doc.Title, doc.Content, metadataJSON,
		doc.ValidFromChapter, doc.ValidToChapter, doc.IsCanonical, doc.ChapterNumber, doc.StoryID,
		doc.CreatedAt, doc.UpdatedAt)
	return err
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, metadata, created_at, updated_at
		FROM embedding_documents
		WHERE id = $1
	`
	var doc memory.Document
	var sourceType string
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.TenantID, &sourceType, &doc.SourceID, &doc.Title, &doc.Content, &metadataJSON, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}

	doc.SourceType = memory.SourceType(sourceType)
	doc.Metadata = map[string]string{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, err
		}
	}
	return &doc, nil
}

// GetBySource retrieves a document by source
func (r *DocumentRepository) GetBySource(ctx context.Context, tenantID uuid.UUID, sourceType memory.SourceType, sourceID uuid.UUID) (*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, metadata, created_at, updated_at
		FROM embedding_documents
		WHERE tenant_id = $1 AND source_type = $2 AND source_id = $3
	`
	var doc memory.Document
	var st string
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, string(sourceType), sourceID).Scan(
		&doc.ID, &doc.TenantID, &st, &doc.SourceID, &doc.Title, &doc.Content, &metadataJSON, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}

	doc.SourceType = memory.SourceType(st)
	doc.Metadata = map[string]string{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, err
		}
	}
	return &doc, nil
}

// Update updates an existing document
func (r *DocumentRepository) Update(ctx context.Context, doc *memory.Document) error {
	metadataJSON, err := doc.MetadataJSON()
	if err != nil {
		return err
	}
	query := `
		UPDATE embedding_documents
		SET title = $1, content = $2, metadata = $3,
			valid_from_chapter = $4, valid_to_chapter = $5, is_canonical = $6,
			chapter_number = $7, story_id = $8, updated_at = $9
		WHERE id = $10
	`
	_, err = r.db.Exec(ctx, query, doc.Title, doc.Content, metadataJSON,
		doc.ValidFromChapter, doc.ValidToChapter, doc.IsCanonical,
		doc.ChapterNumber, doc.StoryID, doc.UpdatedAt, doc.ID)
	return err
}

// Delete deletes a document
func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM embedding_documents WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListByTenant lists documents for a tenant
func (r *DocumentRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, metadata, created_at, updated_at
		FROM embedding_documents
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []*memory.Document
	for rows.Next() {
		var doc memory.Document
		var sourceType string
		var metadataJSON []byte
		if err := rows.Scan(&doc.ID, &doc.TenantID, &sourceType, &doc.SourceID, &doc.Title, &doc.Content, &metadataJSON, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		doc.SourceType = memory.SourceType(sourceType)
		doc.Metadata = map[string]string{}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
				return nil, err
			}
		}
		documents = append(documents, &doc)
	}

	return documents, rows.Err()
}

// ListCanonical lists canonical documents for a story, filtered by chapter range.
// Includes world entities (character, location, artifact, world) that have story_id NULL
// but belong to the same tenant.
func (r *DocumentRepository) ListCanonical(ctx context.Context, storyID uuid.UUID, asOfChapter int) ([]*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, metadata,
		       valid_from_chapter, valid_to_chapter, is_canonical, chapter_number, story_id,
		       created_at, updated_at
		FROM embedding_documents
		WHERE is_canonical = true
		  AND valid_from_chapter <= $2
		  AND (valid_to_chapter IS NULL OR valid_to_chapter > $2)
		  AND (story_id = $1 OR (story_id IS NULL AND tenant_id = (
		      SELECT tenant_id FROM embedding_documents WHERE story_id = $1 LIMIT 1
		  )))
		ORDER BY source_type, valid_from_chapter
	`
	rows, err := r.db.Query(ctx, query, storyID, asOfChapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDocumentsWithCanon(rows)
}

// ListByStoryAndSourceType lists documents for a story filtered by source type
func (r *DocumentRepository) ListByStoryAndSourceType(ctx context.Context, storyID uuid.UUID, sourceType memory.SourceType) ([]*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, metadata,
		       valid_from_chapter, valid_to_chapter, is_canonical, chapter_number, story_id,
		       created_at, updated_at
		FROM embedding_documents
		WHERE story_id = $1 AND source_type = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, storyID, string(sourceType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDocumentsWithCanon(rows)
}

func (r *DocumentRepository) scanDocumentsWithCanon(rows interface{ Next() bool; Scan(dest ...any) error; Err() error }) ([]*memory.Document, error) {
	var documents []*memory.Document
	for rows.Next() {
		var doc memory.Document
		var sourceType string
		var metadataJSON []byte
		if err := rows.Scan(
			&doc.ID, &doc.TenantID, &sourceType, &doc.SourceID, &doc.Title, &doc.Content, &metadataJSON,
			&doc.ValidFromChapter, &doc.ValidToChapter, &doc.IsCanonical, &doc.ChapterNumber, &doc.StoryID,
			&doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		doc.SourceType = memory.SourceType(sourceType)
		doc.Metadata = map[string]string{}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
				return nil, err
			}
		}
		documents = append(documents, &doc)
	}
	return documents, rows.Err()
}
