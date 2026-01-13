package entities

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Phase0TextSplitInput struct {
	Text          string
	MaxChunkChars int
	OverlapChars  int
}

type Phase0TextSplitOutput struct {
	Paragraphs []ParagraphSplit
}

type ParagraphSplit struct {
	ParagraphID int
	StartOffset int
	EndOffset   int
	Text        string
	Chunks      []ParagraphChunk
}

type ParagraphChunk struct {
	ParagraphID int
	ChunkID     int
	StartOffset int
	EndOffset   int
	Text        string
}

func SplitTextIntoParagraphChunks(input Phase0TextSplitInput) (Phase0TextSplitOutput, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return Phase0TextSplitOutput{}, errors.New("text is required")
	}

	maxChars := input.MaxChunkChars
	if maxChars <= 0 {
		maxChars = 800
	}
	overlap := input.OverlapChars
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= maxChars {
		overlap = maxChars / 5
	}

	paragraphs := splitParagraphsWithOffsets(input.Text)
	results := make([]ParagraphSplit, 0, len(paragraphs))

	for paragraphID, paragraph := range paragraphs {
		chunks := splitParagraphIntoChunks(paragraph.Text, paragraph.StartOffset, paragraphID, maxChars, overlap)
		results = append(results, ParagraphSplit{
			ParagraphID: paragraphID,
			StartOffset: paragraph.StartOffset,
			EndOffset:   paragraph.EndOffset,
			Text:        paragraph.Text,
			Chunks:      chunks,
		})
	}

	return Phase0TextSplitOutput{Paragraphs: results}, nil
}

type paragraphSpan struct {
	StartOffset int
	EndOffset   int
	Text        string
}

func splitParagraphsWithOffsets(text string) []paragraphSpan {
	if text == "" {
		return nil
	}

	var spans []paragraphSpan
	i := 0
	for i < len(text) {
		for i < len(text) && isLineBreak(text, i) {
			i = advanceLineBreak(text, i)
		}
		if i >= len(text) {
			break
		}

		start := i
		for i < len(text) {
			if isLineBreak(text, i) {
				j := i
				count := 0
				for j < len(text) && isLineBreak(text, j) {
					j = advanceLineBreak(text, j)
					count++
				}
				if count >= 2 {
					break
				}
				i = j
				continue
			}
			i++
		}

		end := i
		span := strings.TrimRight(text[start:end], "\r\n")
		trimmedEnd := start + len(span)
		spans = append(spans, paragraphSpan{
			StartOffset: start,
			EndOffset:   trimmedEnd,
			Text:        span,
		})
	}

	return spans
}

func splitParagraphIntoChunks(paragraph string, paragraphOffset int, paragraphID int, maxChars int, overlap int) []ParagraphChunk {
	if paragraph == "" {
		return nil
	}

	if runeCount(paragraph) <= maxChars {
		return []ParagraphChunk{
			{
				ParagraphID: paragraphID,
				ChunkID:     0,
				StartOffset: paragraphOffset,
				EndOffset:   paragraphOffset + len(paragraph),
				Text:        paragraph,
			},
		}
	}

	sentences := splitSentencesWithOffsets(paragraph)
	chunks := make([]ParagraphChunk, 0, len(sentences))

	currentStart := -1
	currentEnd := -1
	chunkID := 0
	appendChunk := func(start int, end int) {
		if start < 0 || end <= start {
			return
		}
		chunks = append(chunks, ParagraphChunk{
			ParagraphID: paragraphID,
			ChunkID:     chunkID,
			StartOffset: paragraphOffset + start,
			EndOffset:   paragraphOffset + end,
			Text:        paragraph[start:end],
		})
		chunkID++
	}

	for _, sentence := range sentences {
		if runeCount(sentence.Text) > maxChars {
			return splitFixedWindow(paragraph, paragraphOffset, paragraphID, maxChars, overlap)
		}

		if currentStart == -1 {
			currentStart = sentence.Start
			currentEnd = sentence.End
			continue
		}

		candidate := paragraph[currentStart:sentence.End]
		if runeCount(candidate) <= maxChars {
			currentEnd = sentence.End
			continue
		}

		appendChunk(currentStart, currentEnd)
		currentStart = sentence.Start
		currentEnd = sentence.End
	}

	appendChunk(currentStart, currentEnd)
	if len(chunks) == 0 {
		return splitFixedWindow(paragraph, paragraphOffset, paragraphID, maxChars, overlap)
	}

	return chunks
}

type sentenceSpan struct {
	Start int
	End   int
	Text  string
}

func splitSentencesWithOffsets(text string) []sentenceSpan {
	var spans []sentenceSpan
	start := 0
	for idx, r := range text {
		if r == '.' || r == '!' || r == '?' {
			nextIndex := idx + utf8.RuneLen(r)
			if nextIndex < len(text) {
				nextRune, _ := utf8.DecodeRuneInString(text[nextIndex:])
				if !unicode.IsSpace(nextRune) {
					continue
				}
			}
			end := nextIndex
			segment := strings.TrimSpace(text[start:end])
			if segment != "" {
				segmentStart := start + strings.Index(text[start:end], segment)
				segmentEnd := segmentStart + len(segment)
				spans = append(spans, sentenceSpan{
					Start: segmentStart,
					End:   segmentEnd,
					Text:  text[segmentStart:segmentEnd],
				})
			}
			start = end
		}
	}

	if start < len(text) {
		segment := strings.TrimSpace(text[start:])
		if segment != "" {
			segmentStart := start + strings.Index(text[start:], segment)
			segmentEnd := segmentStart + len(segment)
			spans = append(spans, sentenceSpan{
				Start: segmentStart,
				End:   segmentEnd,
				Text:  text[segmentStart:segmentEnd],
			})
		}
	}

	return spans
}

func splitFixedWindow(text string, paragraphOffset int, paragraphID int, maxChars int, overlap int) []ParagraphChunk {
	var chunks []ParagraphChunk
	if text == "" {
		return chunks
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return chunks
	}

	step := maxChars - overlap
	if step <= 0 {
		step = maxChars
	}

	chunkID := 0
	for start := 0; start < len(runes); start += step {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}
		segment := string(runes[start:end])
		startByte := byteIndexFromRuneIndex(runes, start)
		endByte := byteIndexFromRuneIndex(runes, end)
		chunks = append(chunks, ParagraphChunk{
			ParagraphID: paragraphID,
			ChunkID:     chunkID,
			StartOffset: paragraphOffset + startByte,
			EndOffset:   paragraphOffset + endByte,
			Text:        segment,
		})
		chunkID++
		if end == len(runes) {
			break
		}
	}

	return chunks
}

func byteIndexFromRuneIndex(runes []rune, runeIndex int) int {
	if runeIndex <= 0 {
		return 0
	}
	if runeIndex >= len(runes) {
		return len(string(runes))
	}
	return len(string(runes[:runeIndex]))
}

func runeCount(text string) int {
	return utf8.RuneCountInString(text)
}

func isLineBreak(text string, i int) bool {
	if i < 0 || i >= len(text) {
		return false
	}
	return text[i] == '\n' || text[i] == '\r'
}

func advanceLineBreak(text string, i int) int {
	if i >= len(text) {
		return i
	}
	if text[i] == '\r' {
		if i+1 < len(text) && text[i+1] == '\n' {
			return i + 2
		}
		return i + 1
	}
	if text[i] == '\n' {
		return i + 1
	}
	return i + 1
}
