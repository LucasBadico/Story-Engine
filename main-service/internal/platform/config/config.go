package config

import (
	"fmt"
	"os"
)

// Config holds application configuration
type Config struct {
	Database DatabaseConfig
	GRPC     GRPCConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// GRPCConfig holds gRPC server settings
type GRPCConfig struct {
	Port             int
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	EnableReflection bool
}

// DSN returns the PostgreSQL connection string
func (c DatabaseConfig) DSN() string {
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
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_NAME", "storyengine"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		GRPC: GRPCConfig{
			Port:             getEnvInt("GRPC_PORT", 9090),
			MaxRecvMsgSize:   getEnvInt("GRPC_MAX_RECV_MSG_SIZE", 4194304), // 4MB
			MaxSendMsgSize:   getEnvInt("GRPC_MAX_SEND_MSG_SIZE", 4194304), // 4MB
			EnableReflection: getEnvBool("GRPC_ENABLE_REFLECTION", true),
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
		if value == "true" || value == "1" {
			return true
		}
		if value == "false" || value == "0" {
			return false
		}
	}
	return defaultValue
}
