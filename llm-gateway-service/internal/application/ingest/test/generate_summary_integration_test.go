//go:build integration

package ingest_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	"github.com/story-engine/llm-gateway-service/internal/application/ingest"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

func TestGenerateSummary_GeminiIntegration(t *testing.T) {
	if strings.TrimSpace(os.Getenv("LLM_TESTS_ENABLED")) == "" {
		t.Skip("LLM_TESTS_ENABLED not set; skipping LLM integration tests")
	}

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	useCase := ingest.NewGenerateSummaryUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	output, err := useCase.Execute(ctx, ingest.GenerateSummaryInput{
		EntityType: "character",
		Name:       "Aria",
		Contents: []string{
			"Aria is a mage and a member of the Crimson Order. She protected the Obsidian Tower.",
		},
		Context:  "",
		MaxItems: 3,
	})
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}

	if len(output.Summaries) == 0 {
		t.Fatalf("expected summaries, got none")
	}

	fmt.Println("----")
	fmt.Println(output)
	fmt.Println(len(output.Summaries))
	fmt.Println("----")
	for _, summary := range output.Summaries {
		if strings.TrimSpace(summary) == "" {
			t.Fatalf("summary is empty")
		}
	}
}

func loadGeminiAPIKey(t *testing.T) string {
	if value := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); value != "" {
		return value
	}

	candidates := []string{
		"gemini.keys",
		filepath.Join("llm-gateway-service", "gemini.keys"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(trimmed, "=") {
				parts := strings.SplitN(trimmed, "=", 2)
				trimmed = strings.TrimSpace(parts[1])
			}
			if trimmed != "" {
				return trimmed
			}
		}
	}

	t.Log("gemini.keys not found or empty; set GEMINI_API_KEY")
	return ""
}
