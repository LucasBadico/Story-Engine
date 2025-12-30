package auth

import "errors"

var (
	ErrEmailRequired = errors.New("email is required")
	ErrNameRequired  = errors.New("name is required")
	ErrInvalidRole   = errors.New("invalid role")
	ErrInvalidStatus = errors.New("invalid status")
)

