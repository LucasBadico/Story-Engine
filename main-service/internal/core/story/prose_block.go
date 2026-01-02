package story

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProseKind represents the kind of prose block
type ProseKind string

const (
	ProseKindFinal     ProseKind = "final"
	ProseKindAltA      ProseKind = "alt_a"
	ProseKindAltB      ProseKind = "alt_b"
	ProseKindCleaned   ProseKind = "cleaned"
	ProseKindLocalized ProseKind = "localized"
	ProseKindDraft     ProseKind = "draft"
)

// ProseBlock represents a prose block entity
type ProseBlock struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ChapterID *uuid.UUID `json:"chapter_id,omitempty"` // nullable - can be related via references
	OrderNum  *int       `json:"order_num,omitempty"`   // nullable - only needed if chapter_id is set
	Kind      ProseKind  `json:"kind"`
	Content   string     `json:"content"`
	WordCount int        `json:"word_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewProseBlock creates a new prose block
func NewProseBlock(tenantID uuid.UUID, chapterID *uuid.UUID, orderNum *int, kind ProseKind, content string) (*ProseBlock, error) {
	if !isValidProseKind(kind) {
		return nil, ErrInvalidProseKind
	}
	if orderNum != nil && *orderNum < 1 {
		return nil, ErrInvalidOrderNumber
	}

	now := time.Now()
	wordCount := countWords(content)

	return &ProseBlock{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		Kind:      kind,
		Content:   content,
		WordCount: wordCount,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the prose block entity
func (p *ProseBlock) Validate() error {
	if !isValidProseKind(p.Kind) {
		return ErrInvalidProseKind
	}
	if p.OrderNum != nil && *p.OrderNum < 1 {
		return ErrInvalidOrderNumber
	}
	if p.WordCount < 0 {
		return ErrInvalidWordCount
	}
	return nil
}

// UpdateContent updates the prose content and recalculates word count
func (p *ProseBlock) UpdateContent(content string) {
	p.Content = content
	p.WordCount = countWords(content)
	p.UpdatedAt = time.Now()
}

func isValidProseKind(kind ProseKind) bool {
	return kind == ProseKindFinal ||
		kind == ProseKindAltA ||
		kind == ProseKindAltB ||
		kind == ProseKindCleaned ||
		kind == ProseKindLocalized ||
		kind == ProseKindDraft
}

func countWords(text string) int {
	if text == "" {
		return 0
	}
	words := strings.Fields(text)
	return len(words)
}
