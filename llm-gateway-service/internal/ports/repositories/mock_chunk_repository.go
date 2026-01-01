package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// MockChunkRepository is a mock implementation of ChunkRepository for testing
type MockChunkRepository struct {
	chunks map[uuid.UUID]*memory.Chunk
	byDocument map[uuid.UUID][]*memory.Chunk
	
	createErr      error
	createBatchErr error
	getByIDErr     error
	listErr        error
	deleteErr      error
	searchErr      error
}

// NewMockChunkRepository creates a new mock chunk repository
func NewMockChunkRepository() *MockChunkRepository {
	return &MockChunkRepository{
		chunks:      make(map[uuid.UUID]*memory.Chunk),
		byDocument:  make(map[uuid.UUID][]*memory.Chunk),
	}
}

// SetCreateError sets an error to return on Create
func (m *MockChunkRepository) SetCreateError(err error) {
	m.createErr = err
}

// SetCreateBatchError sets an error to return on CreateBatch
func (m *MockChunkRepository) SetCreateBatchError(err error) {
	m.createBatchErr = err
}

// SetGetByIDError sets an error to return on GetByID
func (m *MockChunkRepository) SetGetByIDError(err error) {
	m.getByIDErr = err
}

// SetListError sets an error to return on ListByDocument
func (m *MockChunkRepository) SetListError(err error) {
	m.listErr = err
}

// SetDeleteError sets an error to return on DeleteByDocument
func (m *MockChunkRepository) SetDeleteError(err error) {
	m.deleteErr = err
}

// SetSearchError sets an error to return on SearchSimilar
func (m *MockChunkRepository) SetSearchError(err error) {
	m.searchErr = err
}

// Create creates a new chunk
func (m *MockChunkRepository) Create(ctx context.Context, chunk *memory.Chunk) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.chunks[chunk.ID] = chunk
	if m.byDocument[chunk.DocumentID] == nil {
		m.byDocument[chunk.DocumentID] = []*memory.Chunk{}
	}
	m.byDocument[chunk.DocumentID] = append(m.byDocument[chunk.DocumentID], chunk)
	return nil
}

// CreateBatch creates multiple chunks in a single transaction
func (m *MockChunkRepository) CreateBatch(ctx context.Context, chunks []*memory.Chunk) error {
	if m.createBatchErr != nil {
		return m.createBatchErr
	}
	for _, chunk := range chunks {
		m.chunks[chunk.ID] = chunk
		if m.byDocument[chunk.DocumentID] == nil {
			m.byDocument[chunk.DocumentID] = []*memory.Chunk{}
		}
		m.byDocument[chunk.DocumentID] = append(m.byDocument[chunk.DocumentID], chunk)
	}
	return nil
}

// GetByID retrieves a chunk by ID
func (m *MockChunkRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Chunk, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	chunk, ok := m.chunks[id]
	if !ok {
		return nil, &NotFoundError{Message: "chunk not found"}
	}
	return chunk, nil
}

// ListByDocument lists chunks for a document
func (m *MockChunkRepository) ListByDocument(ctx context.Context, documentID uuid.UUID) ([]*memory.Chunk, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	chunks, ok := m.byDocument[documentID]
	if !ok {
		return []*memory.Chunk{}, nil
	}
	return chunks, nil
}

// DeleteByDocument deletes all chunks for a document
func (m *MockChunkRepository) DeleteByDocument(ctx context.Context, documentID uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.byDocument, documentID)
	// Remove chunks from main map
	for id, chunk := range m.chunks {
		if chunk.DocumentID == documentID {
			delete(m.chunks, id)
		}
	}
	return nil
}

// SearchSimilar searches for similar chunks using vector similarity
func (m *MockChunkRepository) SearchSimilar(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int, filters *SearchFilters) ([]*memory.Chunk, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	// Simple mock: return all chunks up to limit
	result := []*memory.Chunk{}
	count := 0
	for _, chunk := range m.chunks {
		if count >= limit {
			break
		}
		result = append(result, chunk)
		count++
	}
	return result, nil
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

