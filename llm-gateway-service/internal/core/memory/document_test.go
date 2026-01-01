package memory

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewDocument(t *testing.T) {
	tenantID := uuid.New()
	sourceID := uuid.New()
	title := "Test Document"
	content := "Test content"

	doc := NewDocument(tenantID, SourceTypeStory, sourceID, title, content)

	if doc.ID == uuid.Nil {
		t.Error("Expected document ID to be set")
	}
	if doc.TenantID != tenantID {
		t.Errorf("Expected TenantID %v, got %v", tenantID, doc.TenantID)
	}
	if doc.SourceType != SourceTypeStory {
		t.Errorf("Expected SourceType %q, got %q", SourceTypeStory, doc.SourceType)
	}
	if doc.SourceID != sourceID {
		t.Errorf("Expected SourceID %v, got %v", sourceID, doc.SourceID)
	}
	if doc.Title != title {
		t.Errorf("Expected Title %q, got %q", title, doc.Title)
	}
	if doc.Content != content {
		t.Errorf("Expected Content %q, got %q", content, doc.Content)
	}
}

func TestDocument_Validate(t *testing.T) {
	tenantID := uuid.New()
	sourceID := uuid.New()
	doc := NewDocument(tenantID, SourceTypeStory, sourceID, "Title", "Content")

	if err := doc.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDocument_Validate_InvalidTenantID(t *testing.T) {
	sourceID := uuid.New()
	doc := NewDocument(uuid.Nil, SourceTypeStory, sourceID, "Title", "Content")

	if err := doc.Validate(); err != ErrTenantIDRequired {
		t.Errorf("Expected ErrTenantIDRequired, got %v", err)
	}
}

func TestDocument_Validate_InvalidSourceID(t *testing.T) {
	tenantID := uuid.New()
	doc := NewDocument(tenantID, SourceTypeStory, uuid.Nil, "Title", "Content")

	if err := doc.Validate(); err != ErrSourceIDRequired {
		t.Errorf("Expected ErrSourceIDRequired, got %v", err)
	}
}

func TestDocument_Validate_EmptyContent(t *testing.T) {
	tenantID := uuid.New()
	sourceID := uuid.New()
	doc := NewDocument(tenantID, SourceTypeStory, sourceID, "Title", "")

	if err := doc.Validate(); err != ErrContentRequired {
		t.Errorf("Expected ErrContentRequired, got %v", err)
	}
}

func TestDocument_SourceTypes(t *testing.T) {
	tenantID := uuid.New()
	sourceID := uuid.New()

	sourceTypes := []SourceType{
		SourceTypeStory,
		SourceTypeChapter,
		SourceTypeScene,
		SourceTypeBeat,
		SourceTypeProseBlock,
	}

	for _, st := range sourceTypes {
		doc := NewDocument(tenantID, st, sourceID, "Title", "Content")
		if doc.SourceType != st {
			t.Errorf("Expected SourceType %q, got %q", st, doc.SourceType)
		}
		if err := doc.Validate(); err != nil {
			t.Errorf("Expected no error for SourceType %q, got %v", st, err)
		}
	}
}

