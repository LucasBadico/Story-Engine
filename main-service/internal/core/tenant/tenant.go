package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// Tenant represents a workspace/tenant entity
type Tenant struct {
	ID                 uuid.UUID
	Name               string
	Status             TenantStatus
	ActiveLLMProfileID *uuid.UUID // nullable, for future phases
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          *uuid.UUID // nullable
}

// NewTenant creates a new tenant with default values
func NewTenant(name string, createdBy *uuid.UUID) (*Tenant, error) {
	if name == "" {
		return nil, ErrTenantNameRequired
	}

	now := time.Now()
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Status:    TenantStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
	}, nil
}

// Validate validates the tenant entity
func (t *Tenant) Validate() error {
	if t.Name == "" {
		return ErrTenantNameRequired
	}
	if t.Status != TenantStatusActive && t.Status != TenantStatusSuspended && t.Status != TenantStatusDeleted {
		return ErrInvalidStatus
	}
	return nil
}

// IsActive returns true if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// UpdateName updates the tenant name
func (t *Tenant) UpdateName(name string) error {
	if name == "" {
		return ErrTenantNameRequired
	}
	t.Name = name
	t.UpdatedAt = time.Now()
	return nil
}

// Suspend suspends the tenant
func (t *Tenant) Suspend() {
	t.Status = TenantStatusSuspended
	t.UpdatedAt = time.Now()
}

// Activate activates the tenant
func (t *Tenant) Activate() {
	t.Status = TenantStatusActive
	t.UpdatedAt = time.Now()
}
