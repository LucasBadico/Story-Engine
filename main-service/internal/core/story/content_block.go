package story

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContentType represents the type of content block
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
	ContentTypeAudio ContentType = "audio"
	ContentTypeEmbed ContentType = "embed"
	ContentTypeLink  ContentType = "link"
)

// ContentKind represents the kind/variation of content block
type ContentKind string

const (
	ContentKindFinal     ContentKind = "final"
	ContentKindAltA      ContentKind = "alt_a"
	ContentKindAltB      ContentKind = "alt_b"
	ContentKindCleaned   ContentKind = "cleaned"
	ContentKindLocalized ContentKind = "localized"
	ContentKindDraft     ContentKind = "draft"
	ContentKindThumbnail ContentKind = "thumbnail"
)

// ContentMetadata holds type-specific metadata
type ContentMetadata struct {
	// Text metadata
	WordCount *int `json:"word_count,omitempty"`

	// Image metadata
	AltText  *string `json:"alt_text,omitempty"`
	Caption  *string `json:"caption,omitempty"`
	Width    *int    `json:"width,omitempty"`
	Height   *int    `json:"height,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`

	// Video metadata
	Provider     *string `json:"provider,omitempty"`
	VideoID      *string `json:"video_id,omitempty"`
	Duration     *int    `json:"duration,omitempty"` // in seconds
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	// Audio metadata
	Transcript *string `json:"transcript,omitempty"`

	// Embed metadata
	HTML *string `json:"html,omitempty"`

	// Link metadata
	Title    *string `json:"title,omitempty"`
	Desc     *string `json:"description,omitempty"`
	ImageURL *string `json:"image_url,omitempty"`
	SiteName *string `json:"site_name,omitempty"`
}

// ContentBlock represents a content block entity (unified prose/image/video/etc)
type ContentBlock struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	ChapterID *uuid.UUID      `json:"chapter_id,omitempty"` // nullable - can be related via references
	OrderNum  *int            `json:"order_num,omitempty"`   // nullable - only needed if chapter_id is set
	Type      ContentType     `json:"type"`
	Kind      ContentKind     `json:"kind"`
	Content   string          `json:"content"` // text content or URL depending on type
	Metadata  ContentMetadata `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// NewContentBlock creates a new content block
func NewContentBlock(tenantID uuid.UUID, chapterID *uuid.UUID, orderNum *int, contentType ContentType, kind ContentKind, content string) (*ContentBlock, error) {
	if !isValidContentType(contentType) {
		return nil, ErrInvalidContentType
	}
	if !isValidContentKind(kind) {
		return nil, ErrInvalidContentKind
	}
	if orderNum != nil && *orderNum < 1 {
		return nil, ErrInvalidOrderNumber
	}
	if content == "" {
		return nil, ErrContentRequired
	}

	now := time.Now()
	block := &ContentBlock{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		Type:      contentType,
		Kind:      kind,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set default metadata based on type
	if contentType == ContentTypeText {
		wordCount := countWords(content)
		block.Metadata.WordCount = &wordCount
	}

	return block, nil
}

// Validate validates the content block entity
func (c *ContentBlock) Validate() error {
	if !isValidContentType(c.Type) {
		return ErrInvalidContentType
	}
	if !isValidContentKind(c.Kind) {
		return ErrInvalidContentKind
	}
	if c.OrderNum != nil && *c.OrderNum < 1 {
		return ErrInvalidOrderNumber
	}
	if c.Content == "" {
		return ErrContentRequired
	}
	if c.Type == ContentTypeText && c.Metadata.WordCount != nil && *c.Metadata.WordCount < 0 {
		return ErrInvalidWordCount
	}
	if c.Type == ContentTypeImage {
		if c.Metadata.Width != nil && *c.Metadata.Width < 1 {
			return ErrInvalidDimensions
		}
		if c.Metadata.Height != nil && *c.Metadata.Height < 1 {
			return ErrInvalidDimensions
		}
	}
	return nil
}

// UpdateContent updates the content and recalculates metadata if needed
func (c *ContentBlock) UpdateContent(content string) error {
	if content == "" {
		return ErrContentRequired
	}
	c.Content = content
	if c.Type == ContentTypeText {
		wordCount := countWords(content)
		c.Metadata.WordCount = &wordCount
	}
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateMetadata updates the metadata
func (c *ContentBlock) UpdateMetadata(metadata ContentMetadata) error {
	// Validate dimensions if present
	if c.Type == ContentTypeImage {
		if metadata.Width != nil && *metadata.Width < 1 {
			return ErrInvalidDimensions
		}
		if metadata.Height != nil && *metadata.Height < 1 {
			return ErrInvalidDimensions
		}
	}
	c.Metadata = metadata
	c.UpdatedAt = time.Now()
	return nil
}

// MetadataToJSON converts metadata to JSON bytes for database storage
func (c *ContentBlock) MetadataToJSON() ([]byte, error) {
	return json.Marshal(c.Metadata)
}

// MetadataFromJSON loads metadata from JSON bytes
func (c *ContentBlock) MetadataFromJSON(data []byte) error {
	return json.Unmarshal(data, &c.Metadata)
}

func isValidContentType(contentType ContentType) bool {
	return contentType == ContentTypeText ||
		contentType == ContentTypeImage ||
		contentType == ContentTypeVideo ||
		contentType == ContentTypeAudio ||
		contentType == ContentTypeEmbed ||
		contentType == ContentTypeLink
}

func isValidContentKind(kind ContentKind) bool {
	return kind == ContentKindFinal ||
		kind == ContentKindAltA ||
		kind == ContentKindAltB ||
		kind == ContentKindCleaned ||
		kind == ContentKindLocalized ||
		kind == ContentKindDraft ||
		kind == ContentKindThumbnail
}

func countWords(text string) int {
	if text == "" {
		return 0
	}
	words := strings.Fields(text)
	return len(words)
}
