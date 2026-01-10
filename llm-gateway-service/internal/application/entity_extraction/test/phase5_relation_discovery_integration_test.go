//go:build integration

package entity_extraction_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	entityextraction "github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

func TestPhase5RelationDiscovery_GeminiIntegration(t *testing.T) {
	if strings.TrimSpace(os.Getenv("LLM_TESTS_ENABLED")) == "" {
		t.Skip("LLM_TESTS_ENABLED not set; skipping LLM integration tests")
	}

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	useCase := entityextraction.NewPhase5RelationDiscoveryUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)

	inputText := "Ari swore loyalty to the Order of the Sun before entering the tower."

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	output, err := useCase.Execute(ctx, entityextraction.Phase5RelationDiscoveryInput{
		RequestID: "req-rel-1",
		Context: entityextraction.Phase5Context{
			Type: "scene",
			ID:   "scene-uuid",
		},
		Text: entityextraction.Phase5TextSpec{
			Mode:          "spans",
			GlobalSummary: []string{"Ari declares loyalty to the Order of the Sun."},
			Spans: []entityextraction.Phase5Span{
				{
					SpanID: "span:1",
					Start:  0,
					End:    len(inputText),
					Text:   inputText,
				},
			},
		},
		EntityFindings: []entityextraction.Phase5EntityFinding{
			{
				Ref:      "finding:character:0",
				Type:     "character",
				Name:     "Ari",
				Summary:  "Young mage apprentice.",
				Mentions: []string{"span:1"},
			},
			{
				Ref:      "finding:faction:1",
				Type:     "faction",
				Name:     "Order of the Sun",
				Summary:  "A militant religious order.",
				Mentions: []string{"span:1"},
			},
		},
		SuggestedRelationsBySourceType: map[string]entityextraction.Phase5PerEntityRelationMap{
			"character": {
				EntityType: "character",
				Version:    1,
				Relations: map[string]entityextraction.Phase5RelationConstraintSpec{
					"member_of": {
						PairCandidates: []string{"faction"},
						Description:    "Character belongs to a group or organization.",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("phase5 discovery failed: %v", err)
	}

	if len(output.Relations) == 0 {
		t.Fatalf("expected relations, got none")
	}

	fmt.Println("===============")
	fmt.Println(output.Relations)
	fmt.Println("===============")
	found := false
	for _, rel := range output.Relations {
		if rel.RelationType == "member_of" && rel.Source.Type == "character" && rel.Target.Type == "faction" {
			found = true
			if rel.Evidence.SpanID != "span:1" {
				t.Fatalf("expected span:1 evidence, got %q", rel.Evidence.SpanID)
			}
			if strings.TrimSpace(rel.Evidence.Quote) == "" {
				t.Fatalf("expected evidence quote")
			}
			if rel.Confidence <= 0 {
				t.Fatalf("expected confidence > 0")
			}
		}
	}

	if !found {
		t.Fatalf("expected member_of relation")
	}
}
