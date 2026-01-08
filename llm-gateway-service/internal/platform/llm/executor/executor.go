package executor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Provider interface {
	Name() string
	Generate(ctx context.Context, prompt string) (string, error)
}

type ProviderConfig struct {
	Name        string
	MaxParallel int
	QPS         int
}

type Config struct {
	DefaultProvider string
	QueueSize       int
	Providers       []ProviderConfig
}

type Request struct {
	Ctx        context.Context
	Prompt     string
	Provider   string
	RequestID  string
	ResponseCh chan Response
	ErrorCh    chan error
}

type Response struct {
	Text     string
	Provider string
}

type Executor struct {
	inCh            chan *Request
	defaultProvider string
	pools           map[string]*providerPool
}

type providerPool struct {
	provider Provider
	queue    chan *Request
	limiter  *time.Ticker
}

func New(config Config, providers []Provider) (*Executor, error) {
	if len(providers) == 0 {
		return nil, errors.New("no LLM providers configured")
	}

	queueSize := config.QueueSize
	if queueSize <= 0 {
		queueSize = 100
	}

	pools := make(map[string]*providerPool)
	configByName := make(map[string]ProviderConfig)
	for _, cfg := range config.Providers {
		configByName[strings.ToLower(cfg.Name)] = cfg
	}

	for _, provider := range providers {
		name := strings.ToLower(provider.Name())
		cfg, ok := configByName[name]
		if !ok {
			cfg = ProviderConfig{Name: name, MaxParallel: 1}
		}
		if cfg.MaxParallel <= 0 {
			cfg.MaxParallel = 1
		}
		queue := make(chan *Request, queueSize)
		pool := &providerPool{
			provider: provider,
			queue:    queue,
		}
		if cfg.QPS > 0 {
			interval := time.Second / time.Duration(cfg.QPS)
			if interval <= 0 {
				interval = time.Second
			}
			pool.limiter = time.NewTicker(interval)
		}
		for i := 0; i < cfg.MaxParallel; i++ {
			go pool.worker()
		}
		pools[name] = pool
	}

	executor := &Executor{
		inCh:            make(chan *Request, queueSize),
		defaultProvider: strings.ToLower(config.DefaultProvider),
		pools:           pools,
	}

	go executor.dispatch()
	return executor, nil
}

func (e *Executor) Submit(ctx context.Context, prompt string, provider string) (string, error) {
	responseCh := make(chan Response, 1)
	errorCh := make(chan error, 1)

	req := &Request{
		Ctx:        ctx,
		Prompt:     prompt,
		Provider:   provider,
		ResponseCh: responseCh,
		ErrorCh:    errorCh,
	}

	if err := e.Enqueue(req); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errorCh:
		return "", err
	case response := <-responseCh:
		return response.Text, nil
	}
}

func (e *Executor) Enqueue(req *Request) error {
	if req == nil {
		return errors.New("request is nil")
	}
	if req.Ctx == nil {
		req.Ctx = context.Background()
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if req.ResponseCh == nil {
		req.ResponseCh = make(chan Response, 1)
	}
	if req.ErrorCh == nil {
		req.ErrorCh = make(chan error, 1)
	}

	select {
	case e.inCh <- req:
		return nil
	default:
		return errors.New("llm executor queue is full")
	}
}

func (e *Executor) dispatch() {
	for req := range e.inCh {
		providerName := strings.ToLower(strings.TrimSpace(req.Provider))
		if providerName == "" || providerName == "auto" {
			providerName = e.defaultProvider
		}
		pool, ok := e.pools[providerName]
		if !ok {
			sendError(req, fmt.Errorf("llm provider not registered: %s", providerName))
			continue
		}
		select {
		case pool.queue <- req:
		default:
			sendError(req, fmt.Errorf("llm provider queue is full: %s", providerName))
		}
	}
}

func (p *providerPool) worker() {
	for req := range p.queue {
		if req.Ctx != nil && req.Ctx.Err() != nil {
			sendError(req, req.Ctx.Err())
			continue
		}

		if p.limiter != nil {
			select {
			case <-req.Ctx.Done():
				sendError(req, req.Ctx.Err())
				continue
			case <-p.limiter.C:
			}
		}

		text, err := p.provider.Generate(req.Ctx, req.Prompt)
		if err != nil {
			sendError(req, err)
			continue
		}
		sendResponse(req, Response{
			Text:     text,
			Provider: p.provider.Name(),
		})
	}
}

func sendResponse(req *Request, response Response) {
	select {
	case req.ResponseCh <- response:
	default:
	}
}

func sendError(req *Request, err error) {
	select {
	case req.ErrorCh <- err:
	default:
	}
}
