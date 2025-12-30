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
	ID        uuid.UUID
	SceneID   uuid.UUID
	Kind      ProseKind
	Content   string
	WordCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProseBlock creates a new prose block
func NewProseBlock(sceneID uuid.UUID, kind ProseKind, content string) (*ProseBlock, error) {
	if !isValidProseKind(kind) {
		return nil, ErrInvalidProseKind
	}

	now := time.Now()
	wordCount := countWords(content)

	return &ProseBlock{
		ID:        uuid.New(),
		SceneID:   sceneID,
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

