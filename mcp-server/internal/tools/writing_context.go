package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterWritingContextTool registers the get_writing_context tool on the MCP server.
func RegisterWritingContextTool(s *server.MCPServer, gw *gateway.Client, defaultStoryID, defaultTenantID string) {
	tool := mcp.NewTool("get_writing_context",
		mcp.WithDescription("Get a comprehensive writing context pack for a specific chapter/scene. Assembles canon documents, character states, recent events, and scene-relevant search results into one coherent block. Use this before writing or continuing a scene."),
		mcp.WithString("story_id", mcp.Description("Story UUID. Optional if a default is configured.")),
		mcp.WithNumber("chapter", mcp.Required(), mcp.Description("Chapter number to get writing context for")),
		mcp.WithString("scene", mcp.Description("Optional scene description or keywords to focus the context on")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storyID := resolveStoryID(request.Params.Arguments, defaultStoryID)
		if storyID == "" {
			return mcp.NewToolResultError("story_id is required — provide it or set --default-story-id"), nil
		}

		chapterF, ok := request.Params.Arguments["chapter"].(float64)
		if !ok || chapterF < 0 {
			return mcp.NewToolResultError("chapter is required and must be >= 0"), nil
		}
		chapter := int(chapterF)

		scene, _ := request.Params.Arguments["scene"].(string)

		// Fetch canon for this chapter
		eventLimit := 20
		canon, err := gw.GetCanon(ctx, storyID, chapter, nil, "revealed_narrative", &eventLimit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get canon: %v", err)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Writing Context — Chapter %d", chapter))
		if scene != "" {
			sb.WriteString(fmt.Sprintf(" — %s", scene))
		}
		sb.WriteString("\n\n")

		// Section 1: Canon (cold + warm)
		if len(canon.Canon) > 0 {
			sb.WriteString("## Story Canon\n\n")
			for _, entry := range canon.Canon {
				sb.WriteString(fmt.Sprintf("**%s:** %s\n\n", entry.Title, truncate(entry.Content, 500)))
			}
		}

		// Section 2: Character + Relationship States
		if len(canon.States) > 0 {
			sb.WriteString("## Current States\n\n")
			for _, s := range canon.States {
				sb.WriteString(fmt.Sprintf("**%s** (%s, ch %d):", s.EntityType, s.EntityID, s.ChapterNumber))
				for k, v := range s.StateData {
					sb.WriteString(fmt.Sprintf(" %s=%v;", k, v))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}

		// Section 3: Recent Events
		if len(canon.Events) > 0 {
			sb.WriteString("## Recent Events\n\n")
			for _, ev := range canon.Events {
				sb.WriteString(formatEvent(&ev))
			}
			sb.WriteString("\n")
		}

		// Section 4: Scene-relevant search (if scene hint + tenant configured)
		if scene != "" && defaultTenantID != "" {
			searchQuery := fmt.Sprintf("chapter %d %s", chapter, scene)
			searchResult, err := gw.Search(ctx, defaultTenantID, searchQuery, 5)
			if err == nil && len(searchResult.Chunks) > 0 {
				sb.WriteString("## Relevant Memory\n\n")
				for _, chunk := range searchResult.Chunks {
					sb.WriteString(fmt.Sprintf("- (%.2f) %s\n", chunk.Score, truncate(chunk.Content, 200)))
				}
				sb.WriteString("\n")
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
