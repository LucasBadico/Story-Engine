package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseGranularCanonFile routes a granular story_canon/ file to the appropriate parser
// based on filename prefix. The filename should be just the base name (e.g., "04_character_akiba_haruka.md").
//
// Returns:
//   - entityType: "character", "location", "artifact", "relationship", "premise", "themes", "world_rules"
//   - parsed: the parsed result (ParsedCharacter, ParsedLocation, etc.)
//   - err: error if the file type is unrecognized
func ParseGranularCanonFile(filename, content string) (entityType string, parsed interface{}, err error) {
	// Strip numeric prefix like "04_" and extension ".md"
	base := filename
	base = strings.TrimSuffix(base, ".md")

	// Remove leading numeric prefix (e.g., "04_")
	numPrefixRe := regexp.MustCompile(`^\d+_`)
	base = numPrefixRe.ReplaceAllString(base, "")

	content = trimMarkdownComments(content)

	switch {
	case strings.HasPrefix(base, "character_"):
		return "character", parseGranularCharacter(content), nil

	case strings.HasPrefix(base, "location_"):
		return "location", parseGranularLocation(content), nil

	case strings.HasPrefix(base, "artifact_"):
		return "artifact", parseGranularArtifact(content), nil

	case strings.HasPrefix(base, "relationship_"):
		return "relationship", parseGranularRelationship(content), nil

	case base == "premise":
		return "premise", parseGranularPremise(content), nil

	case base == "themes":
		return "themes", parseGranularThemes(content), nil

	case base == "world_rules":
		return "world_rules", parseGranularWorldRules(content), nil

	default:
		return "", nil, fmt.Errorf("unrecognized granular canon file type: %s", filename)
	}
}

// parseGranularCharacter parses a granular character file.
// The file has ### header with character info and body with description.
func parseGranularCharacter(content string) ParsedCharacter {
	headers, bodies := splitBySectionsOrdered(content, "### ")
	if len(headers) == 0 {
		return ParsedCharacter{Description: cleanMarkdownEscapes(strings.TrimSpace(content))}
	}

	ch := parseCharacterHeader(headers[0])
	ch.Description = cleanMarkdownEscapes(strings.TrimSpace(bodies[0]))
	return ch
}

// parseGranularLocation parses a granular location file.
func parseGranularLocation(content string) ParsedLocation {
	headers, bodies := splitBySectionsOrdered(content, "### ")
	if len(headers) == 0 {
		return ParsedLocation{Description: cleanMarkdownEscapes(strings.TrimSpace(content))}
	}

	return ParsedLocation{
		Name:        cleanMarkdownEscapes(strings.TrimSpace(headers[0])),
		Description: cleanMarkdownEscapes(strings.TrimSpace(bodies[0])),
	}
}

// parseGranularArtifact parses a granular artifact file.
func parseGranularArtifact(content string) ParsedArtifact {
	headers, bodies := splitBySectionsOrdered(content, "### ")
	if len(headers) == 0 {
		return ParsedArtifact{Description: cleanMarkdownEscapes(strings.TrimSpace(content))}
	}

	return ParsedArtifact{
		Name:        cleanMarkdownEscapes(strings.TrimSpace(headers[0])),
		Description: cleanMarkdownEscapes(strings.TrimSpace(bodies[0])),
	}
}

// parseGranularRelationship parses a granular relationship file.
func parseGranularRelationship(content string) ParsedRelationship {
	headers, bodies := splitBySectionsOrdered(content, "### ")
	if len(headers) == 0 {
		return ParsedRelationship{Description: cleanMarkdownEscapes(strings.TrimSpace(content))}
	}

	rel := parseRelationshipHeader(headers[0])
	rel.Description = cleanMarkdownEscapes(strings.TrimSpace(bodies[0]))
	return rel
}

// parseGranularPremise extracts the premise text from a granular premise file.
func parseGranularPremise(content string) string {
	// Strip all # headers and return the body text
	lines := strings.Split(content, "\n")
	var bodyLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		bodyLines = append(bodyLines, line)
	}
	return cleanMarkdownEscapes(strings.TrimSpace(strings.Join(bodyLines, "\n")))
}

// parseGranularThemes extracts themes as bullet list from a granular themes file.
func parseGranularThemes(content string) []string {
	return extractBulletList(content)
}

// parseGranularWorldRules extracts world rules as numbered list from a granular world_rules file.
func parseGranularWorldRules(content string) []string {
	return extractNumberedList(content)
}
