package ingest

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

type GenerateSummaryUseCase struct {
	model  llm.RouterModel
	logger *logger.Logger
}

func NewGenerateSummaryUseCase(model llm.RouterModel, logger *logger.Logger) *GenerateSummaryUseCase {
	return &GenerateSummaryUseCase{
		model:  model,
		logger: logger,
	}
}

type GenerateSummaryInput struct {
	EntityType string
	Name       string
	Contents   []string
	Context    string
	MaxItems   int
}

type GenerateSummaryOutput struct {
	Summaries []string
}

//go:embed prompts/generate_entity_summary.prompt
var generateEntitySummaryPromptTemplate string

//go:embed prompts/generate_entity_summary_repair.prompt
var generateEntitySummaryRepairPromptTemplate string

type generateSummaryRawOutput struct {
	Summaries []string `json:"summaries"`
}

func (uc *GenerateSummaryUseCase) Execute(ctx context.Context, input GenerateSummaryInput) (GenerateSummaryOutput, error) {
	entityType := strings.TrimSpace(input.EntityType)
	name := strings.TrimSpace(input.Name)
	contents := normalizeSummaryContents(input.Contents)
	if entityType == "" || name == "" || len(contents) == 0 {
		return GenerateSummaryOutput{}, errors.New("entity type, name, and contents are required")
	}

	maxItems := input.MaxItems
	if maxItems <= 0 {
		maxItems = 3
	}

	prompt := buildGenerateEntitySummaryPrompt(entityType, name, contents, input.Context, maxItems)
	raw, err := uc.model.Generate(ctx, prompt)
	if err != nil {
		uc.logger.Error("summary model failed", "error", err)
		return GenerateSummaryOutput{}, err
	}

	uc.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", raw))

	output, err := parseGenerateSummaryOutput(raw)
	if err != nil {
		uc.logger.Error("failed to parse summary output", "error", err)
		repaired, repairErr := uc.repairSummaryOutput(ctx, raw, input.MaxItems)
		if repairErr != nil {
			return GenerateSummaryOutput{}, err
		}
		output = repaired
	}

	if len(output.Summaries) > maxItems {
		output.Summaries = output.Summaries[:maxItems]
	}

	return GenerateSummaryOutput{
		Summaries: output.Summaries,
	}, nil
}

func buildGenerateEntitySummaryPrompt(entityType string, name string, contents []string, context string, maxItems int) string {
	prompt := generateEntitySummaryPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{entity_type}}", entityType)
	prompt = strings.ReplaceAll(prompt, "{{entity_name}}", name)
	prompt = strings.ReplaceAll(prompt, "{{content_blocks}}", formatSummaryContentBlocks(contents))
	prompt = strings.ReplaceAll(prompt, "{{context_if_any}}", strings.TrimSpace(context))
	prompt = strings.ReplaceAll(prompt, "{{max_items}}", fmt.Sprintf("%d", maxItems))
	return strings.TrimSpace(prompt) + "\n"
}

func buildGenerateEntitySummaryRepairPrompt(rawOutput string, maxItems int) string {
	prompt := generateEntitySummaryRepairPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{raw_output}}", strings.TrimSpace(rawOutput))
	prompt = strings.ReplaceAll(prompt, "{{max_items}}", fmt.Sprintf("%d", maxItems))
	return strings.TrimSpace(prompt) + "\n"
}

func (uc *GenerateSummaryUseCase) repairSummaryOutput(ctx context.Context, raw string, maxItems int) (GenerateSummaryOutput, error) {
	if strings.TrimSpace(raw) == "" {
		return GenerateSummaryOutput{}, errors.New("empty summary output")
	}

	repairPrompt := buildGenerateEntitySummaryRepairPrompt(raw, maxItems)
	repairedRaw, err := uc.model.Generate(ctx, repairPrompt)
	if err != nil {
		uc.logger.Error("summary repair model failed", "error", err)
		return GenerateSummaryOutput{}, err
	}

	uc.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", repairedRaw))

	return parseGenerateSummaryOutput(repairedRaw)
}

func parseGenerateSummaryOutput(raw string) (GenerateSummaryOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return GenerateSummaryOutput{}, errors.New("empty summary output")
	}

	clean = stripCodeFences(clean)

	var output generateSummaryRawOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return GenerateSummaryOutput{Summaries: sanitizeSummaries(output.Summaries)}, nil
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return GenerateSummaryOutput{Summaries: sanitizeSummaries(output.Summaries)}, nil
		}
	}

	return GenerateSummaryOutput{}, errors.New("invalid summary output JSON")
}

func sanitizeSummaries(values []string) []string {
	if len(values) == 1 {
		trimmed := strings.TrimSpace(values[0])
		if trimmed != "" {
			if split := splitSummarySentences(trimmed); len(split) > 1 {
				values = split
			}
		}
	}

	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	return clean
}

func normalizeSummaryContents(contents []string) []string {
	if len(contents) == 0 {
		return nil
	}

	clean := make([]string, 0, len(contents))
	for _, value := range contents {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	return clean
}

func formatSummaryContentBlocks(contents []string) string {
	if len(contents) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, content := range contents {
		builder.WriteString(fmt.Sprintf("[%d]\n\"\"\"\n%s\n\"\"\"\n\n", i+1, content))
	}
	return strings.TrimSpace(builder.String())
}

func splitSummarySentences(text string) []string {
	var sentences []string
	start := 0
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '.', '!', '?':
			isEnd := i+1 == len(text) || text[i+1] == ' '
			if isEnd {
				segment := strings.TrimSpace(text[start : i+1])
				if segment != "" {
					sentences = append(sentences, segment)
				}
				start = i + 1
			}
		}
	}

	if start < len(text) {
		segment := strings.TrimSpace(text[start:])
		if segment != "" {
			sentences = append(sentences, segment)
		}
	}

	return sentences
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
