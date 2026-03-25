package parser

import (
	"strings"
)

// ParseStoryCanon parses a 01_story_canon.md file into a ParsedCanon struct.
func ParseStoryCanon(content string) (*ParsedCanon, error) {
	content = trimMarkdownComments(content)
	result := &ParsedCanon{}

	// Extract ## sections
	sections := splitBySections(content, "## ")

	for header, body := range sections {
		cleanHeader := strings.TrimSpace(cleanMarkdownEscapes(header))
		switch {
		case strings.EqualFold(cleanHeader, "Premise"):
			result.Premise = cleanMarkdownEscapes(strings.TrimSpace(body))

		case strings.EqualFold(cleanHeader, "Themes"):
			result.Themes = extractBulletList(body)

		case strings.EqualFold(cleanHeader, "World Rules"):
			result.WorldRules = extractNumberedList(body)

		case strings.EqualFold(cleanHeader, "Characters"):
			result.Characters = parseCharacters(body)

		case strings.EqualFold(cleanHeader, "Relationships"):
			result.Relationships = parseRelationships(body)

		case strings.EqualFold(cleanHeader, "Locations"):
			result.Locations = parseLocations(body)

		case strings.EqualFold(cleanHeader, "Artifacts"):
			result.Artifacts = parseArtifacts(body)
		}
	}

	return result, nil
}

// parseCharacters parses ### character sub-headers from the Characters section.
func parseCharacters(body string) []ParsedCharacter {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var characters []ParsedCharacter

	for i, header := range headers {
		ch := parseCharacterHeader(header)
		ch.Description = cleanMarkdownEscapes(strings.TrimSpace(bodies[i]))
		characters = append(characters, ch)
	}
	return characters
}

// parseCharacterHeader parses a character header like:
// "Akiba, Haruka \- protagonist \- 25 years"
func parseCharacterHeader(header string) ParsedCharacter {
	ch := ParsedCharacter{}

	// Try splitting by " \- " first (escaped dash), then " - "
	var parts []string
	if strings.Contains(header, ` \- `) {
		parts = strings.SplitN(header, ` \- `, 3)
	} else if strings.Contains(header, ` - `) {
		parts = strings.SplitN(header, ` - `, 3)
	} else {
		ch.Name = cleanMarkdownEscapes(strings.TrimSpace(header))
		return ch
	}

	ch.Name = cleanMarkdownEscapes(strings.TrimSpace(parts[0]))
	if len(parts) > 1 {
		ch.Role = cleanMarkdownEscapes(strings.TrimSpace(parts[1]))
	}
	if len(parts) > 2 {
		ch.Age = cleanMarkdownEscapes(strings.TrimSpace(parts[2]))
	}
	return ch
}

// parseRelationships parses ### relationship sub-headers with A ↔ B pattern.
func parseRelationships(body string) []ParsedRelationship {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var relationships []ParsedRelationship

	for i, header := range headers {
		rel := parseRelationshipHeader(header)
		rel.Description = cleanMarkdownEscapes(strings.TrimSpace(bodies[i]))
		relationships = append(relationships, rel)
	}
	return relationships
}

// parseRelationshipHeader parses "Haruka ↔ Tetsuo" into character names.
func parseRelationshipHeader(header string) ParsedRelationship {
	rel := ParsedRelationship{}
	parts := strings.SplitN(header, " ↔ ", 2)
	if len(parts) == 2 {
		rel.CharacterA = cleanMarkdownEscapes(strings.TrimSpace(parts[0]))
		rel.CharacterB = cleanMarkdownEscapes(strings.TrimSpace(parts[1]))
	} else {
		rel.CharacterA = cleanMarkdownEscapes(strings.TrimSpace(header))
	}
	return rel
}

// parseLocations parses ### location sub-headers.
func parseLocations(body string) []ParsedLocation {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var locations []ParsedLocation

	for i, header := range headers {
		loc := ParsedLocation{
			Name:        cleanMarkdownEscapes(strings.TrimSpace(header)),
			Description: cleanMarkdownEscapes(strings.TrimSpace(bodies[i])),
		}
		locations = append(locations, loc)
	}
	return locations
}

// parseArtifacts parses ### artifact sub-headers.
func parseArtifacts(body string) []ParsedArtifact {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var artifacts []ParsedArtifact

	for i, header := range headers {
		art := ParsedArtifact{
			Name:        cleanMarkdownEscapes(strings.TrimSpace(header)),
			Description: cleanMarkdownEscapes(strings.TrimSpace(bodies[i])),
		}
		artifacts = append(artifacts, art)
	}
	return artifacts
}
