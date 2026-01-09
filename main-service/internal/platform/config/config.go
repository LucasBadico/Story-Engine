package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds application configuration
type Config struct {
	Database DatabaseConfig
	GRPC     GRPCConfig
	HTTP     HTTPConfig
	LLM      LLMConfig
	Cleanup  CleanupConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Driver   string // "postgres" or "sqlite"
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	Path     string // SQLite database file path
}

// GRPCConfig holds gRPC server settings
type GRPCConfig struct {
	Port             int
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	EnableReflection bool
}

// HTTPConfig holds HTTP server settings
type HTTPConfig struct {
	Port int
}

// LLMConfig holds LLM gateway integration settings
type LLMConfig struct {
	Enabled bool
	Redis   RedisConfig
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// CleanupConfig holds cleanup job settings
type CleanupConfig struct {
	// Draft relations cleanup settings
	DraftTTLHours int // Time-to-live for draft relations with temp refs (default: 24 hours)

	// Orphan relations cleanup settings
	OrphanRetentionDays int // Days to retain orphan relations before deletion (default: 30 days)

	// Cleanup job intervals (in hours)
	CleanupIntervalHours int // How often to run cleanup jobs (default: 24 hours)
}

// DSN returns the database connection string based on the driver
func (c DatabaseConfig) DSN() string {
	if c.Driver == "sqlite" {
		return c.Path
	}
	// Default to PostgreSQL
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_NAME", "storyengine"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Path:     getEnv("DB_PATH", "./story-engine.db"),
		},
		GRPC: GRPCConfig{
			Port:             getEnvInt("GRPC_PORT", 9090),
			MaxRecvMsgSize:   getEnvInt("GRPC_MAX_RECV_MSG_SIZE", 4194304), // 4MB
			MaxSendMsgSize:   getEnvInt("GRPC_MAX_SEND_MSG_SIZE", 4194304), // 4MB
			EnableReflection: getEnvBool("GRPC_ENABLE_REFLECTION", true),
		},
		HTTP: HTTPConfig{
			Port: getEnvInt("HTTP_PORT", 8080),
		},
		LLM: LLMConfig{
			Enabled: getEnvBool("LLM_GATEWAY_ENABLED", false),
			Redis: RedisConfig{
				Addr:     getEnv("LLM_GATEWAY_REDIS_ADDR", "localhost:6379"),
				Password: getEnv("LLM_GATEWAY_REDIS_PASSWORD", ""),
				DB:       getEnvInt("LLM_GATEWAY_REDIS_DB", 0),
			},
		},
		Cleanup: CleanupConfig{
			DraftTTLHours:        getEnvInt("CLEANUP_DRAFT_TTL_HOURS", 24),
			OrphanRetentionDays:  getEnvInt("CLEANUP_ORPHAN_RETENTION_DAYS", 30),
			CleanupIntervalHours: getEnvInt("CLEANUP_INTERVAL_HOURS", 24),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if strings.ToLower(value) == "true" || value == "1" {
			return true
		}
		if value == "false" || value == "0" {
			return false
		}
	}
	return defaultValue
}
