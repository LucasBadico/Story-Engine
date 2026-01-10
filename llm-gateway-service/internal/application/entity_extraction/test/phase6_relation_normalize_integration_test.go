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

func TestPhase6RelationNormalize_WithPhase5GeminiIntegration(t *testing.T) {
	if strings.TrimSpace(os.Getenv("LLM_TESTS_ENABLED")) == "" {
		t.Skip("LLM_TESTS_ENABLED not set; skipping LLM integration tests")
	}

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	phase5 := entityextraction.NewPhase5RelationDiscoveryUseCase(
		gemini.NewRouterModel(apiKey, model),
		logger.New(),
	)
	phase6 := entityextraction.NewPhase6RelationNormalizeUseCase(logger.New())
	phase6.SetSummaryModel(gemini.NewRouterModel(apiKey, model))

	inputText := "Ari swore loyalty to the Order of the Sun before entering the tower."

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	phase5Output, err := phase5.Execute(ctx, entityextraction.Phase5RelationDiscoveryInput{
		RequestID: "req-rel-6",
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
	if len(phase5Output.Relations) == 0 {
		t.Fatalf("expected relations from phase5, got none")
	}

	phase6Output, err := phase6.Execute(ctx, entityextraction.Phase6RelationNormalizeInput{
		RequestID: "req-rel-6",
		Context: entityextraction.Phase5Context{
			Type: "scene",
			ID:   "scene-uuid",
		},
		Relations: phase5Output.Relations,
		RefMap: map[string]entityextraction.Phase6ResolvedRef{
			"finding:character:0": {ID: "uuid-ari", Type: "character", Name: "Ari"},
			"finding:faction:1":   {ID: "uuid-order", Type: "faction", Name: "Order of the Sun"},
		},
		SuggestedRelationsBySourceType: map[string]entityextraction.Phase5PerEntityRelationMap{
			"character": {
				EntityType: "character",
				Version:    1,
				Relations: map[string]entityextraction.Phase5RelationConstraintSpec{
					"member_of": {
						PairCandidates: []string{"faction"},
					},
				},
			},
		},
		RelationTypes: map[string]entityextraction.Phase6RelationTypeDefinition{
			"member_of": {Mirror: "has_member", PreferredDirection: "source_to_target", Semantics: "Source belongs to a group."},
		},
	})
	if err != nil {
		t.Fatalf("phase6 normalize failed: %v", err)
	}

	fmt.Println("================")
	fmt.Println(phase6Output.Relations)
	fmt.Println("================")

	if len(phase6Output.Relations) == 0 {
		t.Fatalf("expected normalized relations, got none")
	}
	if phase6Output.Relations[0].Summary == "" {
		t.Fatalf("expected summary")
	}
}
