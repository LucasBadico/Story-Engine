package executor

import (
	"context"

	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type RouterModelAdapter struct {
	executor *Executor
	provider string
}

func NewRouterModelAdapter(executor *Executor, provider string) llm.RouterModel {
	return &RouterModelAdapter{
		executor: executor,
		provider: provider,
	}
}

func (m *RouterModelAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	return m.executor.Submit(ctx, prompt, m.provider)
}
