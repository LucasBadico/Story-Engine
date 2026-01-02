package story

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrImageBlockURLRequired = errors.New("image URL is required")
	ErrInvalidImageKind      = errors.New("invalid image kind")
	ErrInvalidImageOrderNumber = errors.New("order number must be greater than 0")
	ErrInvalidDimensions     = errors.New("width and height must be positive")
)

// ImageKind represents the kind of image block
type ImageKind string

const (
	ImageKindFinal     ImageKind = "final"
	ImageKindAltA      ImageKind = "alt_a"
	ImageKindAltB      ImageKind = "alt_b"
	ImageKindDraft     ImageKind = "draft"
	ImageKindThumbnail ImageKind = "thumbnail"
)

// ImageBlock represents an image block entity
type ImageBlock struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ChapterID *uuid.UUID `json:"chapter_id,omitempty"` // nullable - can be related via references
	OrderNum  *int       `json:"order_num,omitempty"`  // nullable - only needed if chapter_id is set
	Kind      ImageKind  `json:"kind"`
	ImageURL  string     `json:"image_url"`
	AltText   *string    `json:"alt_text,omitempty"`
	Caption   *string    `json:"caption,omitempty"`
	Width     *int       `json:"width,omitempty"`
	Height    *int       `json:"height,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewImageBlock creates a new image block
func NewImageBlock(tenantID uuid.UUID, chapterID *uuid.UUID, orderNum *int, kind ImageKind, imageURL string) (*ImageBlock, error) {
	if !isValidImageKind(kind) {
		return nil, ErrInvalidImageKind
	}
	if imageURL == "" {
		return nil, ErrImageBlockURLRequired
	}
	if orderNum != nil && *orderNum < 1 {
		return nil, ErrInvalidImageOrderNumber
	}

	now := time.Now()
	return &ImageBlock{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		Kind:      kind,
		ImageURL:  imageURL,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the image block entity
func (i *ImageBlock) Validate() error {
	if !isValidImageKind(i.Kind) {
		return ErrInvalidImageKind
	}
	if i.ImageURL == "" {
		return ErrImageBlockURLRequired
	}
	if i.OrderNum != nil && *i.OrderNum < 1 {
		return ErrInvalidImageOrderNumber
	}
	if i.Width != nil && *i.Width < 1 {
		return ErrInvalidDimensions
	}
	if i.Height != nil && *i.Height < 1 {
		return ErrInvalidDimensions
	}
	return nil
}

// UpdateImageURL updates the image URL
func (i *ImageBlock) UpdateImageURL(imageURL string) error {
	if imageURL == "" {
		return ErrImageBlockURLRequired
	}
	i.ImageURL = imageURL
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateAltText updates the alt text
func (i *ImageBlock) UpdateAltText(altText *string) {
	i.AltText = altText
	i.UpdatedAt = time.Now()
}

// UpdateCaption updates the caption
func (i *ImageBlock) UpdateCaption(caption *string) {
	i.Caption = caption
	i.UpdatedAt = time.Now()
}

// UpdateDimensions updates width and height
func (i *ImageBlock) UpdateDimensions(width, height *int) error {
	if width != nil && *width < 1 {
		return ErrInvalidDimensions
	}
	if height != nil && *height < 1 {
		return ErrInvalidDimensions
	}
	i.Width = width
	i.Height = height
	i.UpdatedAt = time.Now()
	return nil
}

func isValidImageKind(kind ImageKind) bool {
	return kind == ImageKindFinal ||
		kind == ImageKindAltA ||
		kind == ImageKindAltB ||
		kind == ImageKindDraft ||
		kind == ImageKindThumbnail
}

