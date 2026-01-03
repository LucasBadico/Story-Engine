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
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestWorldHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant first
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	handler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test World", "description": "A test world", "genre": "Fantasy"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(body))
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

		if world, ok := resp["world"].(map[string]interface{}); ok {
			if world["name"] != "Test World" {
				t.Errorf("expected name 'Test World', got %v", world["name"])
			}
			if world["description"] != "A test world" {
				t.Errorf("expected description 'A test world', got %v", world["description"])
			}
			if world["genre"] != "Fantasy" {
				t.Errorf("expected genre 'Fantasy', got %v", world["genre"])
			}
		} else {
			t.Error("response missing world")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test World"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// X-Tenant-ID header not set
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		// Middleware returns 400 (ValidationError) when X-Tenant-ID is missing
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestWorldHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant and world
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	worldBody := `{"name": "Get Test World"}`
	worldReq := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(worldBody))
	worldReq.Header.Set("Content-Type", "application/json")
	worldReq.Header.Set("X-Tenant-ID", tenantID)
	worldW := httptest.NewRecorder()
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	withTenantMiddleware(worldHandler.Create).ServeHTTP(worldW, worldReq)

	if worldW.Code != http.StatusCreated {
		t.Fatalf("failed to create world: status %d, body: %s", worldW.Code, worldW.Body.String())
	}

	var worldResp map[string]interface{}
	if err := json.NewDecoder(worldW.Body).Decode(&worldResp); err != nil {
		t.Fatalf("failed to decode world response: %v", err)
	}

	worldObj, ok := worldResp["world"].(map[string]interface{})
	if !ok {
		t.Fatalf("world response missing world object: %v", worldResp)
	}

	worldID, ok := worldObj["id"].(string)
	if !ok {
		t.Fatalf("world response missing id: %v", worldObj)
	}

	t.Run("existing world", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if world, ok := resp["world"].(map[string]interface{}); ok {
			if world["id"] != worldID {
				t.Errorf("expected ID %s, got %v", worldID, world["id"])
			}
			if world["name"] != "Get Test World" {
				t.Errorf("expected name 'Get Test World', got %v", world["name"])
			}
		} else {
			t.Error("response missing world")
		}
	})

	t.Run("non-existing world", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid world ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/not-a-uuid", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "not-a-uuid")
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestWorldHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	handler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)

	// Create multiple worlds
	for i := 1; i <= 3; i++ {
		worldBody := `{"name": "World ` + strconv.Itoa(i) + `"}`
		worldReq := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(worldBody))
		worldReq.Header.Set("Content-Type", "application/json")
		worldReq.Header.Set("X-Tenant-ID", tenantID)
		worldW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(worldW, worldReq)

		if worldW.Code != http.StatusCreated {
			t.Fatalf("failed to create world %d: status %d, body: %s", i, worldW.Code, worldW.Body.String())
		}
	}

	t.Run("list worlds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds?limit=10&offset=0", nil)
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

		if worlds, ok := resp["worlds"].([]interface{}); ok {
			if len(worlds) < 3 {
				t.Errorf("expected at least 3 worlds, got %d", len(worlds))
			}
		} else {
			t.Error("response missing worlds")
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
		req := httptest.NewRequest("GET", "/api/v1/worlds", nil)
		// X-Tenant-ID header not set
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		// Middleware returns 400 (ValidationError) when X-Tenant-ID is missing
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestWorldHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant and world
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	worldBody := `{"name": "Original World"}`
	worldReq := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(worldBody))
	worldReq.Header.Set("Content-Type", "application/json")
	worldReq.Header.Set("X-Tenant-ID", tenantID)
	worldW := httptest.NewRecorder()
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	withTenantMiddleware(worldHandler.Create).ServeHTTP(worldW, worldReq)

	if worldW.Code != http.StatusCreated {
		t.Fatalf("failed to create world: status %d, body: %s", worldW.Code, worldW.Body.String())
	}

	var worldResp map[string]interface{}
	if err := json.NewDecoder(worldW.Body).Decode(&worldResp); err != nil {
		t.Fatalf("failed to decode world response: %v", err)
	}

	worldObj, ok := worldResp["world"].(map[string]interface{})
	if !ok {
		t.Fatalf("world response missing world object: %v", worldResp)
	}

	worldID, ok := worldObj["id"].(string)
	if !ok {
		t.Fatalf("world response missing id: %v", worldObj)
	}

	t.Run("update world name", func(t *testing.T) {
		body := `{"name": "Updated World"}`
		req := httptest.NewRequest("PUT", "/api/v1/worlds/"+worldID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if world, ok := resp["world"].(map[string]interface{}); ok {
			if world["name"] != "Updated World" {
				t.Errorf("expected name 'Updated World', got %v", world["name"])
			}
			if world["id"] != worldID {
				t.Errorf("expected ID %s, got %v", worldID, world["id"])
			}
		} else {
			t.Error("response missing world")
		}
	})

	t.Run("non-existing world", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/worlds/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestWorldHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant and world
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	worldBody := `{"name": "World to Delete"}`
	worldReq := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(worldBody))
	worldReq.Header.Set("Content-Type", "application/json")
	worldReq.Header.Set("X-Tenant-ID", tenantID)
	worldW := httptest.NewRecorder()
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	withTenantMiddleware(worldHandler.Create).ServeHTTP(worldW, worldReq)

	if worldW.Code != http.StatusCreated {
		t.Fatalf("failed to create world: status %d, body: %s", worldW.Code, worldW.Body.String())
	}

	var worldResp map[string]interface{}
	if err := json.NewDecoder(worldW.Body).Decode(&worldResp); err != nil {
		t.Fatalf("failed to decode world response: %v", err)
	}

	worldObj, ok := worldResp["world"].(map[string]interface{})
	if !ok {
		t.Fatalf("world response missing world object: %v", worldResp)
	}

	worldID, ok := worldObj["id"].(string)
	if !ok {
		t.Fatalf("world response missing id: %v", worldObj)
	}

	t.Run("delete existing world", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/worlds/"+worldID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify world is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", worldID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted world, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing world", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/worlds/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(worldHandler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

