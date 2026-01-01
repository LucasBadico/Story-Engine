package embeddings

// Embedder defines the interface for generating embeddings
type Embedder interface {
	// EmbedText generates an embedding for the given text
	EmbedText(text string) ([]float32, error)
	
	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(texts []string) ([][]float32, error)
	
	// Dimension returns the dimension of embeddings produced by this embedder
	Dimension() int
}

