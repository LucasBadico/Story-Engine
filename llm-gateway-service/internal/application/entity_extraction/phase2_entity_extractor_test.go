package entity_extraction

import (
	"strings"
	"testing"
)

func TestFindEvidenceOffset(t *testing.T) {
	text := "Aria met the Crimson Order inside the Obsidian Tower."
	start, end := findEvidenceOffset(text, "Crimson Order")
	if start < 0 || end <= start {
		t.Fatalf("expected evidence offsets, got %d-%d", start, end)
	}
	if text[start:end] != "Crimson Order" {
		t.Fatalf("expected exact match, got %q", text[start:end])
	}
}

func TestBuildPhase2EntityExtractorPrompt(t *testing.T) {
	_, err := buildPhase2EntityExtractorPrompt("unknown", "text", "", nil, 3)
	if err == nil {
		t.Fatalf("expected error for unsupported type")
	}

	prompt, err := buildPhase2EntityExtractorPrompt("character", "Text", "", []Phase2KnownEntity{{Name: "Aria", Summary: "Mage"}}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt == "" {
		t.Fatalf("expected prompt content")
	}
	if !containsAll(prompt, []string{"Text", "candidates", "ALREADY FOUND"}) {
		t.Fatalf("expected prompt to include template content")
	}
}

func containsAll(text string, tokens []string) bool {
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if !strings.Contains(text, token) {
			return false
		}
	}
	return true
}
