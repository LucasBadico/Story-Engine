package relations

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

func getRelationNormalizeParallelism(log *logger.Logger) int {
	return getRelationParallelism("RELATION_NORMALIZE_PARALLELISM", "relation normalize parallelism exceeds CPU count", log)
}

func getRelationMatchParallelism(log *logger.Logger) int {
	return getRelationParallelism("RELATION_MATCH_PARALLELISM", "relation match parallelism exceeds CPU count", log)
}

func getRelationParallelism(envKey string, warnMessage string, log *logger.Logger) int {
	cpuCount := runtime.NumCPU()
	parallelism := cpuCount
	if parallelism < 1 {
		parallelism = 1
	}

	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			parallelism = parsed
		}
	} else if value := strings.TrimSpace(os.Getenv("ENTITY_EXTRACT_PARALLELISM")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			parallelism = parsed
		}
	}

	if log != nil && parallelism > cpuCount {
		log.Warn(
			warnMessage,
			"parallelism", parallelism,
			"cpu_count", cpuCount,
		)
	}

	if parallelism < 1 {
		parallelism = 1
	}

	return parallelism
}
