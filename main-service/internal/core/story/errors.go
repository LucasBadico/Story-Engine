package story

import "errors"

var (
	ErrTitleRequired         = errors.New("title is required")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrInvalidVersionNumber  = errors.New("invalid version number")
	ErrInvalidChapterNumber  = errors.New("invalid chapter number")
	ErrInvalidOrderNumber    = errors.New("invalid order number")
	ErrInvalidBeatType       = errors.New("invalid beat type")
	ErrInvalidProseKind      = errors.New("invalid prose kind")
	ErrInvalidWordCount      = errors.New("invalid word count")
	ErrInvalidEntityType     = errors.New("invalid entity type")
)

