package parser

import (
	"fmt"
	"strings"
)

// ParseTabsIndex parses a tabs_index.txt file into a TabsIndex struct.
func ParseTabsIndex(content string) (*TabsIndex, error) {
	result := &TabsIndex{}
	lines := strings.Split(content, "\n")

	inTabs := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Document ID:") {
			result.DocumentID = strings.TrimSpace(strings.TrimPrefix(trimmed, "Document ID:"))
		} else if strings.HasPrefix(trimmed, "Document:") {
			result.DocumentName = strings.TrimSpace(strings.TrimPrefix(trimmed, "Document:"))
		} else if strings.HasPrefix(trimmed, "Exported at:") {
			result.ExportedAt = strings.TrimSpace(strings.TrimPrefix(trimmed, "Exported at:"))
		} else if trimmed == "Tabs:" {
			inTabs = true
		} else if inTabs && strings.HasPrefix(trimmed, "- ") {
			tab := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			result.Tabs = append(result.Tabs, tab)
		}
	}

	if result.DocumentName == "" {
		return nil, fmt.Errorf("tabs_index: missing Document field")
	}

	return result, nil
}
