package postgres

import (
	"context"
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
	query := `
		INSERT INTO embedding_documents (id, tenant_id, source_type, source_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.TenantID, string(doc.SourceType), doc.SourceID, doc.Title, doc.Content, doc.CreatedAt, doc.UpdatedAt)
	return err
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, created_at, updated_at
		FROM embedding_documents
		WHERE id = $1
	`
	var doc memory.Document
	var sourceType string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.TenantID, &sourceType, &doc.SourceID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}

	doc.SourceType = memory.SourceType(sourceType)
	return &doc, nil
}

// GetBySource retrieves a document by source
func (r *DocumentRepository) GetBySource(ctx context.Context, tenantID uuid.UUID, sourceType memory.SourceType, sourceID uuid.UUID) (*memory.Document, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, title, content, created_at, updated_at
		FROM embedding_documents
		WHERE tenant_id = $1 AND source_type = $2 AND source_id = $3
	`
	var doc memory.Document
	var st string

	err := r.db.QueryRow(ctx, query, tenantID, string(sourceType), sourceID).Scan(
		&doc.ID, &doc.TenantID, &st, &doc.SourceID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}

	doc.SourceType = memory.SourceType(st)
	return &doc, nil
}

// Update updates an existing document
func (r *DocumentRepository) Update(ctx context.Context, doc *memory.Document) error {
	query := `
		UPDATE embedding_documents
		SET title = $1, content = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, doc.Title, doc.Content, doc.UpdatedAt, doc.ID)
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
		SELECT id, tenant_id, source_type, source_id, title, content, created_at, updated_at
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
		if err := rows.Scan(&doc.ID, &doc.TenantID, &sourceType, &doc.SourceID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		doc.SourceType = memory.SourceType(sourceType)
		documents = append(documents, &doc)
	}

	return documents, rows.Err()
}

