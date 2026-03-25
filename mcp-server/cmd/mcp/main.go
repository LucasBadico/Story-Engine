package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/story-engine/mcp-server/internal/gateway"
	"github.com/story-engine/mcp-server/internal/resources"
	"github.com/story-engine/mcp-server/internal/tools"
)

func main() {
	gatewayURL := flag.String("gateway-url", "http://localhost:8081", "LLM Gateway base URL")
	defaultStoryID := flag.String("default-story-id", "", "Default story UUID (optional, so the LLM doesn't have to pass it every time)")
	defaultTenantID := flag.String("default-tenant-id", "", "Default tenant UUID (required for search)")
	sseAddr := flag.String("sse", "", "Start SSE transport on this address (e.g. :8082) instead of stdio")
	flag.Parse()

	// Allow env vars as fallback.
	if *gatewayURL == "http://localhost:8081" {
		if env := os.Getenv("GATEWAY_URL"); env != "" {
			*gatewayURL = env
		}
	}
	if *defaultStoryID == "" {
		if env := os.Getenv("DEFAULT_STORY_ID"); env != "" {
			*defaultStoryID = env
		}
	}
	if *defaultTenantID == "" {
		if env := os.Getenv("DEFAULT_TENANT_ID"); env != "" {
			*defaultTenantID = env
		}
	}

	gw := gateway.NewClient(*gatewayURL)

	mcpServer := server.NewMCPServer(
		"story-engine-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Register tools.
	tools.RegisterSearchTool(mcpServer, gw, *defaultStoryID, *defaultTenantID)
	tools.RegisterCanonTool(mcpServer, gw, *defaultStoryID)
	tools.RegisterCharacterTool(mcpServer, gw, *defaultStoryID)
	tools.RegisterTimelineTool(mcpServer, gw, *defaultStoryID)
	tools.RegisterWritingContextTool(mcpServer, gw, *defaultStoryID, *defaultTenantID)
	tools.RegisterPlotTool(mcpServer, gw, *defaultStoryID)

	// Register resources.
	resources.RegisterStoryResources(mcpServer, gw, *defaultStoryID)

	if *sseAddr != "" {
		// Start SSE transport (for Windows LM Studio connecting to WSL2).
		sseServer := server.NewSSEServer(mcpServer)
		fmt.Fprintf(os.Stderr, "MCP SSE server listening on %s\n", *sseAddr)
		if err := sseServer.Start(*sseAddr); err != nil {
			fmt.Fprintf(os.Stderr, "MCP SSE server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Start stdio transport.
		stdio := server.NewStdioServer(mcpServer)
		if err := stdio.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	}
}
