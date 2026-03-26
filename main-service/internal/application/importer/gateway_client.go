package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// InferenceSceneInput mirrors the gateway's SceneInput type
type InferenceSceneInput struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// InferenceResult mirrors the gateway's InferChapterStructureOutput type
type InferenceResult struct {
	Scenes []InferredScene `json:"scenes"`
}

// InferredScene mirrors the gateway's InferredScene type
type InferredScene struct {
	SceneNumber    int            `json:"scene_number"`
	Title          string         `json:"title"`
	StartParagraph int            `json:"start_paragraph"`
	EndParagraph   int            `json:"end_paragraph"`
	Beats          []InferredBeat `json:"beats"`
}

// InferredBeat mirrors the gateway's InferredBeat type
type InferredBeat struct {
	Type           string `json:"type"`
	Intent         string `json:"intent"`
	StartParagraph int    `json:"start_paragraph"`
	EndParagraph   int    `json:"end_paragraph"`
}

// inferChapterStructureRequest is the JSON body sent to the gateway
type inferChapterStructureRequest struct {
	Scenes []InferenceSceneInput `json:"scenes"`
	Prose  string                `json:"prose"`
}

// GatewayClient is a simple HTTP client for the LLM Gateway inference endpoint
type GatewayClient struct {
	baseURL string
	client  *http.Client
}

// NewGatewayClient creates a new GatewayClient with a generous timeout for LLM inference
func NewGatewayClient(baseURL string) *GatewayClient {
	return &GatewayClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 180 * time.Second, // generous for Gemini Pro
		},
	}
}

// InferChapterStructure calls the gateway to infer chapter structure (scenes and beats)
// from the provided scene descriptions and prose text.
func (c *GatewayClient) InferChapterStructure(ctx context.Context, scenes []InferenceSceneInput, prose string) (*InferenceResult, error) {
	reqBody := inferChapterStructureRequest{
		Scenes: scenes,
		Prose:  prose,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inference request: %w", err)
	}

	url := c.baseURL + "/api/v1/infer-chapter-structure"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create inference request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("inference request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read inference response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inference returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result InferenceResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inference response: %w", err)
	}

	return &result, nil
}
