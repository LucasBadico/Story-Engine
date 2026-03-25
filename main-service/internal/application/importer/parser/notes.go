package parser

import (
	"regexp"
	"strings"
)

// ParseNotes parses a 04_notes.md file into a slice of ParsedNote.
// Each note starts with a date-like pattern (e.g., "3st March 2026:").
func ParseNotes(content string) ([]ParsedNote, error) {
	content = trimMarkdownComments(content)

	var notes []ParsedNote
	lines := strings.Split(content, "\n")

	// Match date patterns like "3st March 2026:", "6th March 2026:", "15th Jan 2025:" etc.
	dateRe := regexp.MustCompile(`^(\d{1,2}(?:st|nd|rd|th)\s+\w+\s+\d{4}):?\s*(.*)`)

	var currentDate string
	var currentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip # headers
		if strings.HasPrefix(trimmed, "# ") {
			continue
		}

		matches := dateRe.FindStringSubmatch(trimmed)
		if len(matches) >= 2 {
			// Save previous note
			if currentDate != "" {
				text := strings.TrimSpace(strings.Join(currentLines, "\n"))
				if text != "" {
					notes = append(notes, ParsedNote{
						Date: currentDate,
						Text: cleanMarkdownEscapes(text),
					})
				}
			}
			currentDate = matches[1]
			currentLines = nil
			// Rest of the line after date
			if len(matches) >= 3 && strings.TrimSpace(matches[2]) != "" {
				currentLines = append(currentLines, strings.TrimSpace(matches[2]))
			}
		} else if currentDate != "" {
			// Continuation lines (including lines starting with \-)
			if trimmed != "" {
				currentLines = append(currentLines, trimmed)
			}
		}
	}

	// Save last note
	if currentDate != "" {
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if text != "" {
			notes = append(notes, ParsedNote{
				Date: currentDate,
				Text: cleanMarkdownEscapes(text),
			})
		}
	}

	return notes, nil
}
