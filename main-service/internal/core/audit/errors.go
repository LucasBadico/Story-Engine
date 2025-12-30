package audit

import "errors"

var (
	ErrActionRequired     = errors.New("action is required")
	ErrEntityTypeRequired = errors.New("entity type is required")
)

