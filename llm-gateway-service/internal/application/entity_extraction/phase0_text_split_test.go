package entity_extraction

import "testing"

func TestSplitTextIntoParagraphChunks(t *testing.T) {
	text := "Paragraph one.\nStill paragraph one.\n\nParagraph two."

	output, err := SplitTextIntoParagraphChunks(Phase0TextSplitInput{
		Text:          text,
		MaxChunkChars: 100,
		OverlapChars:  10,
	})
	if err != nil {
		t.Fatalf("split failed: %v", err)
	}
	if len(output.Paragraphs) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(output.Paragraphs))
	}
	if len(output.Paragraphs[0].Chunks) != 1 {
		t.Fatalf("expected 1 chunk in paragraph 0, got %d", len(output.Paragraphs[0].Chunks))
	}
	if len(output.Paragraphs[1].Chunks) != 1 {
		t.Fatalf("expected 1 chunk in paragraph 1, got %d", len(output.Paragraphs[1].Chunks))
	}
}

func TestSplitTextIntoParagraphChunks_SentenceSplit(t *testing.T) {
	text := "First sentence. Second sentence is longer. Third sentence."

	output, err := SplitTextIntoParagraphChunks(Phase0TextSplitInput{
		Text:          text,
		MaxChunkChars: 35,
		OverlapChars:  5,
	})
	if err != nil {
		t.Fatalf("split failed: %v", err)
	}
	if len(output.Paragraphs) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(output.Paragraphs))
	}
	if len(output.Paragraphs[0].Chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(output.Paragraphs[0].Chunks))
	}
}

func TestSplitTextIntoParagraphChunks_FallbackWindow(t *testing.T) {
	text := "ThisSentenceIsWayTooLongWithoutAnyBreaksSoWeForceWindowSplitting."

	output, err := SplitTextIntoParagraphChunks(Phase0TextSplitInput{
		Text:          text,
		MaxChunkChars: 10,
		OverlapChars:  2,
	})
	if err != nil {
		t.Fatalf("split failed: %v", err)
	}
	if len(output.Paragraphs) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(output.Paragraphs))
	}
	if len(output.Paragraphs[0].Chunks) < 2 {
		t.Fatalf("expected fallback window chunks, got %d", len(output.Paragraphs[0].Chunks))
	}
}
