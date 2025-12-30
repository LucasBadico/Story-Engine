package auth

import (
	"time"

	"github.com/google/uuid"
)

// Role represents the role of a user in a tenant
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Status represents the status of a membership
type MembershipStatus string

const (
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended"
	MembershipStatusDeleted   MembershipStatus = "deleted"
)

// Membership represents the relationship between a user and a tenant
type Membership struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Role      Role
	Status    MembershipStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewMembership creates a new membership with default values
func NewMembership(tenantID, userID uuid.UUID, role Role) (*Membership, error) {
	if role != RoleOwner && role != RoleAdmin && role != RoleEditor && role != RoleViewer {
		return nil, ErrInvalidRole
	}

	now := time.Now()
	return &Membership{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the membership entity
func (m *Membership) Validate() error {
	if m.Role != RoleOwner && m.Role != RoleAdmin && m.Role != RoleEditor && m.Role != RoleViewer {
		return ErrInvalidRole
	}
	if m.Status != MembershipStatusActive && m.Status != MembershipStatusSuspended && m.Status != MembershipStatusDeleted {
		return ErrInvalidStatus
	}
	return nil
}

// IsActive returns true if the membership is active
func (m *Membership) IsActive() bool {
	return m.Status == MembershipStatusActive
}

// IsOwner returns true if the membership role is owner
func (m *Membership) IsOwner() bool {
	return m.Role == RoleOwner
}

// UpdateRole updates the membership role
func (m *Membership) UpdateRole(role Role) error {
	if role != RoleOwner && role != RoleAdmin && role != RoleEditor && role != RoleViewer {
		return ErrInvalidRole
	}
	m.Role = role
	m.UpdatedAt = time.Now()
	return nil
}

// Suspend suspends the membership
func (m *Membership) Suspend() {
	m.Status = MembershipStatusSuspended
	m.UpdatedAt = time.Now()
}

// Activate activates the membership
func (m *Membership) Activate() {
	m.Status = MembershipStatusActive
	m.UpdatedAt = time.Now()
}
