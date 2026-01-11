package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MockIngestionQueue is a mock implementation of IngestionQueue for testing
type MockIngestionQueue struct {
	items             map[string]*QueueItem // key: "tenantID:sourceType:sourceID"
	processingItems   map[string]*QueueItem
	tenants           map[uuid.UUID]bool
	processingTenants map[uuid.UUID]bool

	enqueueErr error
	popErr     error
	listErr    error
}

// NewMockIngestionQueue creates a new mock ingestion queue
func NewMockIngestionQueue() *MockIngestionQueue {
	return &MockIngestionQueue{
		items:             make(map[string]*QueueItem),
		processingItems:   make(map[string]*QueueItem),
		tenants:           make(map[uuid.UUID]bool),
		processingTenants: make(map[uuid.UUID]bool),
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
		TenantID:   tenantID,
		SourceType: sourceType,
		SourceID:   sourceID,
		Timestamp:  time.Now(),
	}
	m.tenants[tenantID] = true
	return nil
}

// Ack removes an item from processing
func (m *MockIngestionQueue) Ack(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := tenantID.String() + ":" + sourceType + ":" + sourceID.String()
	delete(m.processingItems, key)
	m.refreshProcessingTenants(tenantID)
	return nil
}

// Release moves an item from processing back to the queue
func (m *MockIngestionQueue) Release(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := tenantID.String() + ":" + sourceType + ":" + sourceID.String()
	if item, ok := m.processingItems[key]; ok {
		delete(m.processingItems, key)
		item.Timestamp = time.Now()
		m.items[key] = item
		m.tenants[tenantID] = true
	}
	m.refreshProcessingTenants(tenantID)
	return nil
}

// RequeueExpiredProcessing moves expired processing items back to the queue
func (m *MockIngestionQueue) RequeueExpiredProcessing(ctx context.Context, tenantID uuid.UUID, expiredBefore time.Time, limit int) (int, error) {
	count := 0
	for key, item := range m.processingItems {
		if item.TenantID == tenantID && item.Timestamp.Before(expiredBefore) && count < limit {
			delete(m.processingItems, key)
			item.Timestamp = time.Now()
			m.items[key] = item
			m.tenants[tenantID] = true
			count++
		}
	}
	m.refreshProcessingTenants(tenantID)
	return count, nil
}

// Remove removes an item from both queue and processing
func (m *MockIngestionQueue) Remove(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := tenantID.String() + ":" + sourceType + ":" + sourceID.String()
	delete(m.items, key)
	delete(m.processingItems, key)
	m.refreshTenants(tenantID)
	m.refreshProcessingTenants(tenantID)
	return nil
}

// PopStable moves stable items from the queue to processing
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
			item.Timestamp = time.Now()
			m.processingItems[key] = item
			count++
		}
	}
	m.refreshTenants(tenantID)
	if count > 0 {
		m.processingTenants[tenantID] = true
	}
	return result, nil
}

// PopStableBySourceType moves stable items for a source type from the queue to processing
func (m *MockIngestionQueue) PopStableBySourceType(ctx context.Context, tenantID uuid.UUID, sourceType string, stableAt time.Time, limit int) ([]*QueueItem, error) {
	if m.popErr != nil {
		return nil, m.popErr
	}
	var result []*QueueItem
	count := 0
	for key, item := range m.items {
		if item.TenantID == tenantID && item.SourceType == sourceType && item.Timestamp.Before(stableAt) && count < limit {
			result = append(result, item)
			delete(m.items, key)
			item.Timestamp = time.Now()
			m.processingItems[key] = item
			count++
		}
	}
	m.refreshTenants(tenantID)
	if count > 0 {
		m.processingTenants[tenantID] = true
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

// ListTenantsWithProcessingItems returns tenants with processing items
func (m *MockIngestionQueue) ListTenantsWithProcessingItems(ctx context.Context) ([]uuid.UUID, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []uuid.UUID
	for tenantID := range m.processingTenants {
		result = append(result, tenantID)
	}
	return result, nil
}

func (m *MockIngestionQueue) refreshTenants(tenantID uuid.UUID) {
	for _, item := range m.items {
		if item.TenantID == tenantID {
			m.tenants[tenantID] = true
			return
		}
	}
	delete(m.tenants, tenantID)
}

func (m *MockIngestionQueue) refreshProcessingTenants(tenantID uuid.UUID) {
	for _, item := range m.processingItems {
		if item.TenantID == tenantID {
			m.processingTenants[tenantID] = true
			return
		}
	}
	delete(m.processingTenants, tenantID)
}
