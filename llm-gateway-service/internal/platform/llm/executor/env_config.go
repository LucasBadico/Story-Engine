package executor

import (
	"os"
	"strconv"
	"strings"
)

func ConfigFromEnv(defaultProvider string) Config {
	queueSize := getEnvInt("LLM_EXECUTOR_QUEUE_SIZE", 100)
	providers := providerListFromEnv("LLM_EXECUTOR_PROVIDERS")
	if len(providers) == 0 && defaultProvider != "" {
		providers = []string{defaultProvider}
	}

	entries := make([]ProviderConfig, 0, len(providers))
	for _, name := range providers {
		upper := strings.ToUpper(name)
		entries = append(entries, ProviderConfig{
			Name:        name,
			MaxParallel: getEnvInt("LLM_EXECUTOR_"+upper+"_MAX_PARALLEL", 2),
			QPS:         getEnvInt("LLM_EXECUTOR_"+upper+"_QPS", 0),
		})
	}

	return Config{
		DefaultProvider: defaultProvider,
		QueueSize:       queueSize,
		Providers:       entries,
	}
}

func providerListFromEnv(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func getEnvInt(key string, defaultValue int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
