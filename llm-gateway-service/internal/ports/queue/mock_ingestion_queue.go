package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MockIngestionQueue is a mock implementation of IngestionQueue for testing
type MockIngestionQueue struct {
	items map[string]*QueueItem // key: "tenantID:sourceType:sourceID"
	tenants map[uuid.UUID]bool
	
	enqueueErr error
	popErr     error
	listErr    error
}

// NewMockIngestionQueue creates a new mock ingestion queue
func NewMockIngestionQueue() *MockIngestionQueue {
	return &MockIngestionQueue{
		items:   make(map[string]*QueueItem),
		tenants: make(map[uuid.UUID]bool),
	}
}

// SetPushError sets an error to return on Push
func (m *MockIngestionQueue) SetPushError(err error) {
	m.enqueueErr = err
}

// SetPopError sets an error to return on PopStable
func (m *MockIngestionQueue) SetPopError(err error) {
	m.popErr = err
}

// SetListError sets an error to return on ListTenantsWithItems
func (m *MockIngestionQueue) SetListError(err error) {
	m.listErr = err
}


// Push adds/updates item with current timestamp (debounce reset)
func (m *MockIngestionQueue) Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	if m.enqueueErr != nil {
		return m.enqueueErr
	}
	key := tenantID.String() + ":" + sourceType + ":" + sourceID.String()
	m.items[key] = &QueueItem{
		TenantID:  tenantID,
		SourceType: sourceType,
		SourceID:  sourceID,
		Timestamp: time.Now(),
	}
	m.tenants[tenantID] = true
	return nil
}

// Remove removes an item from queue
func (m *MockIngestionQueue) Remove(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := tenantID.String() + ":" + sourceType + ":" + sourceID.String()
	delete(m.items, key)
	// Check if tenant still has items
	hasItems := false
	for _, item := range m.items {
		if item.TenantID == tenantID {
			hasItems = true
			break
		}
	}
	if !hasItems {
		delete(m.tenants, tenantID)
	}
	return nil
}

// PopStable removes and returns stable items from the queue
func (m *MockIngestionQueue) PopStable(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]*QueueItem, error) {
	if m.popErr != nil {
		return nil, m.popErr
	}
	var result []*QueueItem
	count := 0
	for key, item := range m.items {
		if item.TenantID == tenantID && item.Timestamp.Before(stableAt) && count < limit {
			result = append(result, item)
			delete(m.items, key)
			count++
		}
	}
	// If no items left for tenant, remove from tenants map
	hasItems := false
	for _, item := range m.items {
		if item.TenantID == tenantID {
			hasItems = true
			break
		}
	}
	if !hasItems {
		delete(m.tenants, tenantID)
	}
	return result, nil
}

// ListTenantsWithItems returns a list of tenant IDs that have items in the queue
func (m *MockIngestionQueue) ListTenantsWithItems(ctx context.Context) ([]uuid.UUID, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []uuid.UUID
	for tenantID := range m.tenants {
		result = append(result, tenantID)
	}
	return result, nil
}

