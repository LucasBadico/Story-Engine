package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultModel           = "gemini-1.5-flash"
	defaultMaxOutputTokens = 4096
	extendedTimeoutSeconds = 120
)

type RouterModel struct {
	apiKey string
	model  string
	client *http.Client
}

func NewRouterModel(apiKey string, model string) *RouterModel {
	trimmed := model
	if trimmed == "" {
		trimmed = defaultModel
	}

	return &RouterModel{
		apiKey: apiKey,
		model:  trimmed,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type requestPayload struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig,omitempty"`
}

type generationConfig struct {
	Temperature      float32 `json:"temperature,omitempty"`
	MaxOutputTokens  int     `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type responsePayload struct {
	Candidates     []candidate     `json:"candidates"`
	PromptFeedback *promptFeedback `json:"promptFeedback,omitempty"`
}

type candidate struct {
	Content      content `json:"content"`
	FinishReason string  `json:"finishReason,omitempty"`
}

type promptFeedback struct {
	BlockReason string `json:"blockReason,omitempty"`
}

func (m *RouterModel) Generate(ctx context.Context, prompt string) (string, error) {
	return m.generateWithMaxOutputTokens(ctx, prompt, defaultMaxOutputTokens)
}

func (m *RouterModel) GenerateWithMaxOutputTokens(ctx context.Context, prompt string, maxOutputTokens int) (string, error) {
	if maxOutputTokens <= 0 {
		maxOutputTokens = defaultMaxOutputTokens
	}
	return m.generateWithMaxOutputTokens(ctx, prompt, maxOutputTokens)
}

func (m *RouterModel) generateWithMaxOutputTokens(ctx context.Context, prompt string, maxOutputTokens int) (string, error) {
	if m.apiKey == "" {
		return "", errors.New("gemini api key is required")
	}
	if prompt == "" {
		return "", errors.New("prompt is required")
	}

	const maxAttempts = 3

	client := m.client
	if maxOutputTokens > defaultMaxOutputTokens && client.Timeout < extendedTimeoutSeconds*time.Second {
		client = &http.Client{Timeout: extendedTimeoutSeconds * time.Second}
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		body := requestPayload{
			Contents: []content{
				{
					Role: "user",
					Parts: []part{
						{Text: prompt},
					},
				},
			},
			GenerationConfig: generationConfig{
				Temperature:      0.2,
				MaxOutputTokens:  maxOutputTokens,
				ResponseMimeType: "application/json",
			},
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request: %w", err)
		}

		url := fmt.Sprintf(
			"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
			m.model,
			m.apiKey,
		)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("gemini request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if shouldRetryGeminiStatus(resp.StatusCode) && attempt < maxAttempts {
				time.Sleep(backoffDuration(attempt))
				continue
			}
			return "", fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
		}

		var parsed responsePayload
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			return "", fmt.Errorf("failed to decode gemini response: %w", err)
		}

		if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
			return "", errors.New("gemini response missing content")
		}

		finishReason := strings.TrimSpace(parsed.Candidates[0].FinishReason)
		blockReason := ""
		if parsed.PromptFeedback != nil {
			blockReason = strings.TrimSpace(parsed.PromptFeedback.BlockReason)
		}
		log.Printf("[INFO] gemini finish_reason=%s block_reason=%s", finishReason, blockReason)

		text := parsed.Candidates[0].Content.Parts[0].Text
		if finishReason != "" && finishReason != "STOP" {
			if attempt < maxAttempts {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			if strings.TrimSpace(text) == "" {
				return "", fmt.Errorf("gemini finish_reason=%s", finishReason)
			}
			log.Printf("[WARN] gemini finish_reason=%s; returning partial text", finishReason)
			return text, nil
		}

		return text, nil
	}

	return "", errors.New("gemini request failed after retries")
}

func shouldRetryGeminiStatus(status int) bool {
	if status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable {
		return true
	}
	return status >= 500 && status <= 599
}

func backoffDuration(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 500 * time.Millisecond
	case 2:
		return time.Second
	default:
		return 2 * time.Second
	}
}
