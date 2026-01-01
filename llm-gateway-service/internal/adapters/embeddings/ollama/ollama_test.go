package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/story-engine/llm-gateway-service/internal/platform/config"
)

func TestOllamaEmbedder_EmbedText(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("Expected /api/embeddings, got %s", r.URL.Path)
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Return mock embedding response
		response := map[string]interface{}{
			"embedding": []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create embedder with test server URL
	cfg := &config.Config{}
	cfg.Embedding.BaseURL = server.URL
	cfg.Embedding.Model = "nomic-embed-text"
	embedder := NewOllamaEmbedder(cfg)

	// Test EmbedText
	embedding, err := embedder.EmbedText("test text")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embedding) != 5 {
		t.Errorf("Expected embedding length 5, got %d", len(embedding))
	}

	expected := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	for i, v := range embedding {
		if v != expected[i] {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, expected[i], v)
		}
	}
}

func TestOllamaEmbedder_EmbedBatch(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"embedding": []float64{0.1 * float64(callCount), 0.2 * float64(callCount)},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.Config{}
	cfg.Embedding.BaseURL = server.URL
	cfg.Embedding.Model = "nomic-embed-text"
	embedder := NewOllamaEmbedder(cfg)

	// Test EmbedBatch
	texts := []string{"text1", "text2", "text3"}
	embeddings, err := embedder.EmbedBatch(texts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestOllamaEmbedder_EmbedBatch_Empty(t *testing.T) {
	cfg := &config.Config{}
	cfg.Embedding.BaseURL = "http://localhost:11434"
	cfg.Embedding.Model = "nomic-embed-text"
	embedder := NewOllamaEmbedder(cfg)

	embeddings, err := embedder.EmbedBatch([]string{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if embeddings != nil {
		t.Errorf("Expected nil for empty batch, got %v", embeddings)
	}
}

func TestOllamaEmbedder_EmbedText_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	cfg := &config.Config{}
	cfg.Embedding.BaseURL = server.URL
	cfg.Embedding.Model = "nomic-embed-text"
	embedder := NewOllamaEmbedder(cfg)

	_, err := embedder.EmbedText("test")
	if err == nil {
		t.Error("Expected error for API failure, got nil")
	}
}

func TestOllamaEmbedder_Dimension(t *testing.T) {
	tests := []struct {
		model     string
		expected  int
	}{
		{"nomic-embed-text", 768},
		{"mxbai-embed-large", 1024},
		{"all-minilm", 384},
		{"unknown-model", 768}, // default
	}

	for _, tt := range tests {
		cfg := &config.Config{}
		cfg.Embedding.BaseURL = "http://localhost:11434"
		cfg.Embedding.Model = tt.model
		embedder := NewOllamaEmbedder(cfg)

		dim := embedder.Dimension()
		if dim != tt.expected {
			t.Errorf("For model %s, expected dimension %d, got %d", tt.model, tt.expected, dim)
		}
	}
}

func TestNewOllamaEmbedder_DefaultBaseURL(t *testing.T) {
	cfg := &config.Config{}
	cfg.Embedding.Model = "nomic-embed-text"
	embedder := NewOllamaEmbedder(cfg)

	if embedder.baseURL != "http://localhost:11434" {
		t.Errorf("Expected default baseURL http://localhost:11434, got %s", embedder.baseURL)
	}
}

