package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterCanonTool registers the get_canon tool on the MCP server.
func RegisterCanonTool(s *server.MCPServer, gw *gateway.Client, defaultStoryID string) {
	tool := mcp.NewTool("get_canon",
		mcp.WithDescription("Get the assembled story canon at a specific point in the narrative. Returns canonical documents (cold+warm), character/relationship states (hot), and timeline events."),
		mcp.WithString("story_id", mcp.Description("Story UUID. Optional if a default is configured.")),
		mcp.WithNumber("as_of_chapter", mcp.Description("Chapter number — canon is assembled up to this chapter. Defaults to latest if omitted.")),
		mcp.WithNumber("as_of_day", mcp.Description("In-story day number for timeline-consistent mode")),
		mcp.WithString("mode", mcp.Description("Retrieval mode: revealed_narrative (default, spoiler-safe), timeline_consistent, or omniscient")),
		mcp.WithNumber("event_limit", mcp.Description("Maximum number of timeline events to include")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storyID := resolveStoryID(request.Params.Arguments, defaultStoryID)
		if storyID == "" {
			return mcp.NewToolResultError("story_id is required — provide it or set --default-story-id"), nil
		}

		chapter := 9999 // default: show all
		if chapterF, ok := request.Params.Arguments["as_of_chapter"].(float64); ok && chapterF >= 0 {
			chapter = int(chapterF)
		}

		asOfDay := optionalInt(request.Params.Arguments, "as_of_day")
		mode, _ := request.Params.Arguments["mode"].(string)
		eventLimit := optionalInt(request.Params.Arguments, "event_limit")

		canon, err := gw.GetCanon(ctx, storyID, chapter, asOfDay, mode, eventLimit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get canon: %v", err)), nil
		}

		return mcp.NewToolResultText(formatCanon(canon)), nil
	})
}

func formatCanon(c *gateway.CanonResponse) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Story Canon (as of chapter %d, mode: %s)\n\n", c.AsOfChapter, c.Mode))

	// Canon documents (cold + warm layers)
	if len(c.Canon) > 0 {
		sb.WriteString("## Canon Documents\n\n")
		for _, entry := range c.Canon {
			sb.WriteString(fmt.Sprintf("### %s\n", entry.Title))
			sb.WriteString(fmt.Sprintf("*Type: %s, Chapter: %d*\n\n", entry.SourceType, entry.Chapter))
			sb.WriteString(entry.Content)
			sb.WriteString("\n\n")
		}
	}

	// State snapshots (hot layer)
	if len(c.States) > 0 {
		sb.WriteString("## Entity States\n\n")
		for _, s := range c.States {
			sb.WriteString(formatState(&s))
		}
	}

	// Timeline events
	if len(c.Events) > 0 {
		sb.WriteString("## Timeline Events\n\n")
		for _, ev := range c.Events {
			sb.WriteString(formatEvent(&ev))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatState(s *gateway.StateEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s (ch %d)", s.EntityType, s.ChapterNumber))
	if s.TimelineDay != nil {
		sb.WriteString(fmt.Sprintf(", day %d", *s.TimelineDay))
	}
	sb.WriteString("\n")

	for k, v := range s.StateData {
		sb.WriteString(fmt.Sprintf("- **%s:** %v\n", k, v))
	}
	sb.WriteString("\n")
	return sb.String()
}

func formatEvent(ev *gateway.EventEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- **Ch%d", ev.ChapterNumber))
	if ev.TimelineDay != nil {
		sb.WriteString(fmt.Sprintf("/Day%d", *ev.TimelineDay))
	}
	sb.WriteString(":** ")
	sb.WriteString(ev.Description)
	if len(ev.EntitiesInvolved) > 0 {
		sb.WriteString(fmt.Sprintf(" [%s]", strings.Join(ev.EntitiesInvolved, ", ")))
	}
	sb.WriteString("\n")
	return sb.String()
}

// resolveStoryID picks story_id from arguments or falls back to default.
func resolveStoryID(args map[string]interface{}, defaultStoryID string) string {
	if id, ok := args["story_id"].(string); ok && id != "" {
		return id
	}
	return defaultStoryID
}

// optionalInt extracts an optional integer from MCP arguments (which come as float64).
func optionalInt(args map[string]interface{}, key string) *int {
	if v, ok := args[key].(float64); ok {
		i := int(v)
		return &i
	}
	return nil
}
