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
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestTraitHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
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

	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	handler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Brave", "category": "Personality", "description": "A brave character trait"}`
		req := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(body))
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

		if trait, ok := resp["trait"].(map[string]interface{}); ok {
			if trait["name"] != "Brave" {
				t.Errorf("expected name 'Brave', got %v", trait["name"])
			}
			if trait["category"] != "Personality" {
				t.Errorf("expected category 'Personality', got %v", trait["category"])
			}
		} else {
			t.Error("response missing trait")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Brave"}`
		req := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(body))
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
		req := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestTraitHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant and trait
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

	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	traitBody := `{"name": "Get Test Trait"}`
	traitReq := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(traitBody))
	traitReq.Header.Set("Content-Type", "application/json")
	traitReq.Header.Set("X-Tenant-ID", tenantID)
	traitW := httptest.NewRecorder()
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	withTenantMiddleware(traitHandler.Create).ServeHTTP(traitW, traitReq)

	if traitW.Code != http.StatusCreated {
		t.Fatalf("failed to create trait: status %d, body: %s", traitW.Code, traitW.Body.String())
	}

	var traitResp map[string]interface{}
	if err := json.NewDecoder(traitW.Body).Decode(&traitResp); err != nil {
		t.Fatalf("failed to decode trait response: %v", err)
	}

	traitObj, ok := traitResp["trait"].(map[string]interface{})
	if !ok {
		t.Fatalf("trait response missing trait object: %v", traitResp)
	}

	traitID, ok := traitObj["id"].(string)
	if !ok {
		t.Fatalf("trait response missing id: %v", traitObj)
	}

	t.Run("existing trait", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/traits/"+traitID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", traitID)
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if trait, ok := resp["trait"].(map[string]interface{}); ok {
			if trait["id"] != traitID {
				t.Errorf("expected ID %s, got %v", traitID, trait["id"])
			}
			if trait["name"] != "Get Test Trait" {
				t.Errorf("expected name 'Get Test Trait', got %v", trait["name"])
			}
		} else {
			t.Error("response missing trait")
		}
	})

	t.Run("non-existing trait", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/traits/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid trait ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/traits/not-a-uuid", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "not-a-uuid")
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestTraitHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
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

	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	handler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)

	// Create multiple traits
	for i := 1; i <= 3; i++ {
		traitBody := `{"name": "Trait ` + strconv.Itoa(i) + `"}`
		traitReq := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(traitBody))
		traitReq.Header.Set("Content-Type", "application/json")
		traitReq.Header.Set("X-Tenant-ID", tenantID)
		traitW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(traitW, traitReq)

		if traitW.Code != http.StatusCreated {
			t.Fatalf("failed to create trait %d: status %d, body: %s", i, traitW.Code, traitW.Body.String())
		}
	}

	t.Run("list traits", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/traits?limit=10&offset=0", nil)
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

		if traits, ok := resp["traits"].([]interface{}); ok {
			if len(traits) < 3 {
				t.Errorf("expected at least 3 traits, got %d", len(traits))
			}
		} else {
			t.Error("response missing traits")
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
		req := httptest.NewRequest("GET", "/api/v1/traits", nil)
		// X-Tenant-ID header not set
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		// Middleware returns 400 (ValidationError) when X-Tenant-ID is missing
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestTraitHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant and trait
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

	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	traitBody := `{"name": "Original Trait"}`
	traitReq := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(traitBody))
	traitReq.Header.Set("Content-Type", "application/json")
	traitReq.Header.Set("X-Tenant-ID", tenantID)
	traitW := httptest.NewRecorder()
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	withTenantMiddleware(traitHandler.Create).ServeHTTP(traitW, traitReq)

	if traitW.Code != http.StatusCreated {
		t.Fatalf("failed to create trait: status %d, body: %s", traitW.Code, traitW.Body.String())
	}

	var traitResp map[string]interface{}
	if err := json.NewDecoder(traitW.Body).Decode(&traitResp); err != nil {
		t.Fatalf("failed to decode trait response: %v", err)
	}

	traitObj, ok := traitResp["trait"].(map[string]interface{})
	if !ok {
		t.Fatalf("trait response missing trait object: %v", traitResp)
	}

	traitID, ok := traitObj["id"].(string)
	if !ok {
		t.Fatalf("trait response missing id: %v", traitObj)
	}

	t.Run("update trait name", func(t *testing.T) {
		body := `{"name": "Updated Trait"}`
		req := httptest.NewRequest("PUT", "/api/v1/traits/"+traitID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", traitID)
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if trait, ok := resp["trait"].(map[string]interface{}); ok {
			if trait["name"] != "Updated Trait" {
				t.Errorf("expected name 'Updated Trait', got %v", trait["name"])
			}
			if trait["id"] != traitID {
				t.Errorf("expected ID %s, got %v", traitID, trait["id"])
			}
		} else {
			t.Error("response missing trait")
		}
	})

	t.Run("non-existing trait", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/traits/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestTraitHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	// Create a tenant and trait
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

	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	traitBody := `{"name": "Trait to Delete"}`
	traitReq := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(traitBody))
	traitReq.Header.Set("Content-Type", "application/json")
	traitReq.Header.Set("X-Tenant-ID", tenantID)
	traitW := httptest.NewRecorder()
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	withTenantMiddleware(traitHandler.Create).ServeHTTP(traitW, traitReq)

	if traitW.Code != http.StatusCreated {
		t.Fatalf("failed to create trait: status %d, body: %s", traitW.Code, traitW.Body.String())
	}

	var traitResp map[string]interface{}
	if err := json.NewDecoder(traitW.Body).Decode(&traitResp); err != nil {
		t.Fatalf("failed to decode trait response: %v", err)
	}

	traitObj, ok := traitResp["trait"].(map[string]interface{})
	if !ok {
		t.Fatalf("trait response missing trait object: %v", traitResp)
	}

	traitID, ok := traitObj["id"].(string)
	if !ok {
		t.Fatalf("trait response missing id: %v", traitObj)
	}

	t.Run("delete existing trait", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/traits/"+traitID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", traitID)
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify trait is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/traits/"+traitID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", traitID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted trait, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing trait", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/traits/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(traitHandler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

