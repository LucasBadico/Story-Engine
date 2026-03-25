package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseChapterLog parses a 02_chapter_log.md file into a ParsedChapterLog struct.
func ParseChapterLog(content string) (*ParsedChapterLog, error) {
	content = trimMarkdownComments(content)
	result := &ParsedChapterLog{}

	// Extract chapter title from the first # header (not ##)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			candidate := strings.TrimPrefix(trimmed, "# ")
			// Skip generic headers like "Chapter Log"
			if !strings.EqualFold(candidate, "Chapter Log") {
				result.ChapterTitle = cleanMarkdownEscapes(strings.TrimSpace(candidate))
				break
			}
		}
	}

	// Extract ## sections
	sections := splitBySections(content, "## ")

	for header, body := range sections {
		cleanHeader := strings.TrimSpace(cleanMarkdownEscapes(header))
		switch {
		case strings.EqualFold(cleanHeader, "Character States"):
			result.CharacterStates = parseCharacterStates(body)

		case strings.EqualFold(cleanHeader, "Relationship States"):
			result.RelationshipStates = parseRelationshipStates(body)

		case strings.EqualFold(cleanHeader, "Scenes"):
			result.Scenes = parseScenes(body)

		case strings.EqualFold(cleanHeader, "Secrets"):
			result.Secrets = parseSecrets(body)

		case strings.EqualFold(cleanHeader, "Active Conflicts"):
			result.ActiveConflicts = parseConflicts(body)

		case strings.EqualFold(cleanHeader, "Plot Threads"):
			result.PlotThreads = parsePlotThreads(body)

		case strings.EqualFold(cleanHeader, "Timeline"):
			result.Timeline = parseTimeline(body)

		case strings.EqualFold(cleanHeader, "Chapter Locations"):
			result.ChapterLocations = parseLocations(body)

		case strings.EqualFold(cleanHeader, "Chapter Artifacts"):
			result.ChapterArtifacts = extractBulletList(body)

		case strings.EqualFold(cleanHeader, "Arc Context"):
			result.ArcContext = cleanMarkdownEscapes(strings.TrimSpace(body))
		}
	}

	return result, nil
}

// parseCharacterStates parses ### character state sub-headers with **Key**: value pairs.
func parseCharacterStates(body string) []ParsedCharacterState {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var states []ParsedCharacterState

	for i, header := range headers {
		// Header may contain character info like "Akiba, Haruka \- protagonist"
		// We take the full header as the name
		name := cleanMarkdownEscapes(strings.TrimSpace(header))
		fields := extractKeyValuePairs(bodies[i])

		// If no key-value pairs found, check for plain text fields
		if len(fields) == 0 {
			trimmed := strings.TrimSpace(bodies[i])
			if trimmed != "" {
				fields["State"] = cleanMarkdownEscapes(trimmed)
			}
		}

		states = append(states, ParsedCharacterState{
			Name:   name,
			Fields: fields,
		})
	}
	return states
}

// parseRelationshipStates parses ### relationship state sub-headers with A ↔ B pattern.
func parseRelationshipStates(body string) []ParsedRelationshipState {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var states []ParsedRelationshipState

	for i, header := range headers {
		rel := parseRelationshipHeader(header)
		states = append(states, ParsedRelationshipState{
			CharacterA:  rel.CharacterA,
			CharacterB:  rel.CharacterB,
			Description: cleanMarkdownEscapes(strings.TrimSpace(bodies[i])),
		})
	}
	return states
}

// parseScenes parses ### scene sub-headers with "N\. Title" format.
func parseScenes(body string) []ParsedScene {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var scenes []ParsedScene

	// Match patterns like "1\. Title" or "1. Title"
	re := regexp.MustCompile(`^(\d+)\\?\.\s+(.*)$`)

	for i, header := range headers {
		scene := ParsedScene{}
		matches := re.FindStringSubmatch(header)
		if len(matches) == 3 {
			num, err := strconv.Atoi(matches[1])
			if err == nil {
				scene.Number = num
			}
			scene.Title = cleanMarkdownEscapes(strings.TrimSpace(matches[2]))
		} else {
			scene.Title = cleanMarkdownEscapes(strings.TrimSpace(header))
		}
		scene.Description = cleanMarkdownEscapes(strings.TrimSpace(bodies[i]))
		scenes = append(scenes, scene)
	}
	return scenes
}

// parseSecrets parses ### secret sub-headers where header is entity name and body is the secret.
func parseSecrets(body string) []ParsedSecret {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var secrets []ParsedSecret

	for i, header := range headers {
		secrets = append(secrets, ParsedSecret{
			EntityName: cleanMarkdownEscapes(strings.TrimSpace(header)),
			Secret:     cleanMarkdownEscapes(strings.TrimSpace(bodies[i])),
		})
	}
	return secrets
}

// parseConflicts parses ### conflict sub-headers.
func parseConflicts(body string) []ParsedConflict {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var conflicts []ParsedConflict

	for i, header := range headers {
		conflicts = append(conflicts, ParsedConflict{
			Title:       cleanMarkdownEscapes(strings.TrimSpace(header)),
			Description: cleanMarkdownEscapes(strings.TrimSpace(bodies[i])),
		})
	}
	return conflicts
}

// parsePlotThreads parses ### plot thread sub-headers with "Title — type — action" format.
// Body contains description and optional **Entities**: line.
func parsePlotThreads(body string) []ParsedPlotThread {
	headers, bodies := splitBySectionsOrdered(body, "### ")
	var threads []ParsedPlotThread

	for i, header := range headers {
		thread := ParsedPlotThread{}

		// Parse header: "Title — type — action" or "Title - type - action"
		// Try em-dash first, then regular dash
		parts := strings.Split(header, "—")
		if len(parts) < 3 {
			parts = strings.Split(header, " - ")
		}

		if len(parts) >= 3 {
			thread.Title = cleanMarkdownEscapes(strings.TrimSpace(parts[0]))
			thread.Type = strings.ToLower(strings.TrimSpace(cleanMarkdownEscapes(parts[1])))
			thread.Action = strings.ToLower(strings.TrimSpace(cleanMarkdownEscapes(parts[2])))
		} else {
			thread.Title = cleanMarkdownEscapes(strings.TrimSpace(header))
			thread.Type = "secondary"
			thread.Action = "open"
		}

		// Parse body: description + optional **Entities**: line
		bodyText := strings.TrimSpace(bodies[i])
		bodyLines := strings.Split(bodyText, "\n")
		var descLines []string

		for _, line := range bodyLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "**Entities**:") || strings.HasPrefix(trimmed, "**Entities**:") {
				entitiesStr := strings.TrimPrefix(trimmed, "**Entities**:")
				entitiesStr = strings.TrimPrefix(entitiesStr, "**Entities**:")
				thread.Entities = parseEntityLinks(strings.TrimSpace(entitiesStr))
			} else if trimmed != "" {
				descLines = append(descLines, cleanMarkdownEscapes(trimmed))
			}
		}

		thread.Description = strings.Join(descLines, "\n")
		threads = append(threads, thread)
	}

	return threads
}

// parseEntityLinks parses "role: entity_type: name, role: entity_type: name" format.
func parseEntityLinks(text string) []ParsedEntityLink {
	var links []ParsedEntityLink
	parts := strings.Split(text, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Split by ": " — expecting "role: entity_type: name"
		segments := strings.SplitN(part, ":", 3)
		if len(segments) >= 3 {
			links = append(links, ParsedEntityLink{
				Role:       strings.TrimSpace(segments[0]),
				EntityType: strings.TrimSpace(segments[1]),
				Name:       strings.TrimSpace(cleanMarkdownEscapes(segments[2])),
			})
		} else if len(segments) == 2 {
			// "role: name" — guess entity type from role
			role := strings.TrimSpace(segments[0])
			name := strings.TrimSpace(cleanMarkdownEscapes(segments[1]))
			entityType := guessEntityTypeFromRole(role)
			links = append(links, ParsedEntityLink{
				Role:       role,
				EntityType: entityType,
				Name:       name,
			})
		}
	}
	return links
}

func guessEntityTypeFromRole(role string) string {
	switch strings.ToLower(role) {
	case "subject", "catalyst":
		return "character"
	case "location":
		return "location"
	case "artifact":
		return "artifact"
	default:
		return "character"
	}
}

// parseTimeline parses the timeline section, extracting day number from "Day N:" pattern.
func parseTimeline(body string) *ParsedTimeline {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil
	}

	timeline := &ParsedTimeline{
		RawText: cleanMarkdownEscapes(trimmed),
	}

	// Try to extract "Day N:" pattern
	re := regexp.MustCompile(`Day\s+(\d+)`)
	matches := re.FindStringSubmatch(trimmed)
	if len(matches) == 2 {
		num, err := strconv.Atoi(matches[1])
		if err == nil {
			timeline.Day = &num
			timeline.Confidence = "explicit"
		}
	}

	return timeline
}
