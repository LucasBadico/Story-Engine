package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Database struct {
		URL string
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	MainService struct {
		GRPCAddr string
	}
	Embedding struct {
		Provider  string
		BaseURL   string
		APIKey    string
		Model     string
		Dimension int
	}
	LLM struct {
		Provider string
		APIKey   string
		Model    string
	}
	Worker struct {
		DebounceMinutes          int
		PollSeconds              int
		BatchSize                int
		ProcessingTimeoutSeconds int
	}
	HTTP struct {
		Addr string
	}
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{}

	// Database
	cfg.Database.URL = getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/storyengine?sslmode=disable")

	// Redis
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvInt("REDIS_DB", 0)

	// Main Service
	cfg.MainService.GRPCAddr = getEnv("MAIN_SERVICE_GRPC_ADDR", "localhost:50051")

	// Embedding
	cfg.Embedding.Provider = getEnv("EMBEDDING_PROVIDER", "ollama")
	cfg.Embedding.BaseURL = getEnv("EMBEDDING_BASE_URL", "http://localhost:11434")
	cfg.Embedding.APIKey = getEnv("EMBEDDING_API_KEY", "")
	cfg.Embedding.Model = getEnv("EMBEDDING_MODEL", "nomic-embed-text")

	// Set default dimension based on provider/model
	if cfg.Embedding.Provider == "openai" {
		cfg.Embedding.Dimension = 1536 // OpenAI ada-002
	} else if cfg.Embedding.Provider == "ollama" {
		// Default for nomic-embed-text
		cfg.Embedding.Dimension = 768
	} else {
		cfg.Embedding.Dimension = 768
	}

	// LLM Router
	cfg.LLM.Provider = getEnv("LLM_PROVIDER", "gemini")
	cfg.LLM.APIKey = getEnv("LLM_API_KEY", getEnv("GEMINI_API_KEY", ""))
	cfg.LLM.Model = getEnv("LLM_MODEL", getEnv("GEMINI_MODEL", ""))

	// Worker
	cfg.Worker.DebounceMinutes = getEnvInt("WORKER_DEBOUNCE_MINUTES", 5)
	cfg.Worker.PollSeconds = getEnvInt("WORKER_POLL_SECONDS", 60)
	cfg.Worker.BatchSize = getEnvInt("WORKER_BATCH_SIZE", 10)
	cfg.Worker.ProcessingTimeoutSeconds = getEnvInt("WORKER_PROCESSING_TIMEOUT_SECONDS", 600)

	// HTTP API
	cfg.HTTP.Addr = getEnv("HTTP_ADDR", ":8081")

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
