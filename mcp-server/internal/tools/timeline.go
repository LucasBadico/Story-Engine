package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterTimelineTool registers the get_timeline tool on the MCP server.
func RegisterTimelineTool(s *server.MCPServer, gw *gateway.Client, defaultStoryID string) {
	tool := mcp.NewTool("get_timeline",
		mcp.WithDescription("Get the story timeline events up to a specific chapter. Returns a chronological list of narrative events with entities involved."),
		mcp.WithString("story_id", mcp.Description("Story UUID. Optional if a default is configured.")),
		mcp.WithNumber("as_of_chapter", mcp.Required(), mcp.Description("Chapter number — timeline includes events up to this chapter")),
		mcp.WithNumber("as_of_day", mcp.Description("In-story day number for timeline-consistent mode")),
		mcp.WithString("mode", mcp.Description("Retrieval mode: revealed_narrative (default), timeline_consistent, or omniscient")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of events to return")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storyID := resolveStoryID(request.Params.Arguments, defaultStoryID)
		if storyID == "" {
			return mcp.NewToolResultError("story_id is required — provide it or set --default-story-id"), nil
		}

		chapterF, ok := request.Params.Arguments["as_of_chapter"].(float64)
		if !ok || chapterF < 0 {
			return mcp.NewToolResultError("as_of_chapter is required and must be >= 0"), nil
		}
		chapter := int(chapterF)

		asOfDay := optionalInt(request.Params.Arguments, "as_of_day")
		mode, _ := request.Params.Arguments["mode"].(string)
		limit := optionalInt(request.Params.Arguments, "limit")

		canon, err := gw.GetCanon(ctx, storyID, chapter, asOfDay, mode, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get timeline: %v", err)), nil
		}

		if len(canon.Events) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No timeline events found up to chapter %d.", chapter)), nil
		}

		var sb fmt.Stringer = buildTimeline(canon, chapter)
		return mcp.NewToolResultText(sb.(interface{ String() string }).String()), nil
	})
}

type timelineBuilder struct {
	s string
}

func (t *timelineBuilder) String() string { return t.s }

func buildTimeline(canon *gateway.CanonResponse, chapter int) *timelineBuilder {
	var sb string
	sb = fmt.Sprintf("# Timeline (up to chapter %d)\n\n", chapter)
	for _, ev := range canon.Events {
		sb += formatEvent(&ev)
	}
	return &timelineBuilder{s: sb}
}
