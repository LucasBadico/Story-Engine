package versioning

import "errors"

var (
	ErrSourceStoryRequired = errors.New("source story is required")
	ErrInvalidVersionNumber = errors.New("invalid version number")
)

