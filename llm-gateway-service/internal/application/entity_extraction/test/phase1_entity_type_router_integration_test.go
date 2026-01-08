//go:build integration

package entity_extraction_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	entityextraction "github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

func TestPhase1EntityTypeRouter_GeminiIntegration(t *testing.T) {
	if strings.TrimSpace(os.Getenv("LLM_TESTS_ENABLED")) == "" {
		t.Skip("LLM_TESTS_ENABLED not set; skipping LLM integration tests")
	}

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	router := entityextraction.NewPhase1EntityTypeRouterUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)

	testCases := []struct {
		name          string
		text          string
		expectedTypes []string
	}{
		{
			name: "location_and_faction",
			text: "Aria stepped into the Obsidian Tower and met the Crimson Order.",
			expectedTypes: []string{
				"location",
				"faction",
			},
		},
		{
			name: "characters_joao_maria",
			text: "Joao e Maria perdem a hora e quando percebem o sol está se pondo.",
			expectedTypes: []string{
				"character",
			},
		},
		{
			name: "characters_john_snow_helena",
			text: "- You know nothing, john snow.- said Helena.",
			expectedTypes: []string{
				"character",
			},
		},
		{
			name: "tower_of_gods_factions",
			text: "The Tower of Gods is a place. Mortals refuse to enter it, but the demigods do.",
			expectedTypes: []string{
				"location",
			},
		},
		{
			name: "explicit_faction_treaty",
			text: "The Crimson Order and the Silver Guild signed a treaty after years of war.",
			expectedTypes: []string{
				"faction",
			},
		},
	}

	allowed := map[string]struct{}{
		"character": {},
		"location":  {},
		"artefact":  {},
		"faction":   {},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			input := entityextraction.Phase1EntityTypeRouterInput{
				Text:    testCase.text,
				Context: "",
				EntityTypes: []string{
					"character",
					"location",
					"artefact",
					"faction",
				},
				MaxCandidates: 4,
			}

			output, err := router.Execute(ctx, input)
			if err != nil {
				t.Fatalf("router failed: %v", err)
			}

			if len(output.Candidates) == 0 && len(testCase.expectedTypes) > 0 {
				t.Fatalf("expected candidates, got none")
			}

			actual := make(map[string]struct{})
			for _, candidate := range output.Candidates {
				if _, ok := allowed[candidate.Type]; !ok {
					t.Fatalf("unexpected type: %s", candidate.Type)
				}
				if candidate.Confidence < 0 || candidate.Confidence > 1 {
					t.Fatalf("invalid confidence: %v", candidate.Confidence)
				}
				if strings.TrimSpace(candidate.Why) == "" {
					t.Fatalf("missing rationale for type: %s", candidate.Type)
				}
				actual[candidate.Type] = struct{}{}
			}

			for _, expectedType := range testCase.expectedTypes {
				if _, ok := actual[expectedType]; !ok {
					t.Fatalf("expected type %q not returned", expectedType)
				}
			}
		})
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
