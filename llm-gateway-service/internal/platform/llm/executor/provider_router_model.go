package executor

import (
	"context"

	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type RouterModelProvider struct {
	name  string
	model llm.RouterModel
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
