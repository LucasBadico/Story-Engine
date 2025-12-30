package errors

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrValidation    = errors.New("validation error")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
)

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// AlreadyExistsError represents a resource that already exists
type AlreadyExistsError struct {
	Resource string
	Field    string
	Value    string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists with %s: %s", e.Resource, e.Field, e.Value)
}

func (e *AlreadyExistsError) Unwrap() error {
	return ErrAlreadyExists
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	var notFound *NotFoundError
	return errors.As(err, &notFound) || errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is an AlreadyExistsError
func IsAlreadyExists(err error) bool {
	var exists *AlreadyExistsError
	return errors.As(err, &exists) || errors.Is(err, ErrAlreadyExists)
}

// IsValidation checks if an error is a ValidationError
func IsValidation(err error) bool {
	var validation *ValidationError
	return errors.As(err, &validation) || errors.Is(err, ErrValidation)
}

