package parser

import (
	"strings"
)

// ParseContent parses a 03_content.md file into a slice of ParsedContentScene.
func ParseContent(content string) ([]ParsedContentScene, error) {
	content = trimMarkdownComments(content)
	headers, bodies := splitBySectionsOrdered(content, "### ")

	var scenes []ParsedContentScene
	for i, header := range headers {
		// Clean bold markers from headers like "**Sunday late evening \- Rouge Seasons Hotel**"
		cleanHeader := strings.TrimSpace(header)
		cleanHeader = strings.TrimPrefix(cleanHeader, "**")
		cleanHeader = strings.TrimSuffix(cleanHeader, "**")
		cleanHeader = cleanMarkdownEscapes(strings.TrimSpace(cleanHeader))

		scenes = append(scenes, ParsedContentScene{
			Header:  cleanHeader,
			Content: cleanMarkdownEscapes(strings.TrimSpace(bodies[i])),
		})
	}

	return scenes, nil
}
