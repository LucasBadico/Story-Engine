//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestArtifactRPGStatsHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	artifactRepo := postgres.NewArtifactRepository(db)
	artifactReferenceRepo := postgres.NewArtifactReferenceRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	artifactHandler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)

	// Create artifact
	artifactBody := `{"name": "Test Artifact"}`
	artifactReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(artifactBody))
	artifactReq.Header.Set("Content-Type", "application/json")
	artifactReq.Header.Set("X-Tenant-ID", tenantID)
	artifactReq.SetPathValue("world_id", worldID)
	artifactW := httptest.NewRecorder()
	withTenantMiddleware(artifactHandler.Create).ServeHTTP(artifactW, artifactReq)

	if artifactW.Code != http.StatusCreated {
		t.Fatalf("failed to create artifact: status %d, body: %s", artifactW.Code, artifactW.Body.String())
	}

	var artifactResp map[string]interface{}
	if err := json.NewDecoder(artifactW.Body).Decode(&artifactResp); err != nil {
		t.Fatalf("failed to decode artifact response: %v", err)
	}

	artifactObj, ok := artifactResp["artifact"].(map[string]interface{})
	if !ok {
		t.Fatalf("artifact response missing artifact object: %v", artifactResp)
	}

	artifactID, ok := artifactObj["id"].(string)
	if !ok {
		t.Fatalf("artifact response missing id: %v", artifactObj)
	}

	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	createStatsUseCase := artifactstatsapp.NewCreateArtifactStatsUseCase(artifactStatsRepo, artifactRepo, eventRepo, log)
	getActiveStatsUseCase := artifactstatsapp.NewGetActiveArtifactStatsUseCase(artifactStatsRepo, log)
	listHistoryUseCase := artifactstatsapp.NewListArtifactStatsHistoryUseCase(artifactStatsRepo, log)
	activateVersionUseCase := artifactstatsapp.NewActivateArtifactStatsVersionUseCase(artifactStatsRepo, log)
	handler := NewArtifactRPGStatsHandler(createStatsUseCase, getActiveStatsUseCase, listHistoryUseCase, activateVersionUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"stats": {"strength": 5, "dexterity": 3}}`
		req := httptest.NewRequest("POST", "/api/v1/artifacts/"+artifactID+"/rpg-stats", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", artifactID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if stats, ok := resp["stats"].(map[string]interface{}); ok {
			if stats["artifact_id"] != artifactID {
				t.Errorf("expected artifact_id %s, got %v", artifactID, stats["artifact_id"])
			}
		} else {
			t.Error("response missing stats")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"stats": {"strength": 5}}`
		req := httptest.NewRequest("POST", "/api/v1/artifacts/"+artifactID+"/rpg-stats", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", artifactID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestArtifactRPGStatsHandler_GetActive(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	artifactRepo := postgres.NewArtifactRepository(db)
	artifactReferenceRepo := postgres.NewArtifactReferenceRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	artifactHandler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)

	// Create artifact
	artifactBody := `{"name": "Test Artifact"}`
	artifactReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(artifactBody))
	artifactReq.Header.Set("Content-Type", "application/json")
	artifactReq.Header.Set("X-Tenant-ID", tenantID)
	artifactReq.SetPathValue("world_id", worldID)
	artifactW := httptest.NewRecorder()
	withTenantMiddleware(artifactHandler.Create).ServeHTTP(artifactW, artifactReq)

	if artifactW.Code != http.StatusCreated {
		t.Fatalf("failed to create artifact: status %d, body: %s", artifactW.Code, artifactW.Body.String())
	}

	var artifactResp map[string]interface{}
	if err := json.NewDecoder(artifactW.Body).Decode(&artifactResp); err != nil {
		t.Fatalf("failed to decode artifact response: %v", err)
	}

	artifactObj, ok := artifactResp["artifact"].(map[string]interface{})
	if !ok {
		t.Fatalf("artifact response missing artifact object: %v", artifactResp)
	}

	artifactID, ok := artifactObj["id"].(string)
	if !ok {
		t.Fatalf("artifact response missing id: %v", artifactObj)
	}

	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	createStatsUseCase := artifactstatsapp.NewCreateArtifactStatsUseCase(artifactStatsRepo, artifactRepo, eventRepo, log)
	getActiveStatsUseCase := artifactstatsapp.NewGetActiveArtifactStatsUseCase(artifactStatsRepo, log)
	listHistoryUseCase := artifactstatsapp.NewListArtifactStatsHistoryUseCase(artifactStatsRepo, log)
	activateVersionUseCase := artifactstatsapp.NewActivateArtifactStatsVersionUseCase(artifactStatsRepo, log)
	handler := NewArtifactRPGStatsHandler(createStatsUseCase, getActiveStatsUseCase, listHistoryUseCase, activateVersionUseCase, log)

	// Create stats
	createBody := `{"stats": {"strength": 5, "dexterity": 3}}`
	createReq := httptest.NewRequest("POST", "/api/v1/artifacts/"+artifactID+"/rpg-stats", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Tenant-ID", tenantID)
	createReq.SetPathValue("id", artifactID)
	createW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("failed to create stats: status %d, body: %s", createW.Code, createW.Body.String())
	}

	t.Run("get active stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/artifacts/"+artifactID+"/rpg-stats", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", artifactID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.GetActive).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := resp["stats"]; !ok {
			t.Error("response missing stats")
		}
	})
}

