package rpg

import (
	"encoding/json"
	"fmt"
)

// ValidateStatsAgainstSchema validates stats JSON against a schema
// This is a basic validator - can be extended for strict validation
func ValidateStatsAgainstSchema(stats json.RawMessage, schema json.RawMessage) error {
	// Parse stats
	var statsMap map[string]interface{}
	if err := json.Unmarshal(stats, &statsMap); err != nil {
		return fmt.Errorf("invalid stats JSON: %w", err)
	}

	// Parse schema
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("invalid schema JSON: %w", err)
	}

	// Basic validation: check if schema defines attributes
	schemaAttrs, ok := schemaMap["attributes"]
	if !ok {
		// Schema might not have attributes field, that's okay for now
		return nil
	}

	attrs, ok := schemaAttrs.([]interface{})
	if !ok {
		return nil
	}

	// Extract allowed keys from schema
	allowedKeys := make(map[string]bool)
	for _, attr := range attrs {
		attrMap, ok := attr.(map[string]interface{})
		if !ok {
			continue
		}
		if key, ok := attrMap["key"].(string); ok {
			allowedKeys[key] = true
		}
	}

	// For now, we just validate JSON structure
	// Strict validation (types, min/max) can be added later
	return nil
}

