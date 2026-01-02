package embeddings

// MockEmbedder is a mock implementation of Embedder for testing
type MockEmbedder struct {
	dimension int
	embedFunc func(text string) ([]float32, error)
	embedBatchFunc func(texts []string) ([][]float32, error)
}

// NewMockEmbedder creates a new mock embedder
func NewMockEmbedder(dimension int) *MockEmbedder {
	return &MockEmbedder{
		dimension: dimension,
		embedFunc: func(text string) ([]float32, error) {
			// Return a simple fake embedding
			embedding := make([]float32, dimension)
			for i := range embedding {
				embedding[i] = 0.1 // Simple fake value
			}
			return embedding, nil
		},
		embedBatchFunc: func(texts []string) ([][]float32, error) {
			embeddings := make([][]float32, len(texts))
			for i := range embeddings {
				embedding := make([]float32, dimension)
				for j := range embedding {
					embedding[j] = 0.1
				}
				embeddings[i] = embedding
			}
			return embeddings, nil
		},
	}
}

// SetEmbedFunc allows setting a custom embed function
func (m *MockEmbedder) SetEmbedFunc(fn func(text string) ([]float32, error)) {
	m.embedFunc = fn
}

// SetEmbedBatchFunc allows setting a custom embed batch function
func (m *MockEmbedder) SetEmbedBatchFunc(fn func(texts []string) ([][]float32, error)) {
	m.embedBatchFunc = fn
}

// EmbedText generates an embedding for the given text
func (m *MockEmbedder) EmbedText(text string) ([]float32, error) {
	return m.embedFunc(text)
}

// EmbedBatch generates embeddings for multiple texts
func (m *MockEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	return m.embedBatchFunc(texts)
}

// Dimension returns the dimension of embeddings produced by this embedder
func (m *MockEmbedder) Dimension() int {
	return m.dimension
}


