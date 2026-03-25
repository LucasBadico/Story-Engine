package mainservice

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

// Client communicates with the Main Service HTTP API.
type Client struct {
	baseURL    string
	tenantID   string
	httpClient *http.Client
}

// NewClient creates a main service client.
func NewClient(baseURL, tenantID string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		tenantID: tenantID,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ---- Plot types ----

type PlotEntity struct {
	ID         string `json:"id"`
	Role       string `json:"role"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

type Plot struct {
	ID          string       `json:"id"`
	StoryID     string       `json:"story_id"`
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	Entities    []PlotEntity `json:"entities"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

type PlotDevelopment struct {
	ID                 string  `json:"id"`
	PlotID             string  `json:"plot_id"`
	ChapterNumber      int     `json:"chapter_number"`
	Action             string  `json:"action"`
	Description        string  `json:"description"`
	Notes              string  `json:"notes,omitempty"`
	TimelineDay        *int    `json:"timeline_day,omitempty"`
	TimelineConfidence *string `json:"timeline_confidence,omitempty"`
	TimelineEventID    *string `json:"timeline_event_id,omitempty"`
	CreatedAt          string  `json:"created_at"`
}

// ListPlots calls GET /api/v1/stories/{story_id}/plots
func (c *Client) ListPlots(ctx context.Context, storyID string) ([]Plot, error) {
	u := fmt.Sprintf("%s/api/v1/stories/%s/plots", c.baseURL, url.PathEscape(storyID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating plots request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", c.tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling plots endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("plots endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var plots []Plot
	if err := json.NewDecoder(resp.Body).Decode(&plots); err != nil {
		return nil, fmt.Errorf("decoding plots response: %w", err)
	}
	return plots, nil
}

// GetPlot calls GET /api/v1/plots/{id}
func (c *Client) GetPlot(ctx context.Context, plotID string) (*Plot, error) {
	u := fmt.Sprintf("%s/api/v1/plots/%s", c.baseURL, url.PathEscape(plotID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating plot request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", c.tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling plot endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("plot endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var plot Plot
	if err := json.NewDecoder(resp.Body).Decode(&plot); err != nil {
		return nil, fmt.Errorf("decoding plot response: %w", err)
	}
	return &plot, nil
}

// ListPlotDevelopments calls GET /api/v1/plots/{id}/developments?as_of_chapter=N
func (c *Client) ListPlotDevelopments(ctx context.Context, plotID string, asOfChapter *int) ([]PlotDevelopment, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/plots/%s/developments", c.baseURL, url.PathEscape(plotID)))
	if err != nil {
		return nil, fmt.Errorf("building developments URL: %w", err)
	}

	if asOfChapter != nil {
		q := u.Query()
		q.Set("as_of_chapter", strconv.Itoa(*asOfChapter))
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating developments request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", c.tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling developments endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("developments endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var devs []PlotDevelopment
	if err := json.NewDecoder(resp.Body).Decode(&devs); err != nil {
		return nil, fmt.Errorf("decoding developments response: %w", err)
	}
	return devs, nil
}
