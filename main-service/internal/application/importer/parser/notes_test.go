package parser

import (
	"testing"
)

const fixtureNotes = `<!-- generated: split from Google Doc export -->
<!-- source document: 1_0_arc_build_up_prologue -->
<!-- tab path: Notes -->

# Notes

3st March 2026: I'm doing small edition to the early chapters.
\- last line edited to hold the Day Zero Anchor



6th March 2026: Me again doing editions, added a little wink to later chapter events. Sorry early readers and thank you to keep up with my shenanigans.
`

func TestParseNotes(t *testing.T) {
	notes, err := ParseNotes(fixtureNotes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("len(notes) = %d, want 2", len(notes))
	}

	if notes[0].Date != "3st March 2026" {
		t.Errorf("notes[0].Date = %q, want %q", notes[0].Date, "3st March 2026")
	}
	if notes[0].Text == "" {
		t.Error("notes[0].Text is empty")
	}

	if notes[1].Date != "6th March 2026" {
		t.Errorf("notes[1].Date = %q, want %q", notes[1].Date, "6th March 2026")
	}
	if notes[1].Text == "" {
		t.Error("notes[1].Text is empty")
	}
}
