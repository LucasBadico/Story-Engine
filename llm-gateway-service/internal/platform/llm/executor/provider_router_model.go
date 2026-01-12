package executor

import (
	"context"

	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type RouterModelProvider struct {
	name  string
	model llm.RouterModel
}

type routerModelMaxOutputTokens interface {
	GenerateWithMaxOutputTokens(ctx context.Context, prompt string, maxOutputTokens int) (string, error)
}

func NewRouterModelProvider(name string, model llm.RouterModel) *RouterModelProvider {
	return &RouterModelProvider{
		name:  name,
		model: model,
	}
}

func (p *RouterModelProvider) Name() string {
	return p.name
}

func (p *RouterModelProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.model.Generate(ctx, prompt)
}

func (p *RouterModelProvider) GenerateWithMaxOutputTokens(ctx context.Context, prompt string, maxOutputTokens int) (string, error) {
	if model, ok := p.model.(routerModelMaxOutputTokens); ok {
		return model.GenerateWithMaxOutputTokens(ctx, prompt, maxOutputTokens)
	}
	return p.model.Generate(ctx, prompt)
}
