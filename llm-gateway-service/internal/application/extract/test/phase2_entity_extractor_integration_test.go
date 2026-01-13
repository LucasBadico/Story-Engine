//go:build integration

package extract_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/entities"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

func TestPhase2CharacterExtractor_GeminiSummary(t *testing.T) {
	skipIfLLMDisabled(t)

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	extractor := entities.NewPhase2CharacterExtractorUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	output, err := extractor.Execute(ctx, entities.Phase2EntityExtractorInput{
		Text:         "Aria stepped into the Obsidian Tower and met the Crimson Order.",
		Context:      "",
		AlreadyFound: nil,
	})
	if err != nil {
		t.Fatalf("extractor failed: %v", err)
	}

	if len(output.Candidates) == 0 {
		t.Fatalf("expected candidates, got none")
	}

	foundSummary := false
	for _, candidate := range output.Candidates {
		if strings.TrimSpace(candidate.Name) == "" {
			t.Fatalf("candidate missing name")
		}
		if strings.TrimSpace(candidate.Evidence) == "" {
			t.Fatalf("candidate missing evidence")
		}
		if strings.TrimSpace(candidate.Summary) != "" {
			foundSummary = true
		}
	}

	if !foundSummary {
		t.Fatalf("expected at least one candidate with summary")
	}
}

func TestPhase2LocationExtractor_GeminiSummary(t *testing.T) {
	skipIfLLMDisabled(t)

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	extractor := entities.NewPhase2LocationExtractorUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	output, err := extractor.Execute(ctx, entities.Phase2EntityExtractorInput{
		Text:         "Aria entered the Obsidian Tower and rested inside.",
		Context:      "",
		AlreadyFound: nil,
	})
	if err != nil {
		t.Fatalf("extractor failed: %v", err)
	}

	if len(output.Candidates) == 0 {
		t.Fatalf("expected candidates, got none")
	}

	hasSummary := false
	for _, candidate := range output.Candidates {
		if strings.TrimSpace(candidate.Name) == "" {
			t.Fatalf("candidate missing name")
		}
		if strings.TrimSpace(candidate.Evidence) == "" {
			t.Fatalf("candidate missing evidence")
		}
		if strings.TrimSpace(candidate.Summary) != "" {
			hasSummary = true
		}
	}
	if !hasSummary {
		t.Fatalf("expected at least one candidate with summary")
	}
}

func TestPhase2FactionExtractor_GeminiSummary(t *testing.T) {
	skipIfLLMDisabled(t)

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	extractor := entities.NewPhase2FactionExtractorUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	output, err := extractor.Execute(ctx, entities.Phase2EntityExtractorInput{
		Text:         "The Crimson Order declared war on the Silver Guild.",
		Context:      "",
		AlreadyFound: nil,
	})
	if err != nil {
		t.Fatalf("extractor failed: %v", err)
	}

	if len(output.Candidates) == 0 {
		t.Fatalf("expected candidates, got none")
	}

	hasSummary := false
	for _, candidate := range output.Candidates {
		if strings.TrimSpace(candidate.Name) == "" {
			t.Fatalf("candidate missing name")
		}
		if strings.TrimSpace(candidate.Evidence) == "" {
			t.Fatalf("candidate missing evidence")
		}
		if strings.TrimSpace(candidate.Summary) != "" {
			hasSummary = true
		}
	}
	if !hasSummary {
		t.Fatalf("expected at least one candidate with summary")
	}
}

func TestPhase2Orchestrator_GeminiIntegration(t *testing.T) {
	skipIfLLMDisabled(t)

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	log := logger.New()
	router := entities.NewPhase1EntityTypeRouterUseCase(
		gemini.NewRouterModel(apiKey, model),
		log,
	)
	orchestrator := entities.NewPhase2EntryUseCase(
		gemini.NewRouterModel(apiKey, model),
		log,
		nil,
	)

	text := strings.Join([]string{
		// Paragraph 1 — small (character + location)
		"Aria stepped through the gates of the Obsidian Tower.",

		// Paragraph 2 — small/medium (artefact)
		"The air grew colder as she tightened her grip on the blackened key hanging from her belt.",

		// Paragraph 3 — medium (faction + location)
		"Inside the Obsidian Tower, Aria encountered members of the Crimson Order, their crimson cloaks marking them as guardians of the place.",

		// Paragraph 4 — medium (artefact reuse + implicit importance)
		"One of the guardians glanced at the blackened key and fell silent, as if the object carried a weight none of them wished to name.",

		// Paragraph 5 — medium (cross-entity interaction + exit)
		"After warning the Crimson Order of the approaching danger, Aria turned away and left the Obsidian Tower behind her.",
	}, "\n\n")

	split, err := entities.SplitTextIntoParagraphChunks(entities.Phase0TextSplitInput{
		Text:          text,
		MaxChunkChars: 200,
		OverlapChars:  20,
	})
	if err != nil {
		t.Fatalf("split failed: %v", err)
	}

	var routedChunks []entities.Phase2RoutedChunk
	for _, paragraph := range split.Paragraphs {
		for _, chunk := range paragraph.Chunks {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			output, err := router.Execute(ctx, entities.Phase1EntityTypeRouterInput{
				Text:    chunk.Text,
				Context: "",
				EntityTypes: []string{
					"character",
					"location",
					"artefact",
					"faction",
				},
				MaxCandidates: 4,
			})
			cancel()
			if err != nil {
				t.Fatalf("router failed: %v", err)
			}
			types := make([]string, 0, len(output.Candidates))
			for _, candidate := range output.Candidates {
				types = append(types, candidate.Type)
			}
			routedChunks = append(routedChunks, entities.Phase2RoutedChunk{
				ParagraphID: paragraph.ParagraphID,
				ChunkID:     chunk.ChunkID,
				StartOffset: chunk.StartOffset,
				EndOffset:   chunk.EndOffset,
				Text:        chunk.Text,
				Types:       types,
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	findings, err := orchestrator.Execute(ctx, entities.Phase2EntryInput{
		Context:               "",
		MaxCandidatesPerChunk: 4,
		Chunks:                routedChunks,
	})
	if err != nil {
		t.Fatalf("orchestrator failed: %v", err)
	}

	hasAria := false
	hasTower := false
	hasOrder := false
	for _, finding := range findings.Findings {
		fmt.Println("-----")
		fmt.Println("found: ")
		fmt.Println(finding)
		fmt.Println("-----")
		nameLower := strings.ToLower(finding.Name)
		switch {
		case strings.Contains(nameLower, "aria"):
			hasAria = true
			if len(finding.Occurrences) < 2 {
				t.Fatalf("expected multiple occurrences for Aria")
			}
		case strings.Contains(nameLower, "obsidian"):
			hasTower = true
		case strings.Contains(nameLower, "crimson"):
			hasOrder = true
		}
	}

	if !hasAria || !hasTower || !hasOrder {
		t.Fatalf("expected Aria, Obsidian Tower, and Crimson Order findings")
	}
}

func skipIfLLMDisabled(t *testing.T) {
	if strings.TrimSpace(os.Getenv("LLM_TESTS_ENABLED")) == "" {
		t.Skip("LLM_TESTS_ENABLED not set; skipping LLM integration tests")
	}
}
