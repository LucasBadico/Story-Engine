package entity_extraction

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type Phase2EntityExtractorUseCase struct {
	model      llm.RouterModel
	logger     *logger.Logger
	entityType string
}

func NewPhase2CharacterExtractorUseCase(model llm.RouterModel, logger *logger.Logger) *Phase2EntityExtractorUseCase {
	return newPhase2EntityExtractorUseCase(model, logger, "character")
}

func NewPhase2LocationExtractorUseCase(model llm.RouterModel, logger *logger.Logger) *Phase2EntityExtractorUseCase {
	return newPhase2EntityExtractorUseCase(model, logger, "location")
}

func NewPhase2ArtefactExtractorUseCase(model llm.RouterModel, logger *logger.Logger) *Phase2EntityExtractorUseCase {
	return newPhase2EntityExtractorUseCase(model, logger, "artefact")
}

func NewPhase2FactionExtractorUseCase(model llm.RouterModel, logger *logger.Logger) *Phase2EntityExtractorUseCase {
	return newPhase2EntityExtractorUseCase(model, logger, "faction")
}

func newPhase2EntityExtractorUseCase(model llm.RouterModel, logger *logger.Logger, entityType string) *Phase2EntityExtractorUseCase {
	return &Phase2EntityExtractorUseCase{
		model:      model,
		logger:     logger,
		entityType: entityType,
	}
}

type Phase2EntityExtractorInput struct {
	Text          string
	Context       string
	MaxCandidates int
}

type Phase2EntityExtractorOutput struct {
	EntityType string
	Candidates []Phase2EntityCandidate
}

type Phase2EntityCandidate struct {
	Name        string `json:"name"`
	Evidence    string `json:"evidence"`
	StartOffset int
	EndOffset   int
}

type phase2EntityCandidateRaw struct {
	Name     string `json:"name"`
	Evidence string `json:"evidence"`
}

type phase2EntityExtractorRawOutput struct {
	Candidates []phase2EntityCandidateRaw `json:"candidates"`
}

//go:embed prompts/phase2_character_extractor.prompt
var phase2CharacterExtractorPromptTemplate string

//go:embed prompts/phase2_location_extractor.prompt
var phase2LocationExtractorPromptTemplate string

//go:embed prompts/phase2_artefact_extractor.prompt
var phase2ArtefactExtractorPromptTemplate string

//go:embed prompts/phase2_faction_extractor.prompt
var phase2FactionExtractorPromptTemplate string

func (u *Phase2EntityExtractorUseCase) Execute(ctx context.Context, input Phase2EntityExtractorInput) (Phase2EntityExtractorOutput, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return Phase2EntityExtractorOutput{}, errors.New("text is required")
	}

	maxCandidates := input.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = 5
	}

	prompt, err := buildPhase2EntityExtractorPrompt(u.entityType, text, input.Context, maxCandidates)
	if err != nil {
		return Phase2EntityExtractorOutput{}, err
	}

	raw, err := u.model.Generate(ctx, prompt)
	if err != nil {
		u.logger.Error("phase2 extractor model failed", "error", err)
		return Phase2EntityExtractorOutput{}, err
	}

	u.logger.Info("====model answer====\n%s\n===================", raw)

	parsed, err := parsePhase2EntityExtractorOutput(raw)
	if err != nil {
		u.logger.Error("failed to parse phase2 extractor output", "error", err)
		return Phase2EntityExtractorOutput{}, err
	}

	candidates := make([]Phase2EntityCandidate, 0, len(parsed.Candidates))
	for _, candidate := range parsed.Candidates {
		evidence := strings.TrimSpace(candidate.Evidence)
		if evidence == "" {
			evidence = strings.TrimSpace(candidate.Name)
		}
		start, end := findEvidenceOffset(text, evidence)
		if start < 0 {
			u.logger.Warn("evidence not found in text", "entity_type", u.entityType, "evidence", evidence)
			continue
		}
		candidates = append(candidates, Phase2EntityCandidate{
			Name:        strings.TrimSpace(candidate.Name),
			Evidence:    evidence,
			StartOffset: start,
			EndOffset:   end,
		})
	}

	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}

	return Phase2EntityExtractorOutput{
		EntityType: u.entityType,
		Candidates: candidates,
	}, nil
}

func buildPhase2EntityExtractorPrompt(entityType string, text string, context string, maxCandidates int) (string, error) {
	template, ok := phase2PromptTemplateByType(entityType)
	if !ok {
		return "", fmt.Errorf("unsupported entity type: %s", entityType)
	}

	prompt := template
	prompt = strings.ReplaceAll(prompt, "{{selected_text}}", text)
	prompt = strings.ReplaceAll(prompt, "{{context_if_any}}", strings.TrimSpace(context))
	prompt = strings.ReplaceAll(prompt, "{{max_candidates}}", fmt.Sprintf("%d", maxCandidates))
	return strings.TrimSpace(prompt) + "\n", nil
}

func phase2PromptTemplateByType(entityType string) (string, bool) {
	switch entityType {
	case "character":
		return phase2CharacterExtractorPromptTemplate, true
	case "location":
		return phase2LocationExtractorPromptTemplate, true
	case "artefact":
		return phase2ArtefactExtractorPromptTemplate, true
	case "faction":
		return phase2FactionExtractorPromptTemplate, true
	default:
		return "", false
	}
}

func parsePhase2EntityExtractorOutput(raw string) (phase2EntityExtractorRawOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return phase2EntityExtractorRawOutput{}, errors.New("empty extractor output")
	}

	clean = stripCodeFences(clean)

	var output phase2EntityExtractorRawOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return output, nil
	}

	if strings.HasPrefix(clean, "[") {
		var candidates []phase2EntityCandidateRaw
		if err := json.Unmarshal([]byte(clean), &candidates); err == nil {
			return phase2EntityExtractorRawOutput{Candidates: candidates}, nil
		}
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return output, nil
		}
	}

	return phase2EntityExtractorRawOutput{}, errors.New("invalid extractor output JSON")
}

func findEvidenceOffset(text string, evidence string) (int, int) {
	if evidence == "" {
		return -1, -1
	}
	if start := strings.Index(text, evidence); start >= 0 {
		return start, start + len(evidence)
	}
	trimmed := strings.TrimSpace(evidence)
	if trimmed != evidence {
		if start := strings.Index(text, trimmed); start >= 0 {
			return start, start + len(trimmed)
		}
	}
	return -1, -1
}
