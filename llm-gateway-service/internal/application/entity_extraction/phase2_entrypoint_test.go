package entity_extraction

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"strings"
)

func TestPhase2EntryUseCase_DedupesByName(t *testing.T) {
	logger := logger.New()
	useCase := &Phase2EntryUseCase{
		extractors: map[string]*Phase2EntityExtractorUseCase{},
		logger:     logger,
	}

	store := map[string]map[string]*Phase2EntityFinding{}
	useCase.mergeFinding(context.Background(), store, "character", "Aria", Phase2EntityOccurrence{
		ParagraphID: 0,
		ChunkID:     0,
		StartOffset: 0,
		EndOffset:   4,
		Evidence:    "Aria",
	}, "", "")
	useCase.mergeFinding(context.Background(), store, "character", "aria", Phase2EntityOccurrence{
		ParagraphID: 0,
		ChunkID:     1,
		StartOffset: 10,
		EndOffset:   14,
		Evidence:    "aria",
	}, "", "")

	byName := store["character"]
	if len(byName) != 1 {
		t.Fatalf("expected 1 deduped finding, got %d", len(byName))
	}
	for _, finding := range byName {
		if len(finding.Occurrences) != 2 {
			t.Fatalf("expected 2 occurrences, got %d", len(finding.Occurrences))
		}
	}
}

func TestPhase2EntryUseCase_OrchestratesChunks(t *testing.T) {
	useCase := &Phase2EntryUseCase{
		extractors: map[string]*Phase2EntityExtractorUseCase{
			"character": newMockExtractor("character", [][]phase2EntityCandidateRaw{
				{
					{Name: "Aria", Evidence: "Aria", Summary: "Mage"},
				},
			}),
			"location": newMockExtractor("location", [][]phase2EntityCandidateRaw{
				{
					{Name: "Obsidian Tower", Evidence: "Obsidian Tower", Summary: "Ancient tower"},
				},
			}),
		},
		logger: logger.New(),
	}

	ctx := context.Background()
	output, err := useCase.Execute(ctx, Phase2EntryInput{
		Context:               "World of Test",
		MaxCandidatesPerChunk: 3,
		Chunks: []Phase2RoutedChunk{
			{
				ParagraphID: 0,
				ChunkID:     0,
				StartOffset: 0,
				EndOffset:   40,
				Text:        "Aria entered the Obsidian Tower.",
				Types:       []string{"character", "location"},
			},
		},
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(output.Findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(output.Findings))
	}

	foundAria := false
	foundTower := false
	for _, finding := range output.Findings {
		if finding.Name == "Aria" {
			foundAria = true
			if finding.Summary == "" {
				t.Fatalf("expected summary for Aria")
			}
			if len(finding.Occurrences) != 1 {
				t.Fatalf("expected 1 occurrence for Aria")
			}
		}
		if finding.Name == "Obsidian Tower" {
			foundTower = true
			if finding.Occurrences[0].StartOffset <= 0 {
				t.Fatalf("expected absolute offset for Obsidian Tower")
			}
		}
	}
	if !foundAria || !foundTower {
		t.Fatalf("expected findings for Aria and Obsidian Tower")
	}
}

func TestPhase2EntryUseCase_UpdatesSummaryFromLaterChunk(t *testing.T) {
	mockModel := &mockRouterModel{
		outputs: [][]phase2EntityCandidateRaw{
			{
				{Name: "Aria", Evidence: "Aria", Summary: "Mage"},
			},
			{
				{Name: "Aria", Evidence: "Aria", Summary: "Mage of the Crimson Order"},
			},
		},
	}

	useCase := &Phase2EntryUseCase{
		extractors: map[string]*Phase2EntityExtractorUseCase{
			"character": newMockExtractorWithModel("character", mockModel),
		},
		logger: logger.New(),
	}

	ctx := context.Background()
	output, err := useCase.Execute(ctx, Phase2EntryInput{
		Context:               "World of Test",
		MaxCandidatesPerChunk: 3,
		Chunks: []Phase2RoutedChunk{
			{
				ParagraphID: 0,
				ChunkID:     0,
				StartOffset: 0,
				EndOffset:   40,
				Text:        "Aria arrived.",
				Types:       []string{"character"},
			},
			{
				ParagraphID: 1,
				ChunkID:     0,
				StartOffset: 50,
				EndOffset:   90,
				Text:        "Aria serves the Crimson Order.",
				Types:       []string{"character"},
			},
		},
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(output.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(output.Findings))
	}

	if output.Findings[0].Summary != "Mage of the Crimson Order" {
		t.Fatalf("expected summary update, got %q", output.Findings[0].Summary)
	}
	if len(output.Findings[0].Occurrences) != 2 {
		t.Fatalf("expected 2 occurrences, got %d", len(output.Findings[0].Occurrences))
	}

	if len(mockModel.prompts) < 2 {
		t.Fatalf("expected prompts for both chunks, got %d", len(mockModel.prompts))
	}
	if !strings.Contains(mockModel.prompts[1], "ALREADY FOUND") || !strings.Contains(mockModel.prompts[1], "Aria") {
		t.Fatalf("expected already found entities in prompt")
	}
}

type mockRouterModel struct {
	outputs [][]phase2EntityCandidateRaw
	call    int
	prompts []string
}

func (m *mockRouterModel) Generate(_ context.Context, prompt string) (string, error) {
	m.prompts = append(m.prompts, prompt)
	if len(m.outputs) == 0 {
		return `{"candidates":[]}`, nil
	}
	index := m.call
	if index >= len(m.outputs) {
		index = len(m.outputs) - 1
	}
	m.call++
	payload := phase2EntityExtractorRawOutput{Candidates: m.outputs[index]}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func newMockExtractor(entityType string, outputs [][]phase2EntityCandidateRaw) *Phase2EntityExtractorUseCase {
	return newMockExtractorWithModel(entityType, &mockRouterModel{outputs: outputs})
}

func newMockExtractorWithModel(entityType string, model *mockRouterModel) *Phase2EntityExtractorUseCase {
	return &Phase2EntityExtractorUseCase{
		entityType: entityType,
		logger:     logger.New(),
		model:      model,
	}
}
