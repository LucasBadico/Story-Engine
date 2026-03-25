package parser

import (
	"regexp"
	"strings"
)

// splitBySections splits content by a header prefix (e.g. "## " or "### ").
// Returns a map of header text -> body text. Headers are trimmed of the prefix.
func splitBySections(content, prefix string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentHeader string
	var currentBody []string

	for _, line := range lines {
		if strings.HasPrefix(line, prefix) && !strings.HasPrefix(line, prefix+"#") {
			// Save previous section
			if currentHeader != "" {
				sections[currentHeader] = strings.TrimSpace(strings.Join(currentBody, "\n"))
			}
			currentHeader = strings.TrimSpace(strings.TrimPrefix(line, prefix))
			currentBody = nil
		} else if currentHeader != "" {
			currentBody = append(currentBody, line)
		}
	}
	// Save last section
	if currentHeader != "" {
		sections[currentHeader] = strings.TrimSpace(strings.Join(currentBody, "\n"))
	}
	return sections
}

// splitBySectionsOrdered splits content by a header prefix, preserving order.
// Returns slices of header and body pairs.
func splitBySectionsOrdered(content, prefix string) (headers []string, bodies []string) {
	lines := strings.Split(content, "\n")

	var currentHeader string
	var currentBody []string

	for _, line := range lines {
		if strings.HasPrefix(line, prefix) && !strings.HasPrefix(line, prefix+"#") {
			if currentHeader != "" {
				headers = append(headers, currentHeader)
				bodies = append(bodies, strings.TrimSpace(strings.Join(currentBody, "\n")))
			}
			currentHeader = strings.TrimSpace(strings.TrimPrefix(line, prefix))
			currentBody = nil
		} else if currentHeader != "" {
			currentBody = append(currentBody, line)
		}
	}
	if currentHeader != "" {
		headers = append(headers, currentHeader)
		bodies = append(bodies, strings.TrimSpace(strings.Join(currentBody, "\n")))
	}
	return
}

// extractBulletList extracts items from lines starting with "* ".
func extractBulletList(content string) []string {
	var items []string
	lines := strings.Split(content, "\n")
	var currentItem []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "* ") {
			if len(currentItem) > 0 {
				items = append(items, cleanMarkdownEscapes(strings.TrimSpace(strings.Join(currentItem, " "))))
			}
			currentItem = []string{strings.TrimPrefix(trimmed, "* ")}
		} else if len(currentItem) > 0 && trimmed != "" {
			// Continuation line for indented sub-items
			currentItem = append(currentItem, trimmed)
		}
	}
	if len(currentItem) > 0 {
		items = append(items, cleanMarkdownEscapes(strings.TrimSpace(strings.Join(currentItem, " "))))
	}
	return items
}

// extractNumberedList extracts items from numbered lists (1. , 2. , etc.)
func extractNumberedList(content string) []string {
	var items []string
	lines := strings.Split(content, "\n")
	re := regexp.MustCompile(`^\d+\.\s+`)
	var currentItem []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if re.MatchString(trimmed) {
			if len(currentItem) > 0 {
				items = append(items, cleanMarkdownEscapes(strings.TrimSpace(strings.Join(currentItem, " "))))
			}
			currentItem = []string{re.ReplaceAllString(trimmed, "")}
		} else if len(currentItem) > 0 && trimmed != "" {
			currentItem = append(currentItem, trimmed)
		}
	}
	if len(currentItem) > 0 {
		items = append(items, cleanMarkdownEscapes(strings.TrimSpace(strings.Join(currentItem, " "))))
	}
	return items
}

// extractKeyValuePairs parses lines with **Key**: value pattern.
func extractKeyValuePairs(content string) map[string]string {
	pairs := make(map[string]string)
	re := regexp.MustCompile(`\*\*([^*]+)\*\*:\s*(.*)`)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			pairs[key] = cleanMarkdownEscapes(value)
		}
	}
	return pairs
}

// trimMarkdownComments removes <!-- ... --> comment blocks from content.
func trimMarkdownComments(content string) string {
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return re.ReplaceAllString(content, "")
}

// cleanMarkdownEscapes removes backslash escapes from markdown text.
// For example: \- → -, \= → =, \[ → [, \] → ]
func cleanMarkdownEscapes(text string) string {
	re := regexp.MustCompile(`\\([-=\[\]()~>#+.*!|{}])`)
	return re.ReplaceAllString(text, "$1")
}

// getSectionBody extracts the body text for a given ## section from content.
// Returns empty string if section is not found.
func getSectionBody(content, sectionName string) string {
	sections := splitBySections(content, "## ")
	for key, body := range sections {
		if strings.EqualFold(strings.TrimSpace(cleanMarkdownEscapes(key)), sectionName) {
			return body
		}
	}
	return ""
}
