package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
)

// RegisterStoryResources registers story:// URI resources on the MCP server.
func RegisterStoryResources(s *server.MCPServer, gw *gateway.Client, defaultStoryID string) {
	if defaultStoryID == "" {
		return
	}

	// story://stories/{id}/identity — canon cold layer (premise, themes)
	identityURI := fmt.Sprintf("story://stories/%s/identity", defaultStoryID)
	s.AddResource(mcp.NewResource(
		identityURI,
		"Story Identity",
		mcp.WithResourceDescription("The story's canonical identity — premise, themes, and core world knowledge from the cold layer."),
		mcp.WithMIMEType("text/plain"),
	), func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		canon, err := gw.GetCanon(ctx, defaultStoryID, 999999, nil, "omniscient", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch canon: %w", err)
		}

		var sb strings.Builder
		sb.WriteString("# Story Canon\n\n")
		for _, entry := range canon.Canon {
			sb.WriteString(fmt.Sprintf("## %s\n", entry.Title))
			sb.WriteString(entry.Content)
			sb.WriteString("\n\n")
		}

		if len(canon.Canon) == 0 {
			sb.WriteString("No canon entries found.\n")
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     sb.String(),
			},
		}, nil
	})

	// story://stories/{id}/world-rules — world rules from canon
	rulesURI := fmt.Sprintf("story://stories/%s/world-rules", defaultStoryID)
	s.AddResource(mcp.NewResource(
		rulesURI,
		"World Rules",
		mcp.WithResourceDescription("The story's world rules and constraints."),
		mcp.WithMIMEType("text/plain"),
	), func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		canon, err := gw.GetCanon(ctx, defaultStoryID, 999999, nil, "omniscient", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch world rules: %w", err)
		}

		var sb strings.Builder
		sb.WriteString("# World Rules\n\n")
		found := false
		for _, entry := range canon.Canon {
			if strings.Contains(strings.ToLower(entry.SourceType), "lore") ||
				strings.Contains(strings.ToLower(entry.Title), "rule") ||
				strings.Contains(strings.ToLower(entry.Title), "world") {
				sb.WriteString(fmt.Sprintf("## %s\n", entry.Title))
				sb.WriteString(entry.Content)
				sb.WriteString("\n\n")
				found = true
			}
		}
		if !found {
			sb.WriteString("No world rules defined yet.\n")
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     sb.String(),
			},
		}, nil
	})
}
