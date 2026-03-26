package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/world"
)

// HTTPAPIClient implements MainServiceAPI by calling the main service REST API.
type HTTPAPIClient struct {
	baseURL  string
	tenantID uuid.UUID
	client   *http.Client
}

// NewHTTPAPIClient creates a new HTTP API client for the main service.
func NewHTTPAPIClient(baseURL string, tenantID uuid.UUID) *HTTPAPIClient {
	return &HTTPAPIClient{
		baseURL:  baseURL,
		tenantID: tenantID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- helpers ---

func (c *HTTPAPIClient) doJSON(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", c.tenantID.String())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

// --- Listing ---

func (c *HTTPAPIClient) ListCharacters(ctx context.Context, worldID uuid.UUID) ([]*world.Character, error) {
	var resp struct {
		Characters []*world.Character `json:"characters"`
	}
	path := fmt.Sprintf("/api/v1/worlds/%s/characters?limit=10000&offset=0", worldID)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Characters, nil
}

func (c *HTTPAPIClient) ListLocations(ctx context.Context, worldID uuid.UUID) ([]*world.Location, error) {
	var resp struct {
		Locations []*world.Location `json:"locations"`
	}
	path := fmt.Sprintf("/api/v1/worlds/%s/locations?limit=10000&offset=0", worldID)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Locations, nil
}

func (c *HTTPAPIClient) ListArtifacts(ctx context.Context, worldID uuid.UUID) ([]*world.Artifact, error) {
	var resp struct {
		Artifacts []*world.Artifact `json:"artifacts"`
	}
	path := fmt.Sprintf("/api/v1/worlds/%s/artifacts?limit=10000&offset=0", worldID)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Artifacts, nil
}

func (c *HTTPAPIClient) ListChapters(ctx context.Context, storyID uuid.UUID) ([]*story.Chapter, error) {
	var resp struct {
		Chapters []*story.Chapter `json:"chapters"`
	}
	path := fmt.Sprintf("/api/v1/stories/%s/chapters", storyID)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Chapters, nil
}

// --- Creation ---

func (c *HTTPAPIClient) CreateCharacter(ctx context.Context, input APICreateCharacterInput) (*world.Character, error) {
	body := map[string]any{
		"name":        input.Name,
		"description": input.Description,
	}
	var resp struct {
		Character *world.Character `json:"character"`
	}
	path := fmt.Sprintf("/api/v1/worlds/%s/characters", input.WorldID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Character, nil
}

func (c *HTTPAPIClient) CreateLocation(ctx context.Context, input APICreateLocationInput) (*world.Location, error) {
	body := map[string]any{
		"name":        input.Name,
		"description": input.Description,
		"type":        "general",
	}
	var resp struct {
		Location *world.Location `json:"location"`
	}
	path := fmt.Sprintf("/api/v1/worlds/%s/locations", input.WorldID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Location, nil
}

func (c *HTTPAPIClient) CreateArtifact(ctx context.Context, input APICreateArtifactInput) (*world.Artifact, error) {
	body := map[string]any{
		"name":          input.Name,
		"description":   input.Description,
		"rarity":        "common",
		"character_ids": []string{},
		"location_ids":  []string{},
	}
	var resp struct {
		Artifact *world.Artifact `json:"artifact"`
	}
	path := fmt.Sprintf("/api/v1/worlds/%s/artifacts", input.WorldID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Artifact, nil
}

func (c *HTTPAPIClient) CreateChapter(ctx context.Context, input APICreateChapterInput) (*story.Chapter, error) {
	body := map[string]any{
		"story_id": input.StoryID.String(),
		"number":   input.Number,
		"title":    input.Title,
	}
	if input.Status != "" {
		body["status"] = input.Status
	}
	var resp struct {
		Chapter *story.Chapter `json:"chapter"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/chapters", body, &resp); err != nil {
		return nil, err
	}
	return resp.Chapter, nil
}

func (c *HTTPAPIClient) CreateScene(ctx context.Context, input APICreateSceneInput) (*story.Scene, error) {
	body := map[string]any{
		"story_id":  input.StoryID.String(),
		"order_num": input.OrderNum,
		"goal":      input.Goal,
	}
	if input.ChapterID != nil {
		body["chapter_id"] = input.ChapterID.String()
	}
	var resp struct {
		Scene *story.Scene `json:"scene"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/scenes", body, &resp); err != nil {
		return nil, err
	}
	return resp.Scene, nil
}

func (c *HTTPAPIClient) CreateBeat(ctx context.Context, input APICreateBeatInput) (*story.Beat, error) {
	body := map[string]any{
		"scene_id":  input.SceneID.String(),
		"order_num": input.OrderNum,
		"type":      string(input.Type),
		"intent":    input.Intent,
	}
	var resp struct {
		Beat *story.Beat `json:"beat"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/beats", body, &resp); err != nil {
		return nil, err
	}
	return resp.Beat, nil
}

func (c *HTTPAPIClient) CreateContentBlock(ctx context.Context, input APICreateContentBlockInput) (*story.ContentBlock, error) {
	if input.ChapterID == nil {
		return nil, fmt.Errorf("chapter_id is required for content block creation")
	}
	body := map[string]any{
		"type":    string(input.Type),
		"kind":    string(input.Kind),
		"content": input.Content,
	}
	if input.OrderNum != nil {
		body["order_num"] = *input.OrderNum
	}
	var resp struct {
		ContentBlock *story.ContentBlock `json:"content_block"`
	}
	path := fmt.Sprintf("/api/v1/chapters/%s/content-blocks", input.ChapterID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.ContentBlock, nil
}

func (c *HTTPAPIClient) CreateContentAnchor(ctx context.Context, contentBlockID uuid.UUID, entityType string, entityID uuid.UUID) (*story.ContentAnchor, error) {
	body := map[string]any{
		"entity_type": entityType,
		"entity_id":   entityID.String(),
	}
	var resp struct {
		Anchor *story.ContentAnchor `json:"anchor"`
	}
	path := fmt.Sprintf("/api/v1/content-blocks/%s/anchors", contentBlockID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Anchor, nil
}

// Newer entities use PascalCase JSON since domain structs lack json tags.
// Go's json.Unmarshal is case-insensitive, so we can decode into the same structs.

func (c *HTTPAPIClient) CreateCharacterState(ctx context.Context, input APICreateCharacterStateInput) (*world.CharacterState, error) {
	body := map[string]any{
		"chapter_number": input.ChapterNumber,
		"emotion":        input.Emotion,
		"goal":           input.Goal,
		"secrets":        input.Secrets,
		"clothes":        input.Clothes,
		"physical_state": input.PhysicalState,
	}
	if input.ChapterID != nil {
		body["chapter_id"] = input.ChapterID.String()
	}
	if input.TimelineDay != nil {
		body["timeline_day"] = *input.TimelineDay
	}
	if input.TimelineConfidence != nil {
		body["timeline_confidence"] = *input.TimelineConfidence
	}
	var resp struct {
		CharacterState *world.CharacterState `json:"character_state"`
	}
	path := fmt.Sprintf("/api/v1/characters/%s/states", input.CharacterID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.CharacterState, nil
}

func (c *HTTPAPIClient) CreateRelationshipState(ctx context.Context, input APICreateRelationshipStateInput) (*world.RelationshipState, error) {
	body := map[string]any{
		"chapter_id":          input.ChapterID.String(),
		"chapter_number":      input.ChapterNumber,
		"tension_description": input.TensionDescription,
		"entities":            input.Entities,
	}
	if input.TimelineDay != nil {
		body["timeline_day"] = *input.TimelineDay
	}
	if input.TimelineConfidence != nil {
		body["timeline_confidence"] = *input.TimelineConfidence
	}
	var resp struct {
		RelationshipState *world.RelationshipState `json:"relationship_state"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/relationship-states", body, &resp); err != nil {
		return nil, err
	}
	return resp.RelationshipState, nil
}

func (c *HTTPAPIClient) CreateTimelineEvent(ctx context.Context, input APICreateTimelineEventInput) (*story.TimelineEvent, error) {
	body := map[string]any{
		"chapter_id":     input.ChapterID.String(),
		"chapter_number": input.ChapterNumber,
		"description":    input.Description,
	}
	if input.TimelineDay != nil {
		body["timeline_day"] = *input.TimelineDay
	}
	if input.TimelineConfidence != nil {
		body["timeline_confidence"] = *input.TimelineConfidence
	}
	var resp struct {
		TimelineEvent *story.TimelineEvent `json:"timeline_event"`
	}
	path := fmt.Sprintf("/api/v1/stories/%s/timeline-events", input.StoryID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.TimelineEvent, nil
}

func (c *HTTPAPIClient) CreatePlot(ctx context.Context, input APICreatePlotInput) (*story.Plot, error) {
	body := map[string]any{
		"title":       input.Title,
		"type":        input.Type,
		"description": input.Description,
	}
	var resp struct {
		Plot *story.Plot `json:"plot"`
	}
	path := fmt.Sprintf("/api/v1/stories/%s/plots", input.StoryID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Plot, nil
}

func (c *HTTPAPIClient) CreatePlotDevelopment(ctx context.Context, input APICreatePlotDevelopmentInput) (*story.PlotDevelopment, error) {
	body := map[string]any{
		"chapter_number": input.ChapterNumber,
		"action":         input.Action,
		"description":    input.Description,
		"notes":          "",
	}
	var resp struct {
		PlotDevelopment *story.PlotDevelopment `json:"plot_development"`
	}
	path := fmt.Sprintf("/api/v1/plots/%s/developments", input.PlotID)
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.PlotDevelopment, nil
}
