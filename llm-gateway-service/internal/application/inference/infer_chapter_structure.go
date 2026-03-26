package inference

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/story-engine/llm-gateway-service/internal/platform/llm/executor"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

//go:embed prompts/infer_chapter_structure.prompt
var inferChapterStructurePromptTemplate string

type InferChapterStructureUseCase struct {
	executor *executor.Executor
	provider string
	logger   *logger.Logger
}

func NewInferChapterStructureUseCase(executor *executor.Executor, provider string, logger *logger.Logger) *InferChapterStructureUseCase {
	return &InferChapterStructureUseCase{
		executor: executor,
		provider: provider,
		logger:   logger,
	}
}

// Input types

type SceneInput struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type InferChapterStructureInput struct {
	Scenes []SceneInput `json:"scenes"`
	Prose  string       `json:"prose"`
}

// Output types

type InferredBeat struct {
	Type           string `json:"type"`
	Intent         string `json:"intent"`
	StartParagraph int    `json:"start_paragraph"`
	EndParagraph   int    `json:"end_paragraph"`
}

type InferredScene struct {
	SceneNumber    int            `json:"scene_number"`
	Title          string         `json:"title"`
	StartParagraph int            `json:"start_paragraph"`
	EndParagraph   int            `json:"end_paragraph"`
	Beats          []InferredBeat `json:"beats"`
}

type InferChapterStructureOutput struct {
	Scenes []InferredScene `json:"scenes"`
}

func (uc *InferChapterStructureUseCase) Execute(ctx context.Context, input InferChapterStructureInput) (*InferChapterStructureOutput, error) {
	if len(input.Scenes) == 0 {
		return nil, errors.New("at least one scene is required")
	}
	if strings.TrimSpace(input.Prose) == "" {
		return nil, errors.New("prose text is required")
	}
	if uc.executor == nil {
		return nil, errors.New("llm executor is required")
	}

	// Build prompt
	prompt := buildInferPrompt(input)

	// Call LLM
	raw, err := uc.executor.Submit(ctx, prompt, uc.provider)
	if err != nil {
		uc.logger.Error("chapter structure inference failed", "error", err)
		return nil, fmt.Errorf("llm inference failed: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", raw))

	// Parse response
	output, err := parseInferOutput(raw)
	if err != nil {
		uc.logger.Error("failed to parse inference output", "error", err)
		return nil, fmt.Errorf("failed to parse inference output: %w", err)
	}

	return output, nil
}

func buildInferPrompt(input InferChapterStructureInput) string {
	prompt := inferChapterStructurePromptTemplate

	// Format scenes
	var scenesBuilder strings.Builder
	for _, scene := range input.Scenes {
		scenesBuilder.WriteString(fmt.Sprintf("Scene %d: \"%s\"\nDescription: %s\n\n", scene.Number, scene.Title, scene.Description))
	}
	prompt = strings.ReplaceAll(prompt, "{{scenes}}", strings.TrimSpace(scenesBuilder.String()))

	// Format numbered prose
	paragraphs := strings.Split(input.Prose, "\n\n")
	var proseBuilder strings.Builder
	for i, para := range paragraphs {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" {
			continue
		}
		proseBuilder.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, trimmed))
	}
	prompt = strings.ReplaceAll(prompt, "{{numbered_prose}}", strings.TrimSpace(proseBuilder.String()))

	return strings.TrimSpace(prompt) + "\n"
}

func parseInferOutput(raw string) (*InferChapterStructureOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return nil, errors.New("empty inference output")
	}

	clean = stripCodeFences(clean)

	var output InferChapterStructureOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return &output, nil
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return &output, nil
		}
	}

	return nil, errors.New("invalid inference output JSON")
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
