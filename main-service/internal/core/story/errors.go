package story

import "errors"

var (
	ErrTitleRequired         = errors.New("title is required")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrInvalidVersionNumber  = errors.New("invalid version number")
	ErrInvalidChapterNumber  = errors.New("invalid chapter number")
	ErrInvalidOrderNumber    = errors.New("invalid order number")
	ErrInvalidBeatType       = errors.New("invalid beat type")
	ErrInvalidProseKind      = errors.New("invalid prose kind") // deprecated, use ErrInvalidContentKind
	ErrInvalidContentType    = errors.New("invalid content type")
	ErrInvalidContentKind    = errors.New("invalid content kind")
	ErrContentRequired       = errors.New("content is required")
	ErrInvalidWordCount      = errors.New("invalid word count")
	ErrInvalidEntityType     = errors.New("invalid entity type")
	ErrInvalidDimensions     = errors.New("width and height must be positive")
)

