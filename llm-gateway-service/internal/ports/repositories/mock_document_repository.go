package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// MockDocumentRepository is a mock implementation of DocumentRepository for testing
type MockDocumentRepository struct {
	documents map[uuid.UUID]*memory.Document
	bySource  map[string]*memory.Document // key: "tenantID:sourceType:sourceID"
	
	createErr    error
	getByIDErr   error
	getBySourceErr error
	updateErr    error
	deleteErr    error
}

// NewMockDocumentRepository creates a new mock document repository
func NewMockDocumentRepository() *MockDocumentRepository {
	return &MockDocumentRepository{
		documents: make(map[uuid.UUID]*memory.Document),
		bySource:  make(map[string]*memory.Document),
	}
}

// SetCreateError sets an error to return on Create
func (m *MockDocumentRepository) SetCreateError(err error) {
	m.createErr = err
}

// SetGetByIDError sets an error to return on GetByID
func (m *MockDocumentRepository) SetGetByIDError(err error) {
	m.getByIDErr = err
}

// SetGetBySourceError sets an error to return on GetBySource
func (m *MockDocumentRepository) SetGetBySourceError(err error) {
	m.getBySourceErr = err
}

// SetUpdateError sets an error to return on Update
func (m *MockDocumentRepository) SetUpdateError(err error) {
	m.updateErr = err
}

// SetDeleteError sets an error to return on Delete
func (m *MockDocumentRepository) SetDeleteError(err error) {
	m.deleteErr = err
}

// makeSourceKey creates a key for bySource map
func (m *MockDocumentRepository) makeSourceKey(tenantID uuid.UUID, sourceType memory.SourceType, sourceID uuid.UUID) string {
	return tenantID.String() + ":" + string(sourceType) + ":" + sourceID.String()
}

// Create creates a new document
func (m *MockDocumentRepository) Create(ctx context.Context, doc *memory.Document) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.documents[doc.ID] = doc
	key := m.makeSourceKey(doc.TenantID, doc.SourceType, doc.SourceID)
	m.bySource[key] = doc
	return nil
}

// GetByID retrieves a document by ID
func (m *MockDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Document, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	doc, ok := m.documents[id]
	if !ok {
		return nil, &NotFoundError{Message: "document not found"}
	}
	return doc, nil
}

// GetBySource retrieves a document by source
func (m *MockDocumentRepository) GetBySource(ctx context.Context, tenantID uuid.UUID, sourceType memory.SourceType, sourceID uuid.UUID) (*memory.Document, error) {
	if m.getBySourceErr != nil {
		return nil, m.getBySourceErr
	}
	key := m.makeSourceKey(tenantID, sourceType, sourceID)
	doc, ok := m.bySource[key]
	if !ok {
		return nil, &NotFoundError{Message: "document not found"}
	}
	return doc, nil
}

// Update updates an existing document
func (m *MockDocumentRepository) Update(ctx context.Context, doc *memory.Document) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.documents[doc.ID] = doc
	key := m.makeSourceKey(doc.TenantID, doc.SourceType, doc.SourceID)
	m.bySource[key] = doc
	return nil
}

// Delete deletes a document
func (m *MockDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	doc, ok := m.documents[id]
	if ok {
		delete(m.documents, id)
		key := m.makeSourceKey(doc.TenantID, doc.SourceType, doc.SourceID)
		delete(m.bySource, key)
	}
	return nil
}

// ListByTenant lists documents for a tenant
func (m *MockDocumentRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*memory.Document, error) {
	var result []*memory.Document
	count := 0
	skipped := 0
	for _, doc := range m.documents {
		if doc.TenantID == tenantID {
			if skipped < offset {
				skipped++
				continue
			}
			if count >= limit {
				break
			}
			result = append(result, doc)
			count++
		}
	}
	return result, nil
}

