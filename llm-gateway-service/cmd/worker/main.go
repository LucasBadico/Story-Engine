package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/story-engine/llm-gateway-service/internal/adapters/db/postgres"
	"github.com/story-engine/llm-gateway-service/internal/adapters/embeddings/ollama"
	"github.com/story-engine/llm-gateway-service/internal/adapters/embeddings/openai"
	grpcadapter "github.com/story-engine/llm-gateway-service/internal/adapters/grpc"
	redisadapter "github.com/story-engine/llm-gateway-service/internal/adapters/redis"
	"github.com/story-engine/llm-gateway-service/internal/application/ingest"
	"github.com/story-engine/llm-gateway-service/internal/application/search"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/platform/database"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/worker"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log := logger.New()

	log.Info("Starting ingestion service worker...")

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Database connected")

	// Initialize Redis client
	redisClient, err := redisadapter.NewClient(cfg)
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	log.Info("Redis connected")

	// Initialize PostgreSQL adapter
	pgDB := postgres.NewDB(db)

	// Initialize repositories
	documentRepo := postgres.NewDocumentRepository(pgDB)
	chunkRepo := postgres.NewChunkRepository(pgDB)
	log.Info("Repositories initialized")

	// Initialize ingestion queue
	ingestionQueue := redisadapter.NewIngestionQueue(redisClient)
	log.Info("Ingestion queue initialized")

	// Initialize gRPC client to main-service
	grpcClient, err := grpcadapter.NewMainServiceClient(cfg.MainService.GRPCAddr)
	if err != nil {
		log.Error("Failed to connect to main-service", "error", err, "addr", cfg.MainService.GRPCAddr)
		os.Exit(1)
	}
	defer grpcClient.Close()
	log.Info("gRPC client connected", "addr", cfg.MainService.GRPCAddr)

	// Initialize embedding provider
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

	// Initialize use cases
	ingestStoryUseCase := ingest.NewIngestStoryUseCase(
		grpcClient,
		documentRepo,
		chunkRepo,
		embedder,
		log,
	)
	ingestChapterUseCase := ingest.NewIngestChapterUseCase(
		grpcClient,
		documentRepo,
		chunkRepo,
		embedder,
		log,
	)
	searchMemoryUseCase := search.NewSearchMemoryUseCase(
		chunkRepo,
		documentRepo,
		embedder,
		log,
	)

	log.Info("Use cases initialized")

	// Initialize and start worker
	debouncedWorker := worker.NewDebouncedWorker(
		ingestionQueue,
		ingestStoryUseCase,
		ingestChapterUseCase,
		log,
		cfg,
	)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// Start worker loop
	log.Info("Starting worker",
		"debounce_minutes", cfg.Worker.DebounceMinutes,
		"poll_seconds", cfg.Worker.PollSeconds,
		"embedding_provider", cfg.Embedding.Provider)

	debouncedWorker.Run(ctx)

	log.Info("Shutting down ingestion service worker...")
}

