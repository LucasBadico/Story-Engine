package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterPlotTool registers the get_plots tool on the MCP server.
func RegisterPlotTool(s *server.MCPServer, gw *gateway.Client, defaultStoryID string) {
	tool := mcp.NewTool("get_plots",
		mcp.WithDescription("Get plot threads for a story at a specific point in the narrative. Returns plot summaries with their current state, developments, and connected entities. Plots are versioned — each chapter may show a different state of the plot thread."),
		mcp.WithString("story_id", mcp.Description("Story UUID. Optional if a default is configured.")),
		mcp.WithNumber("as_of_chapter", mcp.Description("Chapter number — returns plot state as known at this chapter. Defaults to latest if omitted.")),
		mcp.WithString("mode", mcp.Description("Retrieval mode: revealed_narrative (default, spoiler-safe), timeline_consistent, or omniscient")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storyID := resolveStoryID(request.Params.Arguments, defaultStoryID)
		if storyID == "" {
			return mcp.NewToolResultError("story_id is required — provide it or set --default-story-id"), nil
		}

		chapter := 9999
		if chapterF, ok := request.Params.Arguments["as_of_chapter"].(float64); ok && chapterF >= 0 {
			chapter = int(chapterF)
		}

		mode, _ := request.Params.Arguments["mode"].(string)

		canon, err := gw.GetCanon(ctx, storyID, chapter, nil, mode, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get canon: %v", err)), nil
		}

		// Filter canon entries for plot and plot_development source types
		var plotEntries []gateway.CanonEntry
		var devEntries []gateway.CanonEntry
		for _, entry := range canon.Canon {
			switch entry.SourceType {
			case "plot":
				plotEntries = append(plotEntries, entry)
			case "plot_development":
				devEntries = append(devEntries, entry)
			}
		}

		if len(plotEntries) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No plot threads found at chapter %d.", chapter)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Plot Threads (as of chapter %d)\n\n", chapter))

		for _, p := range plotEntries {
			sb.WriteString(fmt.Sprintf("## %s\n\n", p.Title))
			sb.WriteString(p.Content)
			sb.WriteString("\n\n")

			// Find developments for this plot (match by source_id in title)
			plotDevs := filterDevsByPlotTitle(devEntries, p.Title)
			if len(plotDevs) > 0 {
				sb.WriteString("### Developments\n\n")
				for _, d := range plotDevs {
					sb.WriteString(fmt.Sprintf("**%s**\n", d.Title))
					sb.WriteString(d.Content)
					sb.WriteString("\n\n")
				}
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	})
}

func filterDevsByPlotTitle(devs []gateway.CanonEntry, plotTitle string) []gateway.CanonEntry {
	// Plot title is "Plot: X", development title is "Plot Development: X — action (Ch N)"
	// Extract the plot name from the plot title
	plotName := strings.TrimPrefix(plotTitle, "Plot: ")

	var matched []gateway.CanonEntry
	for _, d := range devs {
		if strings.Contains(d.Title, plotName) {
			matched = append(matched, d)
		}
	}
	return matched
}
