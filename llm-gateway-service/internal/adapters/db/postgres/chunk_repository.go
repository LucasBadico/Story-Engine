package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
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
		INSERT INTO embedding_chunks (
			id, document_id, chunk_index, content, embedding, token_count, created_at,
			scene_id, beat_id, beat_type, beat_intent, characters, location_id, location_name,
			timeline, pov_character, content_kind, type, embed_text
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`
	charactersJSON, err := json.Marshal(chunk.Characters)
	if err != nil {
		return fmt.Errorf("failed to marshal characters: %w", err)
	}
	if len(chunk.Characters) == 0 {
		charactersJSON = []byte("[]")
	}

	_, err = r.db.Exec(ctx, query,
		chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content, formatVector(chunk.Embedding), chunk.TokenCount, chunk.CreatedAt,
		chunk.SceneID, chunk.BeatID, chunk.BeatType, chunk.BeatIntent, string(charactersJSON), chunk.LocationID, chunk.LocationName,
		chunk.Timeline, chunk.POVCharacter, chunk.ContentKind, chunk.ChunkType, chunk.EmbedText)
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
		INSERT INTO embedding_chunks (
			id, document_id, chunk_index, content, embedding, token_count, created_at,
			scene_id, beat_id, beat_type, beat_intent, characters, location_id, location_name,
			timeline, pov_character, content_kind, type, embed_text
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	for _, chunk := range chunks {
		charactersJSON, err := json.Marshal(chunk.Characters)
		if err != nil {
			return fmt.Errorf("failed to marshal characters: %w", err)
		}
		if len(chunk.Characters) == 0 {
			charactersJSON = []byte("[]")
		}

		_, err = tx.Exec(ctx, query,
			chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content, formatVector(chunk.Embedding), chunk.TokenCount, chunk.CreatedAt,
			chunk.SceneID, chunk.BeatID, chunk.BeatType, chunk.BeatIntent, string(charactersJSON), chunk.LocationID, chunk.LocationName,
			chunk.Timeline, chunk.POVCharacter, chunk.ContentKind, chunk.ChunkType, chunk.EmbedText)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID retrieves a chunk by ID
func (r *ChunkRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Chunk, error) {
	query := `
		SELECT id, document_id, chunk_index, content, embedding, token_count, created_at,
		       scene_id, beat_id, beat_type, beat_intent, characters, location_id, location_name,
		       timeline, pov_character, content_kind, type, embed_text
		FROM embedding_chunks
		WHERE id = $1
	`
	var chunk memory.Chunk
	var embeddingStr string
	var charactersJSON string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &embeddingStr, &chunk.TokenCount, &chunk.CreatedAt,
		&chunk.SceneID, &chunk.BeatID, &chunk.BeatType, &chunk.BeatIntent, &charactersJSON, &chunk.LocationID, &chunk.LocationName,
		&chunk.Timeline, &chunk.POVCharacter, &chunk.ContentKind, &chunk.ChunkType, &chunk.EmbedText)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("chunk not found")
		}
		return nil, err
	}

	chunk.Embedding = parseVector(embeddingStr)
	if charactersJSON != "" {
		if err := json.Unmarshal([]byte(charactersJSON), &chunk.Characters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal characters: %w", err)
		}
	}
	return &chunk, nil
}

// ListByDocument lists chunks for a document
func (r *ChunkRepository) ListByDocument(ctx context.Context, documentID uuid.UUID) ([]*memory.Chunk, error) {
	query := `
		SELECT id, document_id, chunk_index, content, embedding, token_count, created_at,
		       scene_id, beat_id, beat_type, beat_intent, characters, location_id, location_name,
		       timeline, pov_character, content_kind, type, embed_text
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
		var charactersJSON string
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &embeddingStr, &chunk.TokenCount, &chunk.CreatedAt,
			&chunk.SceneID, &chunk.BeatID, &chunk.BeatType, &chunk.BeatIntent, &charactersJSON, &chunk.LocationID, &chunk.LocationName,
			&chunk.Timeline, &chunk.POVCharacter, &chunk.ContentKind, &chunk.ChunkType, &chunk.EmbedText); err != nil {
			return nil, err
		}
		chunk.Embedding = parseVector(embeddingStr)
		if charactersJSON != "" {
			if err := json.Unmarshal([]byte(charactersJSON), &chunk.Characters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal characters: %w", err)
			}
		}
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
func (r *ChunkRepository) SearchSimilar(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int, cursor *repositories.SearchCursor, filters *repositories.SearchFilters) ([]*repositories.ScoredChunk, error) {
	if filters == nil {
		filters = &repositories.SearchFilters{}
	}

	// Build query with filters
	query := `
		SELECT c.id, c.document_id, c.chunk_index, c.content, c.embedding, c.token_count, c.created_at,
		       c.scene_id, c.beat_id, c.beat_type, c.beat_intent, c.characters, c.location_id, c.location_name,
		       c.timeline, c.pov_character, c.content_kind, c.type, c.embed_text,
		       (c.embedding <=> $%d::vector) AS distance
		FROM embedding_chunks c
		INNER JOIN embedding_documents d ON c.document_id = d.id
		WHERE d.tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	// Add source type filter if provided
	if len(filters.SourceTypes) > 0 {
		placeholders := make([]string, len(filters.SourceTypes))
		for i, st := range filters.SourceTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(st))
			argIndex++
		}
		query += fmt.Sprintf(" AND d.source_type IN (%s)", strings.Join(placeholders, ","))
	}

	// Add chunk type filter if provided
	if len(filters.ChunkTypes) > 0 {
		placeholders := make([]string, len(filters.ChunkTypes))
		for i, ct := range filters.ChunkTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, ct)
			argIndex++
		}
		query += fmt.Sprintf(" AND c.type IN (%s)", strings.Join(placeholders, ","))
	}

	// Add story filter if provided
	if filters.StoryID != nil {
		query += fmt.Sprintf(" AND d.source_id = $%d", argIndex)
		args = append(args, *filters.StoryID)
		argIndex++
	}

	// Add beat type filter if provided
	if len(filters.BeatTypes) > 0 {
		placeholders := make([]string, len(filters.BeatTypes))
		for i, bt := range filters.BeatTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, bt)
			argIndex++
		}
		query += fmt.Sprintf(" AND c.beat_type IN (%s)", strings.Join(placeholders, ","))
	}

	// Add scene filter if provided
	if len(filters.SceneIDs) > 0 {
		placeholders := make([]string, len(filters.SceneIDs))
		for i, sid := range filters.SceneIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, sid)
			argIndex++
		}
		query += fmt.Sprintf(" AND c.scene_id IN (%s)", strings.Join(placeholders, ","))
	}

	// Add location filter if provided
	if len(filters.LocationIDs) > 0 {
		placeholders := make([]string, len(filters.LocationIDs))
		for i, lid := range filters.LocationIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, lid)
			argIndex++
		}
		query += fmt.Sprintf(" AND c.location_id IN (%s)", strings.Join(placeholders, ","))
	}

	// Add characters filter if provided (using JSONB containment)
	if len(filters.Characters) > 0 {
		// Build JSONB array for containment check
		charJSON, err := json.Marshal(filters.Characters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal characters filter: %w", err)
		}
		query += fmt.Sprintf(" AND c.characters @> $%d::jsonb", argIndex)
		args = append(args, string(charJSON))
		argIndex++
	}

	vectorArgIndex := argIndex
	args = append(args, formatVector(embedding))
	argIndex++
	query = fmt.Sprintf(query, vectorArgIndex)

	if cursor != nil {
		query += fmt.Sprintf(`
			AND (
				(c.embedding <=> $%d::vector) > $%d
				OR (
					(c.embedding <=> $%d::vector) = $%d
					AND c.id > $%d
				)
			)
		`, vectorArgIndex, argIndex, vectorArgIndex, argIndex, argIndex+1)
		args = append(args, cursor.Distance, cursor.ChunkID)
		argIndex += 2
	}

	// Add vector similarity search
	query += fmt.Sprintf(`
		ORDER BY c.embedding <=> $%d::vector, c.id
		LIMIT $%d
	`, vectorArgIndex, argIndex)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []*repositories.ScoredChunk
	for rows.Next() {
		var chunk memory.Chunk
		var embeddingStr string
		var charactersJSON string
		var distance float64
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &embeddingStr, &chunk.TokenCount, &chunk.CreatedAt,
			&chunk.SceneID, &chunk.BeatID, &chunk.BeatType, &chunk.BeatIntent, &charactersJSON, &chunk.LocationID, &chunk.LocationName,
			&chunk.Timeline, &chunk.POVCharacter, &chunk.ContentKind, &chunk.ChunkType, &chunk.EmbedText, &distance); err != nil {
			return nil, err
		}
		chunk.Embedding = parseVector(embeddingStr)
		if charactersJSON != "" {
			if err := json.Unmarshal([]byte(charactersJSON), &chunk.Characters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal characters: %w", err)
			}
		}
		chunks = append(chunks, &repositories.ScoredChunk{
			Chunk:    &chunk,
			Distance: distance,
		})
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
