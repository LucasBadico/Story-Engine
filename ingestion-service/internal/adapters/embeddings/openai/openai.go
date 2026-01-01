package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/story-engine/ingestion-service/internal/platform/config"
	"github.com/story-engine/ingestion-service/internal/ports/embeddings"
)

var _ embeddings.Embedder = (*OpenAIEmbedder)(nil)

// OpenAIEmbedder implements the Embedder interface using OpenAI API
type OpenAIEmbedder struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIEmbedder creates a new OpenAI embedder
func NewOpenAIEmbedder(cfg *config.Config) *OpenAIEmbedder {
	return &OpenAIEmbedder{
		apiKey: cfg.Embedding.APIKey,
		model:  cfg.Embedding.Model,
		client: &http.Client{},
	}
}

// EmbedText generates an embedding for a single text
func (e *OpenAIEmbedder) EmbedText(text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *OpenAIEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := map[string]interface{}{
		"input": texts,
		"model": e.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert float64 to float32
	embeddings := make([][]float32, len(apiResp.Data))
	for i, item := range apiResp.Data {
		embeddings[i] = make([]float32, len(item.Embedding))
		for j, v := range item.Embedding {
			embeddings[i][j] = float32(v)
		}
	}

	return embeddings, nil
}

// Dimension returns the dimension of embeddings
func (e *OpenAIEmbedder) Dimension() int {
	// OpenAI text-embedding-ada-002 has 1536 dimensions
	// Other models may differ, but this is the default
	return 1536
}

