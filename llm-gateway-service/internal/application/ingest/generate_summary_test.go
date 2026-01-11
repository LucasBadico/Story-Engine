package ingest

import (
	"strings"
	"testing"
)

func TestBuildGenerateEntitySummaryPrompt(t *testing.T) {
	prompt := buildGenerateSummaryPrompt(GenerateSummaryInput{
		EntityType: "character",
		Name:       "Aria",
		Contents:   []string{"Aria is a mage."},
		Context:    "",
	}, 3)
	if prompt == "" {
		t.Fatalf("expected prompt content")
	}
	if !containsAll(prompt, []string{"Aria", "character", "summaries", "CONTENT BLOCKS"}) {
		t.Fatalf("expected prompt to include placeholders")
	}
}

func TestParseGenerateSummaryOutput(t *testing.T) {
	raw := `{"summaries":["Aria the mage","Aria the hero"]}`
	output, err := parseGenerateSummaryOutput(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(output.Summaries))
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
