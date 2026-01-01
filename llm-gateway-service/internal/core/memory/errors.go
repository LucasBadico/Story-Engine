package memory

import "errors"

var (
	ErrTenantIDRequired   = errors.New("tenant_id is required")
	ErrSourceIDRequired   = errors.New("source_id is required")
	ErrContentRequired    = errors.New("content is required")
	ErrDocumentIDRequired = errors.New("document_id is required")
	ErrInvalidChunkIndex  = errors.New("chunk_index must be >= 0")
	ErrInvalidSourceType  = errors.New("invalid source_type")
)

