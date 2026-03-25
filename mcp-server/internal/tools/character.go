package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterCharacterTool registers the get_character_state tool on the MCP server.
func RegisterCharacterTool(s *server.MCPServer, gw *gateway.Client, defaultStoryID string) {
	tool := mcp.NewTool("get_character_state",
		mcp.WithDescription("Get a specific character's state at a point in the narrative. Fetches canon and filters state snapshots by character entity type. Returns emotional state, goals, secrets, clothes, physical state."),
		mcp.WithString("story_id", mcp.Description("Story UUID. Optional if a default is configured.")),
		mcp.WithString("character_name", mcp.Required(), mcp.Description("Name of the character to look up (matched against state data)")),
		mcp.WithNumber("as_of_chapter", mcp.Required(), mcp.Description("Chapter number — state is computed up to this chapter")),
		mcp.WithNumber("as_of_day", mcp.Description("In-story day number for timeline-consistent mode")),
		mcp.WithString("mode", mcp.Description("Retrieval mode: revealed_narrative (default), timeline_consistent, or omniscient")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storyID := resolveStoryID(request.Params.Arguments, defaultStoryID)
		if storyID == "" {
			return mcp.NewToolResultError("story_id is required — provide it or set --default-story-id"), nil
		}

		characterName, _ := request.Params.Arguments["character_name"].(string)
		if characterName == "" {
			return mcp.NewToolResultError("character_name is required"), nil
		}

		chapterF, ok := request.Params.Arguments["as_of_chapter"].(float64)
		if !ok || chapterF < 0 {
			return mcp.NewToolResultError("as_of_chapter is required and must be >= 0"), nil
		}
		chapter := int(chapterF)

		asOfDay := optionalInt(request.Params.Arguments, "as_of_day")
		mode, _ := request.Params.Arguments["mode"].(string)

		canon, err := gw.GetCanon(ctx, storyID, chapter, asOfDay, mode, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get canon: %v", err)), nil
		}

		// Filter states by entity_type="character"
		var matched []gateway.StateEntry
		for _, s := range canon.States {
			if s.EntityType != "character" {
				continue
			}
			matched = append(matched, s)
		}

		if len(matched) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No character states found at chapter %d. The character may not have state data ingested yet.", chapter)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Character States (as of chapter %d)\n\n", chapter))
		for _, s := range matched {
			sb.WriteString(fmt.Sprintf("### Character %s (ch %d)\n", s.EntityID, s.ChapterNumber))
			for k, v := range s.StateData {
				sb.WriteString(fmt.Sprintf("- **%s:** %v\n", k, v))
			}
			if s.TimelineDay != nil {
				sb.WriteString(fmt.Sprintf("- **timeline_day:** %d\n", *s.TimelineDay))
			}
			sb.WriteString("\n")
		}

		return mcp.NewToolResultText(sb.String()), nil
	})
}
