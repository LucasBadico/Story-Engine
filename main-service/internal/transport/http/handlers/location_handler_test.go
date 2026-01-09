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
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestLocationHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	locationRepo := postgres.NewLocationRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	handler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Location", "type": "City", "description": "A test location"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(body))
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

		if location, ok := resp["location"].(map[string]interface{}); ok {
			if location["name"] != "Test Location" {
				t.Errorf("expected name 'Test Location', got %v", location["name"])
			}
		} else {
			t.Error("response missing location")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test Location"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(body))
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
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(body))
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

func TestLocationHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	locationRepo := postgres.NewLocationRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	handler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)

	// Create location
	locationBody := `{"name": "Get Test Location"}`
	locationReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(locationBody))
	locationReq.Header.Set("Content-Type", "application/json")
	locationReq.Header.Set("X-Tenant-ID", tenantID)
	locationReq.SetPathValue("world_id", worldID)
	locationW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(locationW, locationReq)

	if locationW.Code != http.StatusCreated {
		t.Fatalf("failed to create location: status %d, body: %s", locationW.Code, locationW.Body.String())
	}

	var locationResp map[string]interface{}
	if err := json.NewDecoder(locationW.Body).Decode(&locationResp); err != nil {
		t.Fatalf("failed to decode location response: %v", err)
	}

	locationObj, ok := locationResp["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("location response missing location object: %v", locationResp)
	}

	locationID, ok := locationObj["id"].(string)
	if !ok {
		t.Fatalf("location response missing id: %v", locationObj)
	}

	t.Run("existing location", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/locations/"+locationID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", locationID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if location, ok := resp["location"].(map[string]interface{}); ok {
			if location["id"] != locationID {
				t.Errorf("expected ID %s, got %v", locationID, location["id"])
			}
		} else {
			t.Error("response missing location")
		}
	})

	t.Run("non-existing location", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/locations/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid location ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/locations/not-a-uuid", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "not-a-uuid")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestLocationHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	locationRepo := postgres.NewLocationRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	handler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)

	// Create multiple locations
	for i := 1; i <= 3; i++ {
		locationBody := `{"name": "Location ` + strconv.Itoa(i) + `"}`
		locationReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(locationBody))
		locationReq.Header.Set("Content-Type", "application/json")
		locationReq.Header.Set("X-Tenant-ID", tenantID)
		locationReq.SetPathValue("world_id", worldID)
		locationW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(locationW, locationReq)

		if locationW.Code != http.StatusCreated {
			t.Fatalf("failed to create location %d: status %d, body: %s", i, locationW.Code, locationW.Body.String())
		}
	}

	t.Run("list locations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID+"/locations?limit=10&offset=0", nil)
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

		if locations, ok := resp["locations"].([]interface{}); ok {
			if len(locations) < 3 {
				t.Errorf("expected at least 3 locations, got %d", len(locations))
			}
		} else {
			t.Error("response missing locations")
		}

		if total, ok := resp["total"].(float64); ok {
			if total < 3 {
				t.Errorf("expected total at least 3, got %v", total)
			}
		} else {
			t.Error("response missing total")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID+"/locations", nil)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestLocationHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	locationRepo := postgres.NewLocationRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	handler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)

	// Create location
	locationBody := `{"name": "Original Location"}`
	locationReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(locationBody))
	locationReq.Header.Set("Content-Type", "application/json")
	locationReq.Header.Set("X-Tenant-ID", tenantID)
	locationReq.SetPathValue("world_id", worldID)
	locationW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(locationW, locationReq)

	if locationW.Code != http.StatusCreated {
		t.Fatalf("failed to create location: status %d, body: %s", locationW.Code, locationW.Body.String())
	}

	var locationResp map[string]interface{}
	if err := json.NewDecoder(locationW.Body).Decode(&locationResp); err != nil {
		t.Fatalf("failed to decode location response: %v", err)
	}

	locationObj, ok := locationResp["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("location response missing location object: %v", locationResp)
	}

	locationID, ok := locationObj["id"].(string)
	if !ok {
		t.Fatalf("location response missing id: %v", locationObj)
	}

	t.Run("update location name", func(t *testing.T) {
		body := `{"name": "Updated Location"}`
		req := httptest.NewRequest("PUT", "/api/v1/locations/"+locationID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", locationID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if location, ok := resp["location"].(map[string]interface{}); ok {
			if location["name"] != "Updated Location" {
				t.Errorf("expected name 'Updated Location', got %v", location["name"])
			}
		} else {
			t.Error("response missing location")
		}
	})

	t.Run("non-existing location", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/locations/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestLocationHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	locationRepo := postgres.NewLocationRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	handler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)

	// Create location
	locationBody := `{"name": "Location to Delete"}`
	locationReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/locations", strings.NewReader(locationBody))
	locationReq.Header.Set("Content-Type", "application/json")
	locationReq.Header.Set("X-Tenant-ID", tenantID)
	locationReq.SetPathValue("world_id", worldID)
	locationW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(locationW, locationReq)

	if locationW.Code != http.StatusCreated {
		t.Fatalf("failed to create location: status %d, body: %s", locationW.Code, locationW.Body.String())
	}

	var locationResp map[string]interface{}
	if err := json.NewDecoder(locationW.Body).Decode(&locationResp); err != nil {
		t.Fatalf("failed to decode location response: %v", err)
	}

	locationObj, ok := locationResp["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("location response missing location object: %v", locationResp)
	}

	locationID, ok := locationObj["id"].(string)
	if !ok {
		t.Fatalf("location response missing id: %v", locationObj)
	}

	t.Run("delete existing location", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/locations/"+locationID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", locationID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify location is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/locations/"+locationID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", locationID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted location, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing location", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/locations/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

