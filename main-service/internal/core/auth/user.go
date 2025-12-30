package auth

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User represents a user entity
type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new user with default values
func NewUser(email, name string) (*User, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}
	if name == "" {
		return nil, ErrNameRequired
	}

	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Name:      name,
		Status:    UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmailRequired
	}
	if u.Name == "" {
		return ErrNameRequired
	}
	if u.Status != UserStatusActive && u.Status != UserStatusSuspended && u.Status != UserStatusDeleted {
		return ErrInvalidStatus
	}
	return nil
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// UpdateEmail updates the user email
func (u *User) UpdateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateName updates the user name
func (u *User) UpdateName(name string) error {
	if name == "" {
		return ErrNameRequired
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	return nil
}

// Suspend suspends the user
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now()
}

// Activate activates the user
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}
