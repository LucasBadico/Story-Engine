//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestArtifactHandler_Create(t *testing.T) {
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
	getReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	handler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getReferencesUseCase, addReferenceUseCase, removeReferenceUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Artifact", "description": "A test artifact", "rarity": "common"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(body))
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

		if artifact, ok := resp["artifact"].(map[string]interface{}); ok {
			if artifact["name"] != "Test Artifact" {
				t.Errorf("expected name 'Test Artifact', got %v", artifact["name"])
			}
		} else {
			t.Error("response missing artifact")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test Artifact"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestArtifactHandler_Get(t *testing.T) {
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
	getReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	handler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getReferencesUseCase, addReferenceUseCase, removeReferenceUseCase, log)

	// Create artifact
	artifactBody := `{"name": "Get Test Artifact"}`
	artifactReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(artifactBody))
	artifactReq.Header.Set("Content-Type", "application/json")
	artifactReq.Header.Set("X-Tenant-ID", tenantID)
	artifactReq.SetPathValue("world_id", worldID)
	artifactW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(artifactW, artifactReq)

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

	t.Run("existing artifact", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/artifacts/"+artifactID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", artifactID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if artifact, ok := resp["artifact"].(map[string]interface{}); ok {
			if artifact["id"] != artifactID {
				t.Errorf("expected ID %s, got %v", artifactID, artifact["id"])
			}
		} else {
			t.Error("response missing artifact")
		}
	})

	t.Run("non-existing artifact", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/artifacts/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestArtifactHandler_List(t *testing.T) {
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
	getReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	handler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getReferencesUseCase, addReferenceUseCase, removeReferenceUseCase, log)

	// Create multiple artifacts
	for i := 1; i <= 3; i++ {
		artifactBody := `{"name": "Artifact ` + strconv.Itoa(i) + `"}`
		artifactReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(artifactBody))
		artifactReq.Header.Set("Content-Type", "application/json")
		artifactReq.Header.Set("X-Tenant-ID", tenantID)
		artifactReq.SetPathValue("world_id", worldID)
		artifactW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(artifactW, artifactReq)

		if artifactW.Code != http.StatusCreated {
			t.Fatalf("failed to create artifact %d: status %d, body: %s", i, artifactW.Code, artifactW.Body.String())
		}
	}

	t.Run("list artifacts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID+"/artifacts?limit=10&offset=0", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if artifacts, ok := resp["artifacts"].([]interface{}); ok {
			if len(artifacts) < 3 {
				t.Errorf("expected at least 3 artifacts, got %d", len(artifacts))
			}
		} else {
			t.Error("response missing artifacts")
		}
	})
}

func TestArtifactHandler_Update(t *testing.T) {
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
	getReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	handler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getReferencesUseCase, addReferenceUseCase, removeReferenceUseCase, log)

	// Create artifact
	artifactBody := `{"name": "Original Artifact"}`
	artifactReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(artifactBody))
	artifactReq.Header.Set("Content-Type", "application/json")
	artifactReq.Header.Set("X-Tenant-ID", tenantID)
	artifactReq.SetPathValue("world_id", worldID)
	artifactW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(artifactW, artifactReq)

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

	t.Run("update artifact name", func(t *testing.T) {
		body := `{"name": "Updated Artifact"}`
		req := httptest.NewRequest("PUT", "/api/v1/artifacts/"+artifactID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", artifactID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if artifact, ok := resp["artifact"].(map[string]interface{}); ok {
			if artifact["name"] != "Updated Artifact" {
				t.Errorf("expected name 'Updated Artifact', got %v", artifact["name"])
			}
		} else {
			t.Error("response missing artifact")
		}
	})

	t.Run("non-existing artifact", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/artifacts/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestArtifactHandler_Delete(t *testing.T) {
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
	getReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	handler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getReferencesUseCase, addReferenceUseCase, removeReferenceUseCase, log)

	// Create artifact
	artifactBody := `{"name": "Artifact to Delete"}`
	artifactReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/artifacts", strings.NewReader(artifactBody))
	artifactReq.Header.Set("Content-Type", "application/json")
	artifactReq.Header.Set("X-Tenant-ID", tenantID)
	artifactReq.SetPathValue("world_id", worldID)
	artifactW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(artifactW, artifactReq)

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

	t.Run("delete existing artifact", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/artifacts/"+artifactID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", artifactID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify artifact is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/artifacts/"+artifactID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", artifactID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted artifact, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing artifact", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/artifacts/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

