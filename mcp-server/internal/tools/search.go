package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterSearchTool registers the search_story_memory tool on the MCP server.
func RegisterSearchTool(s *server.MCPServer, gw *gateway.Client, defaultStoryID, defaultTenantID string) {
	tool := mcp.NewTool("search_story_memory",
		mcp.WithDescription("Search story memory using semantic similarity. Returns relevant story content chunks matching the query."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The search query — what you want to find in the story memory")),
		mcp.WithString("story_id", mcp.Description("Story UUID. Optional if a default is configured.")),
		mcp.WithNumber("as_of_chapter", mcp.Description("Chapter number to scope the search to")),
		mcp.WithNumber("as_of_day", mcp.Description("In-story day number to scope the search to")),
		mcp.WithString("mode", mcp.Description("Retrieval mode: 'full', 'warm', or 'hot'. Default: 'full'")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results to return. Default: 10")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, _ := request.Params.Arguments["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		tenantID := defaultTenantID
		if tenantID == "" {
			return mcp.NewToolResultError("no tenant ID configured — set --default-tenant-id"), nil
		}

		limit := 10
		if l, ok := request.Params.Arguments["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}

		result, err := gw.Search(ctx, tenantID, query, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		return mcp.NewToolResultText(formatSearchResults(result)), nil
	})
}

func formatSearchResults(resp *gateway.SearchResponse) string {
	if len(resp.Chunks) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results:\n\n", len(resp.Chunks)))

	for i, chunk := range resp.Chunks {
		sb.WriteString(fmt.Sprintf("--- Result %d (score: %.2f) ---\n", i+1, chunk.Score))

		if chunk.ContentKind != nil {
			sb.WriteString(fmt.Sprintf("Type: %s\n", *chunk.ContentKind))
		}
		if chunk.POVCharacter != nil {
			sb.WriteString(fmt.Sprintf("POV: %s\n", *chunk.POVCharacter))
		}
		if len(chunk.Characters) > 0 {
			sb.WriteString(fmt.Sprintf("Characters: %s\n", strings.Join(chunk.Characters, ", ")))
		}
		if chunk.LocationName != nil {
			sb.WriteString(fmt.Sprintf("Location: %s\n", *chunk.LocationName))
		}
		if chunk.Timeline != nil {
			sb.WriteString(fmt.Sprintf("Timeline: %s\n", *chunk.Timeline))
		}
		if chunk.BeatType != nil {
			sb.WriteString(fmt.Sprintf("Beat: %s\n", *chunk.BeatType))
		}

		sb.WriteString(fmt.Sprintf("\n%s\n\n", chunk.Content))
	}

	return sb.String()
}
