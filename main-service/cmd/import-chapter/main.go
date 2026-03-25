package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/importer"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func main() {
	// Define flags
	folder := flag.String("f", "", "path to the export folder (required)")
	tenantID := flag.String("t", "", "tenant UUID (required)")
	worldID := flag.String("w", "", "world UUID (required)")
	storyID := flag.String("s", "", "story UUID (required)")
	chapterNumber := flag.Int("n", -1, "chapter number (required)")
	chapterID := flag.String("chapter-id", "", "existing chapter UUID (optional)")
	apiURL := flag.String("u", "", "main service API URL (default: MAIN_SERVICE_URL env or http://localhost:7979)")
	gatewayURL := flag.String("gateway-url", "", "LLM Gateway URL for structure inference (optional, default from LLM_GATEWAY_URL env)")

	// Long flag aliases
	flag.StringVar(folder, "folder", "", "path to the export folder (required)")
	flag.StringVar(tenantID, "tenant-id", "", "tenant UUID (required)")
	flag.StringVar(worldID, "world-id", "", "world UUID (required)")
	flag.StringVar(storyID, "story-id", "", "story UUID (required)")
	flag.IntVar(chapterNumber, "chapter-number", -1, "chapter number (required)")
	flag.StringVar(apiURL, "api-url", "", "main service API URL")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: import-chapter [flags]\n\nImport a Google Drive exported chapter folder into Story Engine.\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n  go run ./cmd/import-chapter \\\n    -f ../data/gcloud-exported-chapter \\\n    -t \"tenant-uuid\" \\\n    -w \"world-uuid\" \\\n    -s \"story-uuid\" \\\n    -n 0 \\\n    -u http://localhost:7979\n")
	}

	flag.Parse()

	// Validate required flags
	if *folder == "" || *tenantID == "" || *worldID == "" || *storyID == "" || *chapterNumber < 0 {
		fmt.Fprintln(os.Stderr, "Error: all required flags must be provided (-f, -t, -w, -s, -n)")
		flag.Usage()
		os.Exit(1)
	}

	// Parse UUIDs
	parsedTenantID, err := uuid.Parse(*tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid tenant UUID: %v\n", err)
		os.Exit(1)
	}
	parsedWorldID, err := uuid.Parse(*worldID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid world UUID: %v\n", err)
		os.Exit(1)
	}
	parsedStoryID, err := uuid.Parse(*storyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid story UUID: %v\n", err)
		os.Exit(1)
	}

	var parsedChapterID *uuid.UUID
	if *chapterID != "" {
		id, err := uuid.Parse(*chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid chapter UUID: %v\n", err)
			os.Exit(1)
		}
		parsedChapterID = &id
	}

	// Resolve API URL
	resolvedAPIURL := *apiURL
	if resolvedAPIURL == "" {
		resolvedAPIURL = os.Getenv("MAIN_SERVICE_URL")
	}
	if resolvedAPIURL == "" {
		resolvedAPIURL = "http://localhost:7979"
	}

	// Setup logger
	log := logger.New()

	// Create the HTTP API client
	apiClient := importer.NewHTTPAPIClient(resolvedAPIURL, parsedTenantID)
	log.Info("Using main service API", "url", resolvedAPIURL)

	// Create the import use case
	importUseCase := importer.NewImportChapterUseCase(apiClient, log)

	// Wire optional gateway client for chapter structure inference
	resolvedGatewayURL := *gatewayURL
	if resolvedGatewayURL == "" {
		resolvedGatewayURL = os.Getenv("LLM_GATEWAY_URL")
	}
	if resolvedGatewayURL != "" {
		importUseCase.SetGatewayClient(importer.NewGatewayClient(resolvedGatewayURL))
		log.Info("LLM Gateway enabled for chapter structure inference", "url", resolvedGatewayURL)
	}

	// Build input
	input := importer.ImportChapterInput{
		FolderPath:    *folder,
		TenantID:      parsedTenantID,
		WorldID:       parsedWorldID,
		StoryID:       parsedStoryID,
		ChapterNumber: *chapterNumber,
		ChapterID:     parsedChapterID,
	}

	// Execute import
	log.Info("Starting chapter import",
		"folder", *folder,
		"tenant_id", parsedTenantID,
		"world_id", parsedWorldID,
		"story_id", parsedStoryID,
		"chapter_number", *chapterNumber,
	)

	report, err := importUseCase.Execute(context.Background(), input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: import failed: %v\n", err)
		os.Exit(1)
	}

	// Print report
	printReport(report)
}

func printReport(report *importer.ImportReport) {
	createdCounts := countByType(report.Created)

	fmt.Println()
	fmt.Println("Import Report")
	fmt.Println("=============")

	fmt.Printf("Created: %s\n", formatCounts(createdCounts,
		"chapter", "scene", "beat", "content_block",
		"character", "location", "artifact",
		"character_state", "relationship_state",
		"timeline_event", "plot",
	))
	fmt.Printf("Skipped: %d\n", len(report.Skipped))
	fmt.Printf("Failed:  %d\n", len(report.Failed))

	if len(report.Created) > 0 {
		fmt.Println()
		fmt.Println("Created entities:")
		for _, e := range report.Created {
			fmt.Printf("  [%s] %s (id: %s)\n", e.Type, e.Name, e.ID)
		}
	}

	if len(report.Skipped) > 0 {
		fmt.Println()
		fmt.Println("Skipped entities (already exist):")
		for _, e := range report.Skipped {
			fmt.Printf("  [%s] %s (id: %s)\n", e.Type, e.Name, e.ID)
		}
	}

	if len(report.Failed) > 0 {
		fmt.Println()
		fmt.Println("Failed entities:")
		for _, e := range report.Failed {
			fmt.Printf("  [%s] %s: %s\n", e.Type, e.Name, e.Reason)
		}
	}

	fmt.Println()
}

func countByType(entities []importer.ImportedEntity) map[string]int {
	counts := make(map[string]int)
	for _, e := range entities {
		counts[e.Type]++
	}
	return counts
}

func formatCounts(counts map[string]int, types ...string) string {
	parts := make([]string, 0, len(types))
	for _, t := range types {
		label := t + "s"
		switch t {
		case "chapter":
			label = "chapters"
		case "scene":
			label = "scenes"
		case "beat":
			label = "beats"
		case "content_block":
			label = "content blocks"
		case "character_state":
			label = "character states"
		case "relationship_state":
			label = "relationship states"
		case "timeline_event":
			label = "timeline events"
		case "plot":
			label = "plots"
		case "character":
			label = "characters"
		case "location":
			label = "locations"
		case "artifact":
			label = "artifacts"
		}
		parts = append(parts, fmt.Sprintf("%d %s", counts[t], label))
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
