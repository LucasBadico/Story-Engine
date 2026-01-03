//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	rpgeventapp "github.com/story-engine/main-service/internal/application/rpg/event"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestEventHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)

	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	sceneHandler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, nil, nil, nil, log)

	// Create scene
	sceneBody := `{"story_id": "` + storyID + `", "chapter_id": "` + chapterID + `", "title": "Test Scene", "order_num": 1}`
	sceneReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/scenes", strings.NewReader(sceneBody))
	sceneReq.Header.Set("Content-Type", "application/json")
	sceneReq.Header.Set("X-Tenant-ID", tenantID)
	sceneReq.SetPathValue("id", chapterID)
	sceneW := httptest.NewRecorder()
	withTenantMiddleware(sceneHandler.Create).ServeHTTP(sceneW, sceneReq)

	if sceneW.Code != http.StatusCreated {
		t.Fatalf("failed to create scene: status %d, body: %s", sceneW.Code, sceneW.Body.String())
	}

	var sceneResp map[string]interface{}
	if err := json.NewDecoder(sceneW.Body).Decode(&sceneResp); err != nil {
		t.Fatalf("failed to decode scene response: %v", err)
	}

	sceneObj, ok := sceneResp["scene"].(map[string]interface{})
	if !ok {
		t.Fatalf("scene response missing scene object: %v", sceneResp)
	}

	_, ok = sceneObj["id"].(string)
	if !ok {
		t.Fatalf("scene response missing id: %v", sceneObj)
	}

	eventRepo := postgres.NewEventRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	eventCharacterRepo := postgres.NewEventCharacterRepository(db)
	eventLocationRepo := postgres.NewEventLocationRepository(db)
	eventArtifactRepo := postgres.NewEventArtifactRepository(db)
	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(db)
	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(db)

	createEventUseCase := eventapp.NewCreateEventUseCase(eventRepo, worldRepo, auditLogRepo, log)
	getEventUseCase := eventapp.NewGetEventUseCase(eventRepo, log)
	listEventsUseCase := eventapp.NewListEventsUseCase(eventRepo, log)
	updateEventUseCase := eventapp.NewUpdateEventUseCase(eventRepo, auditLogRepo, log)
	deleteEventUseCase := eventapp.NewDeleteEventUseCase(eventRepo, eventCharacterRepo, eventLocationRepo, eventArtifactRepo, auditLogRepo, log)
	addCharacterUseCase := eventapp.NewAddCharacterToEventUseCase(eventRepo, characterRepo, eventCharacterRepo, log)
	removeCharacterUseCase := eventapp.NewRemoveCharacterFromEventUseCase(eventCharacterRepo, log)
	getCharactersUseCase := eventapp.NewGetEventCharactersUseCase(eventCharacterRepo, log)
	addLocationUseCase := eventapp.NewAddLocationToEventUseCase(eventRepo, locationRepo, eventLocationRepo, log)
	removeLocationUseCase := eventapp.NewRemoveLocationFromEventUseCase(eventLocationRepo, log)
	getLocationsUseCase := eventapp.NewGetEventLocationsUseCase(eventLocationRepo, log)
	addArtifactUseCase := eventapp.NewAddArtifactToEventUseCase(eventRepo, artifactRepo, eventArtifactRepo, log)
	removeArtifactUseCase := eventapp.NewRemoveArtifactFromEventUseCase(eventArtifactRepo, log)
	getArtifactsUseCase := eventapp.NewGetEventArtifactsUseCase(eventArtifactRepo, log)
	getStatChangesUseCase := rpgeventapp.NewGetEventStatChangesUseCase(characterStatsRepo, artifactStatsRepo, log)
	handler := NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addCharacterUseCase, removeCharacterUseCase, getCharactersUseCase, addLocationUseCase, removeLocationUseCase, getLocationsUseCase, addArtifactUseCase, removeArtifactUseCase, getArtifactsUseCase, getStatChangesUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Event", "description": "A test event"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if event, ok := resp["event"].(map[string]interface{}); ok {
			if event["name"] != "Test Event" {
				t.Errorf("expected name 'Test Event', got '%v'", event["name"])
			}
		} else {
			t.Error("response missing event")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test Event"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestEventHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)

	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	sceneHandler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, nil, nil, nil, log)

	// Create scene
	sceneBody := `{"story_id": "` + storyID + `", "chapter_id": "` + chapterID + `", "title": "Test Scene", "order_num": 1}`
	sceneReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/scenes", strings.NewReader(sceneBody))
	sceneReq.Header.Set("Content-Type", "application/json")
	sceneReq.Header.Set("X-Tenant-ID", tenantID)
	sceneReq.SetPathValue("id", chapterID)
	sceneW := httptest.NewRecorder()
	withTenantMiddleware(sceneHandler.Create).ServeHTTP(sceneW, sceneReq)

	if sceneW.Code != http.StatusCreated {
		t.Fatalf("failed to create scene: status %d, body: %s", sceneW.Code, sceneW.Body.String())
	}

	var sceneResp map[string]interface{}
	if err := json.NewDecoder(sceneW.Body).Decode(&sceneResp); err != nil {
		t.Fatalf("failed to decode scene response: %v", err)
	}

	sceneObj, ok := sceneResp["scene"].(map[string]interface{})
	if !ok {
		t.Fatalf("scene response missing scene object: %v", sceneResp)
	}

	_, ok = sceneObj["id"].(string)
	if !ok {
		t.Fatalf("scene response missing id: %v", sceneObj)
	}

	eventRepo := postgres.NewEventRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	eventCharacterRepo := postgres.NewEventCharacterRepository(db)
	eventLocationRepo := postgres.NewEventLocationRepository(db)
	eventArtifactRepo := postgres.NewEventArtifactRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(db)
	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(db)

	createEventUseCase := eventapp.NewCreateEventUseCase(eventRepo, worldRepo, auditLogRepo, log)
	getEventUseCase := eventapp.NewGetEventUseCase(eventRepo, log)
	listEventsUseCase := eventapp.NewListEventsUseCase(eventRepo, log)
	updateEventUseCase := eventapp.NewUpdateEventUseCase(eventRepo, auditLogRepo, log)
	deleteEventUseCase := eventapp.NewDeleteEventUseCase(eventRepo, eventCharacterRepo, eventLocationRepo, eventArtifactRepo, auditLogRepo, log)
	addCharacterUseCase := eventapp.NewAddCharacterToEventUseCase(eventRepo, characterRepo, eventCharacterRepo, log)
	removeCharacterUseCase := eventapp.NewRemoveCharacterFromEventUseCase(eventCharacterRepo, log)
	getCharactersUseCase := eventapp.NewGetEventCharactersUseCase(eventCharacterRepo, log)
	addLocationUseCase := eventapp.NewAddLocationToEventUseCase(eventRepo, locationRepo, eventLocationRepo, log)
	removeLocationUseCase := eventapp.NewRemoveLocationFromEventUseCase(eventLocationRepo, log)
	getLocationsUseCase := eventapp.NewGetEventLocationsUseCase(eventLocationRepo, log)
	addArtifactUseCase := eventapp.NewAddArtifactToEventUseCase(eventRepo, artifactRepo, eventArtifactRepo, log)
	removeArtifactUseCase := eventapp.NewRemoveArtifactFromEventUseCase(eventArtifactRepo, log)
	getArtifactsUseCase := eventapp.NewGetEventArtifactsUseCase(eventArtifactRepo, log)
	getStatChangesUseCase := rpgeventapp.NewGetEventStatChangesUseCase(characterStatsRepo, artifactStatsRepo, log)
	handler := NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addCharacterUseCase, removeCharacterUseCase, getCharactersUseCase, addLocationUseCase, removeLocationUseCase, getLocationsUseCase, addArtifactUseCase, removeArtifactUseCase, getArtifactsUseCase, getStatChangesUseCase, log)

	// Create event
	createBody := `{"name": "Test Event"}`
	createReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/events", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Tenant-ID", tenantID)
	createReq.SetPathValue("world_id", worldID)
	createW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("failed to create event: status %d, body: %s", createW.Code, createW.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.NewDecoder(createW.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	eventObj, ok := createResp["event"].(map[string]interface{})
	if !ok {
		t.Fatalf("create response missing event object: %v", createResp)
	}

	eventID, ok := eventObj["id"].(string)
	if !ok {
		t.Fatalf("create response missing id: %v", eventObj)
	}

	t.Run("get existing event", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/events/"+eventID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", eventID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := resp["event"]; !ok {
			t.Error("response missing event")
		}
	})

	t.Run("get non-existing event", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/events/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

