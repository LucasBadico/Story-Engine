package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
)

const staticTimeout = 5 * time.Second

// FetchRelationTypes loads relation types from main-service.
func FetchRelationTypes(ctx context.Context, baseURL string) (map[string]entity_extraction.Phase6RelationTypeDefinition, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/v1/static/relations"
	var payload map[string]entity_extraction.Phase6RelationTypeDefinition
	if err := fetchJSON(ctx, url, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// FetchRelationMap loads a relation map for a specific entity type.
func FetchRelationMap(ctx context.Context, baseURL, entityType string) (*entity_extraction.Phase5PerEntityRelationMap, error) {
	url := fmt.Sprintf("%s/api/v1/static/relations/%s", strings.TrimRight(baseURL, "/"), entityType)
	var payload entity_extraction.Phase5PerEntityRelationMap
	if err := fetchJSON(ctx, url, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func fetchJSON(ctx context.Context, url string, target interface{}) error {
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("url is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: staticTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
