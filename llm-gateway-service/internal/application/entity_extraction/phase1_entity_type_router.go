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

type Phase1EntityTypeRouterUseCase struct {
	model  llm.RouterModel
	logger *logger.Logger
}

func NewPhase1EntityTypeRouterUseCase(model llm.RouterModel, logger *logger.Logger) *Phase1EntityTypeRouterUseCase {
	return &Phase1EntityTypeRouterUseCase{
		model:  model,
		logger: logger,
	}
}

type Phase1EntityTypeRouterInput struct {
	Text          string
	Context       string
	EntityTypes   []string
	MaxCandidates int
}

type Candidate struct {
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Why        string  `json:"why"`
}

type Phase1EntityTypeRouterOutput struct {
	Candidates []Candidate `json:"candidates"`
}

//go:embed prompts/phase1_entity_type_router.prompt
var phase1EntityTypeRouterPromptTemplate string

//go:embed prompts/phase1_entity_type_router_repair.prompt
var phase1EntityTypeRouterRepairPromptTemplate string

var EntityTypesAvailable string

func (u *Phase1EntityTypeRouterUseCase) Execute(ctx context.Context, input Phase1EntityTypeRouterInput) (Phase1EntityTypeRouterOutput, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return Phase1EntityTypeRouterOutput{}, errors.New("text is required")
	}
	if len(input.EntityTypes) == 0 {
		return Phase1EntityTypeRouterOutput{}, errors.New("entity types are required")
	}

	maxCandidates := input.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = 5
	}

	prompt := buildPhase1EntityTypeRouterPrompt(text, input.Context, input.EntityTypes, maxCandidates)
	raw, err := u.model.Generate(ctx, prompt)
	if err != nil {
		u.logger.Error("router model failed", "error", err)
		return Phase1EntityTypeRouterOutput{}, err
	}

	u.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", raw))

	output, err := parsePhase1EntityTypeRouterOutput(raw)
	if err != nil {
		u.logger.Error("failed to parse router output", "error", err)
		repaired, repairErr := u.repairOutput(ctx, raw, input.EntityTypes)
		if repairErr != nil {
			return Phase1EntityTypeRouterOutput{}, repairErr
		}
		output = repaired
	}

	if len(output.Candidates) > maxCandidates {
		output.Candidates = output.Candidates[:maxCandidates]
	}

	return output, nil
}

func buildPhase1EntityTypeRouterPrompt(text string, context string, entityTypes []string, maxCandidates int) string {
	prompt := phase1EntityTypeRouterPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{entity_types_block}}", renderEntityTypesBlock(entityTypes))
	prompt = strings.ReplaceAll(prompt, "{{selected_text}}", text)
	prompt = strings.ReplaceAll(prompt, "{{context_if_any}}", strings.TrimSpace(context))
	prompt = strings.ReplaceAll(prompt, "{{max_candidates}}", fmt.Sprintf("%d", maxCandidates))
	return strings.TrimSpace(prompt) + "\n"
}

func buildPhase1EntityTypeRouterRepairPrompt(rawOutput string, entityTypes []string) string {
	prompt := phase1EntityTypeRouterRepairPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{entity_types_block}}", renderEntityTypesBlock(entityTypes))
	prompt = strings.ReplaceAll(prompt, "{{raw_output}}", strings.TrimSpace(rawOutput))
	return strings.TrimSpace(prompt) + "\n"
}

func (u *Phase1EntityTypeRouterUseCase) repairOutput(ctx context.Context, raw string, entityTypes []string) (Phase1EntityTypeRouterOutput, error) {
	if strings.TrimSpace(raw) == "" {
		return Phase1EntityTypeRouterOutput{}, errors.New("empty router output")
	}

	repairPrompt := buildPhase1EntityTypeRouterRepairPrompt(raw, entityTypes)
	repairedRaw, err := u.model.Generate(ctx, repairPrompt)
	if err != nil {
		u.logger.Error("router repair model failed", "error", err)
		return Phase1EntityTypeRouterOutput{}, err
	}

	u.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", repairedRaw))

	return parsePhase1EntityTypeRouterOutput(repairedRaw)
}

func parsePhase1EntityTypeRouterOutput(raw string) (Phase1EntityTypeRouterOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return Phase1EntityTypeRouterOutput{}, errors.New("empty router output")
	}

	clean = stripCodeFences(clean)

	var output Phase1EntityTypeRouterOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return output, nil
	}

	if strings.HasPrefix(clean, "[") {
		var candidates []Candidate
		if err := json.Unmarshal([]byte(clean), &candidates); err == nil {
			return Phase1EntityTypeRouterOutput{Candidates: candidates}, nil
		}
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return output, nil
		}
	}

	return Phase1EntityTypeRouterOutput{}, errors.New("invalid router output JSON")
}

func renderEntityTypesBlock(entityTypes []string) string {
	if len(entityTypes) == 0 {
		return ""
	}

	descriptions := map[string]string{
		"character": "A person or sentient being that performs actions, speaks, thinks,\n  or is described as an individual agent.",
		"location":  "A physical or conceptual place where events occur\n  (e.g. buildings, cities, rooms, regions).",
		"artefact":  "A distinct object of narrative importance\n  (e.g. weapons, relics, books, devices, symbols),\n  typically interacted with or referenced as a thing.",
		"faction":   "An organized group, collective, or institution\n  (e.g. guilds, clans, orders, nations, companies),\n  usually referred to by a shared name or identity.",
	}

	builder := strings.Builder{}
	for _, entityType := range entityTypes {
		if entityType == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("- %s:\n", entityType))
		if desc, ok := descriptions[entityType]; ok {
			builder.WriteString("  " + desc + "\n\n")
		} else {
			builder.WriteString("  (no description provided)\n\n")
		}
	}

	return builder.String()
}

func stripCodeFences(input string) string {
	trimmed := strings.TrimSpace(input)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		if strings.HasPrefix(strings.ToLower(trimmed), "json") {
			trimmed = strings.TrimSpace(trimmed[4:])
		}
		if idx := strings.LastIndex(trimmed, "```"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
	}
	return trimmed
}

func extractFirstJSONObject(input string) string {
	start := strings.Index(input, "{")
	if start < 0 {
		return ""
	}

	depth := 0
	for i := start; i < len(input); i++ {
		switch input[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return input[start : i+1]
			}
		}
	}

	return ""
}
