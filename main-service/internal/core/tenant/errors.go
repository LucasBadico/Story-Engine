package tenant

import "errors"

var (
	ErrTenantNameRequired = errors.New("tenant name is required")
	ErrInvalidStatus      = errors.New("invalid tenant status")
)

