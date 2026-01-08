package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/story-engine/llm-gateway-service/internal/adapters/db/postgres"
	"github.com/story-engine/llm-gateway-service/internal/adapters/embeddings/ollama"
	"github.com/story-engine/llm-gateway-service/internal/adapters/embeddings/openai"
	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	"github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
	"github.com/story-engine/llm-gateway-service/internal/application/search"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/platform/database"
	"github.com/story-engine/llm-gateway-service/internal/platform/llm/executor"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/transport/httpapi"
	"github.com/story-engine/llm-gateway-service/internal/transport/httpapi/middleware"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	log.Info("Starting ingestion service API...")

	db, err := database.New(cfg)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Database connected")

	pgDB := postgres.NewDB(db)
	documentRepo := postgres.NewDocumentRepository(pgDB)
	chunkRepo := postgres.NewChunkRepository(pgDB)

	var embedder embeddings.Embedder
	switch cfg.Embedding.Provider {
	case "openai":
		embedder = openai.NewOpenAIEmbedder(cfg)
		log.Info("Using OpenAI embedder", "model", cfg.Embedding.Model)
	case "ollama":
		embedder = ollama.NewOllamaEmbedder(cfg)
		log.Info("Using Ollama embedder", "model", cfg.Embedding.Model, "base_url", cfg.Embedding.BaseURL)
	default:
		log.Error("Unsupported embedding provider", "provider", cfg.Embedding.Provider)
		os.Exit(1)
	}

	searchUseCase := search.NewSearchMemoryUseCase(
		chunkRepo,
		documentRepo,
		embedder,
		log,
	)

	searchHandler := httpapi.NewSearchHandler(searchUseCase, log)

	executorConfig := executor.ConfigFromEnv(cfg.LLM.Provider)
	providers := make([]executor.Provider, 0, len(executorConfig.Providers))
	for _, providerCfg := range executorConfig.Providers {
		switch providerCfg.Name {
		case "gemini":
			if cfg.LLM.APIKey == "" {
				log.Error("Missing Gemini API key for LLM executor")
				os.Exit(1)
			}
			providers = append(providers, executor.NewRouterModelProvider(
				"gemini",
				gemini.NewRouterModel(cfg.LLM.APIKey, cfg.LLM.Model),
			))
			log.Info("LLM executor registered Gemini", "model", cfg.LLM.Model)
		default:
			log.Error("Unsupported LLM provider", "provider", providerCfg.Name)
		}
	}
	if len(providers) == 0 {
		log.Error("No LLM providers registered")
		os.Exit(1)
	}

	llmExecutor, err := executor.New(executorConfig, providers)
	if err != nil {
		log.Error("Failed to start LLM executor", "error", err)
		os.Exit(1)
	}

	routerModel := executor.NewRouterModelAdapter(llmExecutor, cfg.LLM.Provider)

	router := entity_extraction.NewPhase1EntityTypeRouterUseCase(routerModel, log)
	extractor := entity_extraction.NewPhase2EntryUseCase(routerModel, log, nil)
	matcher := entity_extraction.NewPhase3MatchUseCase(chunkRepo, documentRepo, embedder, routerModel, log)
	payload := entity_extraction.NewPhaseTempPayloadUseCase()
	entityExtractUseCase := entity_extraction.NewEntityAndRelationshipsExtractor(
		router,
		extractor,
		matcher,
		payload,
		log,
	)
	entityExtractHandler := httpapi.NewEntityExtractHandler(entityExtractUseCase, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", httpapi.Health)
	mux.HandleFunc("/api/v1/search", searchHandler.Search)
	mux.HandleFunc("/api/v1/entity-extract", entityExtractHandler.Extract)
	mux.HandleFunc("/api/v1/entity-extract/stream", entityExtractHandler.ExtractStream)

	handler := middleware.Chain(
		mux,
		middleware.CORS(),
	)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("HTTP server listening", "addr", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down API...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown failed", "error", err)
	}
}
