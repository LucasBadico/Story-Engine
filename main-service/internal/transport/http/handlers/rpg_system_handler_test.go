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
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestRPGSystemHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	log := logger.New()

	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	handler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test RPG System", "base_stats_schema": {"strength": "int", "dexterity": "int"}}`
		req := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgSystem, ok := resp["rpg_system"].(map[string]interface{}); ok {
			if rpgSystem["name"] != "Test RPG System" {
				t.Errorf("expected name 'Test RPG System', got %v", rpgSystem["name"])
			}
		} else {
			t.Error("response missing rpg_system")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test RPG System", "base_stats_schema": {}}`
		req := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestRPGSystemHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	log := logger.New()

	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	handler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Get Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

	if rpgSystemW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG system: status %d, body: %s", rpgSystemW.Code, rpgSystemW.Body.String())
	}

	var rpgSystemResp map[string]interface{}
	if err := json.NewDecoder(rpgSystemW.Body).Decode(&rpgSystemResp); err != nil {
		t.Fatalf("failed to decode RPG system response: %v", err)
	}

	rpgSystemObj, ok := rpgSystemResp["rpg_system"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG system response missing rpg_system object: %v", rpgSystemResp)
	}

	rpgSystemID, ok := rpgSystemObj["id"].(string)
	if !ok {
		t.Fatalf("RPG system response missing id: %v", rpgSystemObj)
	}

	t.Run("existing rpg_system", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-systems/"+rpgSystemID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgSystem, ok := resp["rpg_system"].(map[string]interface{}); ok {
			if rpgSystem["id"] != rpgSystemID {
				t.Errorf("expected ID %s, got %v", rpgSystemID, rpgSystem["id"])
			}
		} else {
			t.Error("response missing rpg_system")
		}
	})

	t.Run("non-existing rpg_system", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-systems/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestRPGSystemHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	log := logger.New()

	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	handler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create multiple RPG systems
	for i := 1; i <= 3; i++ {
		rpgSystemBody := `{"name": "RPG System ` + strconv.Itoa(i) + `", "base_stats_schema": {}}`
		rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
		rpgSystemReq.Header.Set("Content-Type", "application/json")
		rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
		rpgSystemW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

		if rpgSystemW.Code != http.StatusCreated {
			t.Fatalf("failed to create RPG system %d: status %d, body: %s", i, rpgSystemW.Code, rpgSystemW.Body.String())
		}
	}

	t.Run("list rpg_systems", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-systems?limit=10&offset=0", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgSystems, ok := resp["rpg_systems"].([]interface{}); ok {
			if len(rpgSystems) < 3 {
				t.Errorf("expected at least 3 RPG systems, got %d", len(rpgSystems))
			}
		} else {
			t.Error("response missing rpg_systems")
		}
	})
}

func TestRPGSystemHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	log := logger.New()

	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	handler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Original RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

	if rpgSystemW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG system: status %d, body: %s", rpgSystemW.Code, rpgSystemW.Body.String())
	}

	var rpgSystemResp map[string]interface{}
	if err := json.NewDecoder(rpgSystemW.Body).Decode(&rpgSystemResp); err != nil {
		t.Fatalf("failed to decode RPG system response: %v", err)
	}

	rpgSystemObj, ok := rpgSystemResp["rpg_system"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG system response missing rpg_system object: %v", rpgSystemResp)
	}

	rpgSystemID, ok := rpgSystemObj["id"].(string)
	if !ok {
		t.Fatalf("RPG system response missing id: %v", rpgSystemObj)
	}

	t.Run("update rpg_system name", func(t *testing.T) {
		body := `{"name": "Updated RPG System"}`
		req := httptest.NewRequest("PUT", "/api/v1/rpg-systems/"+rpgSystemID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgSystem, ok := resp["rpg_system"].(map[string]interface{}); ok {
			if rpgSystem["name"] != "Updated RPG System" {
				t.Errorf("expected name 'Updated RPG System', got %v", rpgSystem["name"])
			}
		} else {
			t.Error("response missing rpg_system")
		}
	})

	t.Run("non-existing rpg_system", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/rpg-systems/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestRPGSystemHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	log := logger.New()

	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	handler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "RPG System to Delete", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

	if rpgSystemW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG system: status %d, body: %s", rpgSystemW.Code, rpgSystemW.Body.String())
	}

	var rpgSystemResp map[string]interface{}
	if err := json.NewDecoder(rpgSystemW.Body).Decode(&rpgSystemResp); err != nil {
		t.Fatalf("failed to decode RPG system response: %v", err)
	}

	rpgSystemObj, ok := rpgSystemResp["rpg_system"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG system response missing rpg_system object: %v", rpgSystemResp)
	}

	rpgSystemID, ok := rpgSystemObj["id"].(string)
	if !ok {
		t.Fatalf("RPG system response missing id: %v", rpgSystemObj)
	}

	t.Run("delete existing rpg_system", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/rpg-systems/"+rpgSystemID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify RPG system is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/rpg-systems/"+rpgSystemID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", rpgSystemID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted RPG system, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing rpg_system", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/rpg-systems/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

