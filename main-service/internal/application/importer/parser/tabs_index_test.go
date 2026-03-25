package parser

import (
	"testing"
)

const fixtureTabsIndex = `Document: 1_0_arc_build_up_prologue
Document ID: 12zjS4VEVYYdQGV9iDVFmd45PnwJ9daKpEkm17utIBG8
Exported at: 2026-03-14T14:02:49.042Z

Tabs:

- Story Canon
- Chapter Log
- Content
- Notes
`

func TestParseTabsIndex(t *testing.T) {
	result, err := ParseTabsIndex(fixtureTabsIndex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DocumentName != "1_0_arc_build_up_prologue" {
		t.Errorf("DocumentName = %q, want %q", result.DocumentName, "1_0_arc_build_up_prologue")
	}

	if result.DocumentID != "12zjS4VEVYYdQGV9iDVFmd45PnwJ9daKpEkm17utIBG8" {
		t.Errorf("DocumentID = %q, want %q", result.DocumentID, "12zjS4VEVYYdQGV9iDVFmd45PnwJ9daKpEkm17utIBG8")
	}

	if len(result.Tabs) != 4 {
		t.Fatalf("len(Tabs) = %d, want 4", len(result.Tabs))
	}

	expectedTabs := []string{"Story Canon", "Chapter Log", "Content", "Notes"}
	for i, tab := range expectedTabs {
		if result.Tabs[i] != tab {
			t.Errorf("Tabs[%d] = %q, want %q", i, result.Tabs[i], tab)
		}
	}
}

func TestParseTabsIndex_Empty(t *testing.T) {
	_, err := ParseTabsIndex("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}
