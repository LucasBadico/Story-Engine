package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
)

var _ embeddings.Embedder = (*OllamaEmbedder)(nil)

// OllamaEmbedder implements the Embedder interface using Ollama API
type OllamaEmbedder struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbedder creates a new Ollama embedder
func NewOllamaEmbedder(cfg *config.Config) *OllamaEmbedder {
	baseURL := cfg.Embedding.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaEmbedder{
		baseURL: baseURL,
		model:   cfg.Embedding.Model,
		client:  &http.Client{},
	}
}

// EmbedText generates an embedding for a single text
func (e *OllamaEmbedder) EmbedText(text string) ([]float32, error) {
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
func (e *OllamaEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Ollama processes one at a time, so we'll do sequential calls
	// This could be optimized with goroutines if needed
	embeddings := make([][]float32, 0, len(texts))
	for _, text := range texts {
		embedding, err := e.embedSingle(text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text: %w", err)
		}
		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

// embedSingle generates embedding for a single text
func (e *OllamaEmbedder) embedSingle(text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  e.model,
		"prompt": text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", e.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert float64 to float32
	embedding := make([]float32, len(apiResp.Embedding))
	for i, v := range apiResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// Dimension returns the dimension of embeddings
// Common Ollama embedding models:
// - nomic-embed-text: 768
// - mxbai-embed-large: 1024
// - all-minilm: 384
func (e *OllamaEmbedder) Dimension() int {
	// Default to 768 for nomic-embed-text
	// This should ideally be configurable or detected from the model
	switch e.model {
	case "nomic-embed-text":
		return 768
	case "mxbai-embed-large":
		return 1024
	case "all-minilm":
		return 384
	default:
		// Default assumption for most Ollama embedding models
		return 768
	}
}

