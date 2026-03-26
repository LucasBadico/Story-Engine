package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client communicates with the LLM Gateway HTTP API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a gateway client pointing at baseURL (e.g. http://localhost:8081).
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ---- Canon types (matches RetrieveCanonOutput from gateway) ----

type CanonResponse struct {
	StoryID     string        `json:"story_id"`
	AsOfChapter int           `json:"as_of_chapter"`
	AsOfDay     *int          `json:"as_of_day,omitempty"`
	Mode        string        `json:"mode"`
	Canon       []CanonEntry  `json:"canon"`
	States      []StateEntry  `json:"states"`
	Events      []EventEntry  `json:"events"`
}

type CanonEntry struct {
	ID         string            `json:"id"`
	SourceType string            `json:"source_type"`
	SourceID   string            `json:"source_id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Chapter    int               `json:"chapter,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type StateEntry struct {
	ID                 string                 `json:"id"`
	EntityID           string                 `json:"entity_id"`
	EntityType         string                 `json:"entity_type"`
	ChapterNumber      int                    `json:"chapter_number"`
	TimelineDay        *int                   `json:"timeline_day,omitempty"`
	TimelineConfidence *string                `json:"timeline_confidence,omitempty"`
	StateData          map[string]interface{} `json:"state_data"`
}

type EventEntry struct {
	ID                 string   `json:"id"`
	ChapterNumber      int      `json:"chapter_number"`
	TimelineDay        *int     `json:"timeline_day,omitempty"`
	TimelineConfidence *string  `json:"timeline_confidence,omitempty"`
	Description        string   `json:"description"`
	EntitiesInvolved   []string `json:"entities_involved,omitempty"`
}

// GetCanon calls GET /api/v1/stories/{id}/canon with the given parameters.
func (c *Client) GetCanon(ctx context.Context, storyID string, asOfChapter int, asOfDay *int, mode string, eventLimit *int) (*CanonResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/stories/%s/canon", c.baseURL, url.PathEscape(storyID)))
	if err != nil {
		return nil, fmt.Errorf("building canon URL: %w", err)
	}

	q := u.Query()
	q.Set("as_of_chapter", strconv.Itoa(asOfChapter))
	if asOfDay != nil {
		q.Set("as_of_day", strconv.Itoa(*asOfDay))
	}
	if mode != "" {
		q.Set("mode", mode)
	}
	if eventLimit != nil {
		q.Set("event_limit", strconv.Itoa(*eventLimit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating canon request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling canon endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("canon endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var canon CanonResponse
	if err := json.NewDecoder(resp.Body).Decode(&canon); err != nil {
		return nil, fmt.Errorf("decoding canon response: %w", err)
	}
	return &canon, nil
}

// ---- Search types ----

type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type SearchChunk struct {
	ChunkID      string   `json:"chunk_id"`
	DocumentID   string   `json:"document_id"`
	SourceType   string   `json:"source_type"`
	SourceID     string   `json:"source_id"`
	Content      string   `json:"content"`
	Score        float64  `json:"score"`
	BeatType     *string  `json:"beat_type,omitempty"`
	BeatIntent   *string  `json:"beat_intent,omitempty"`
	Characters   []string `json:"characters,omitempty"`
	LocationName *string  `json:"location_name,omitempty"`
	Timeline     *string  `json:"timeline,omitempty"`
	POVCharacter *string  `json:"pov_character,omitempty"`
	ContentKind  *string  `json:"content_kind,omitempty"`
}

type SearchResponse struct {
	Chunks     []SearchChunk `json:"chunks"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// Search calls POST /api/v1/search with the given tenant ID and query.
func (c *Client) Search(ctx context.Context, tenantID string, query string, limit int) (*SearchResponse, error) {
	body := SearchRequest{
		Query: query,
		Limit: limit,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/search", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("creating search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling search endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search endpoint returned %d: %s", resp.StatusCode, string(respBody))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decoding search response: %w", err)
	}
	return &searchResp, nil
}
