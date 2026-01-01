package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/ingestion-service/internal/core/memory"
	"github.com/story-engine/ingestion-service/internal/ports/repositories"
)

var _ repositories.ChunkRepository = (*ChunkRepository)(nil)

// ChunkRepository implements the chunk repository interface
type ChunkRepository struct {
	db *DB
}

// NewChunkRepository creates a new chunk repository
func NewChunkRepository(db *DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// Create creates a new chunk
func (r *ChunkRepository) Create(ctx context.Context, chunk *memory.Chunk) error {
	query := `
		INSERT INTO embedding_chunks (id, document_id, chunk_index, content, embedding, token_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content, formatVector(chunk.Embedding), chunk.TokenCount, chunk.CreatedAt)
	return err
}

// CreateBatch creates multiple chunks in a single transaction
func (r *ChunkRepository) CreateBatch(ctx context.Context, chunks []*memory.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO embedding_chunks (id, document_id, chunk_index, content, embedding, token_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, chunk := range chunks {
		_, err := tx.Exec(ctx, query,
			chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content, formatVector(chunk.Embedding), chunk.TokenCount, chunk.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID retrieves a chunk by ID
func (r *ChunkRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Chunk, error) {
	query := `
		SELECT id, document_id, chunk_index, content, embedding, token_count, created_at
		FROM embedding_chunks
		WHERE id = $1
	`
	var chunk memory.Chunk
	var embeddingStr string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &embeddingStr, &chunk.TokenCount, &chunk.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("chunk not found")
		}
		return nil, err
	}

	chunk.Embedding = parseVector(embeddingStr)
	return &chunk, nil
}

// ListByDocument lists chunks for a document
func (r *ChunkRepository) ListByDocument(ctx context.Context, documentID uuid.UUID) ([]*memory.Chunk, error) {
	query := `
		SELECT id, document_id, chunk_index, content, embedding, token_count, created_at
		FROM embedding_chunks
		WHERE document_id = $1
		ORDER BY chunk_index ASC
	`
	rows, err := r.db.Query(ctx, query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []*memory.Chunk
	for rows.Next() {
		var chunk memory.Chunk
		var embeddingStr string
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &embeddingStr, &chunk.TokenCount, &chunk.CreatedAt); err != nil {
			return nil, err
		}
		chunk.Embedding = parseVector(embeddingStr)
		chunks = append(chunks, &chunk)
	}

	return chunks, rows.Err()
}

// DeleteByDocument deletes all chunks for a document
func (r *ChunkRepository) DeleteByDocument(ctx context.Context, documentID uuid.UUID) error {
	query := `DELETE FROM embedding_chunks WHERE document_id = $1`
	_, err := r.db.Exec(ctx, query, documentID)
	return err
}

// SearchSimilar searches for similar chunks using vector similarity
func (r *ChunkRepository) SearchSimilar(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int, sourceTypes []memory.SourceType) ([]*memory.Chunk, error) {
	// Build query with optional source type filter
	query := `
		SELECT c.id, c.document_id, c.chunk_index, c.content, c.embedding, c.token_count, c.created_at
		FROM embedding_chunks c
		INNER JOIN embedding_documents d ON c.document_id = d.id
		WHERE d.tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	// Add source type filter if provided
	if len(sourceTypes) > 0 {
		placeholders := make([]string, len(sourceTypes))
		for i, st := range sourceTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(st))
			argIndex++
		}
		query += fmt.Sprintf(" AND d.source_type IN (%s)", strings.Join(placeholders, ","))
	}

	// Add vector similarity search
	query += fmt.Sprintf(`
		ORDER BY c.embedding <=> $%d::vector
		LIMIT $%d
	`, argIndex, argIndex+1)
	args = append(args, formatVector(embedding), limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []*memory.Chunk
	for rows.Next() {
		var chunk memory.Chunk
		var embeddingStr string
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &embeddingStr, &chunk.TokenCount, &chunk.CreatedAt); err != nil {
			return nil, err
		}
		chunk.Embedding = parseVector(embeddingStr)
		chunks = append(chunks, &chunk)
	}

	return chunks, rows.Err()
}

// formatVector formats a float32 slice as a pgvector string
// pgvector expects format: [0.1,0.2,0.3] (no spaces after commas)
func formatVector(vec []float32) string {
	if len(vec) == 0 {
		return "[]"
	}
	strs := make([]string, len(vec))
	for i, v := range vec {
		strs[i] = fmt.Sprintf("%g", v) // %g removes trailing zeros
	}
	return "[" + strings.Join(strs, ",") + "]"
}

// parseVector parses a pgvector string into a float32 slice
func parseVector(s string) []float32 {
	// Remove brackets and whitespace
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	vec := make([]float32, len(parts))
	for i, part := range parts {
		var v float32
		fmt.Sscanf(strings.TrimSpace(part), "%f", &v)
		vec[i] = v
	}
	return vec
}

