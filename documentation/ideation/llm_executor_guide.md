# LLM Executor Guide

This guide explains how to use the LLM executor (rate-limited, multi-provider LLM dispatcher)
from application code.

## Why use the executor?
- Centralizes rate limits across the service.
- Supports multiple LLM providers with a unified API.
- Limits parallelism safely.
- Standardizes logging and error handling.

## Required packages
- Executor types live in `llm-gateway-service/internal/platform/llm/executor`
- LLM request/response types use `llm-gateway-service/internal/ports/llm`

## Basic usage
1) Inject the executor into your use case constructor.
2) Build a `llm.RouterPrompt`.
3) Call `executor.Generate(...)`.

```go
type MyUseCase struct {
	executor *executor.Executor
	logger   *logger.Logger
}

func NewMyUseCase(executor *executor.Executor, logger *logger.Logger) *MyUseCase {
	return &MyUseCase{
		executor: executor,
		logger:   logger,
	}
}

func (uc *MyUseCase) Execute(ctx context.Context, input Input) (Output, error) {
	prompt := llm.RouterPrompt{
		System: "System instructions...",
		User:   "User prompt...",
	}

	resp, err := uc.executor.Submit(ctx, prompt, "gemini")
	if err != nil {
		return Output{}, err
	}

	text := resp
	// parse/validate text here
	return Output{Text: text}, nil
}
```

## Multi-provider routing
The executor can register multiple providers and dispatch based on the request’s `Model`.
Use a consistent model name across the codebase (ex.: `gemini-2.5-flash`).

## Concurrency
Executor handles internal concurrency based on the configured limits. Avoid creating
your own goroutine fan-out for LLM calls unless you also limit concurrency.

## Logging
Executor logs rate-limit events and responses. Your use case can log higher-level
events (start/finish, input size, etc).

## Testing
For unit tests, mock the executor with a fake implementation that returns deterministic
strings. For integration tests, call the real executor behind the `LLM_TESTS_ENABLED`
flag to avoid token burn.

## Note: RouterModel adapter vs direct executor
Today the extraction pipeline receives a `llm.RouterModel`, which is backed by
the executor via `executor.NewRouterModelAdapter(...)`. This already uses the
executor internally. Passing the executor directly into each use case would make
provider selection and future per-call options more explicit, but is not required
for correctness right now.

## Example: ingestion summary
```go
executorConfig := executor.ConfigFromEnv("gemini")
providers := []executor.Provider{
	executor.NewRouterModelProvider("gemini", gemini.NewRouterModel(apiKey, model)),
}
llmExecutor, err := executor.New(executorConfig, providers)
if err != nil {
	return err
}

summaryUC := ingest.NewGenerateSummaryUseCase(llmExecutor, "gemini", logger)
out, err := summaryUC.Execute(ctx, ingest.GenerateSummaryInput{
	EntityType: "character",
	Name:       "Aria",
	Contents:   []string{"Aria is a mage."},
	Context:    "",
	MaxItems:   3,
})
```

## Example: entity extraction (router phase)
```go
executorConfig := executor.ConfigFromEnv("gemini")
providers := []executor.Provider{
	executor.NewRouterModelProvider("gemini", gemini.NewRouterModel(apiKey, model)),
}
llmExecutor, err := executor.New(executorConfig, providers)
if err != nil {
	return err
}

routerModel := executor.NewRouterModelAdapter(llmExecutor, "gemini")
routerUC := entity_extraction.NewPhase1EntityTypeRouterUseCase(routerModel, logger)
out, err := routerUC.Execute(ctx, entity_extraction.Phase1EntityTypeRouterInput{
	Text:    "Aria entered the Obsidian Tower.",
	Context: "",
	EntityTypes: []string{
		"character",
		"location",
		"faction",
	},
	MaxCandidates: 4,
})
```

## Common pitfalls
- Forgetting to pass the executor through constructors.
- Using different model names across modules.
- Skipping JSON validation when the prompt requires strict schema.

## See also
- `documentation/ideation/entity_extract_sse_events.md`
- `documentation/ideation/entity_extraction_add_entity_guide.md`
