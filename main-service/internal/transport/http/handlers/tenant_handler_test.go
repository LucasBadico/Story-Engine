//go:build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestTenantHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	if err := postgres.TruncateTables(nil, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := postgres.NewTenantRepository(postgres.NewDB(db))
	auditLogRepo := postgres.NewAuditLogRepository(postgres.NewDB(db))
	log := logger.New()

	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	handler := NewTenantHandler(createTenantUseCase, tenantRepo, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Tenant"}`
		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if tenant, ok := resp["tenant"].(map[string]interface{}); ok {
			if tenant["name"] != "Test Tenant" {
				t.Errorf("expected name 'Test Tenant', got %v", tenant["name"])
			}
		} else {
			t.Error("response missing tenant")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		body := `{"name": "Duplicate Tenant"}`
		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("unexpected error on first creation: status %d", w.Code)
		}

		// Try to create duplicate
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		handler.Create(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w2.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(w2.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errResp.Code != "ALREADY_EXISTS" {
			t.Errorf("expected code ALREADY_EXISTS, got %s", errResp.Code)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		body := `{"name": invalid}`
		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestTenantHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	if err := postgres.TruncateTables(nil, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := postgres.NewTenantRepository(postgres.NewDB(db))
	auditLogRepo := postgres.NewAuditLogRepository(postgres.NewDB(db))
	log := logger.New()

	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	handler := NewTenantHandler(createTenantUseCase, tenantRepo, log)

	// Create a tenant first
	createBody := `{"name": "Get Test Tenant"}`
	createReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.Create(createW, createReq)

	var createResp map[string]interface{}
	json.NewDecoder(createW.Body).Decode(&createResp)
	tenantID := createResp["tenant"].(map[string]interface{})["id"].(string)

	t.Run("existing tenant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/tenants/"+tenantID, nil)
		req.SetPathValue("id", tenantID)
		w := httptest.NewRecorder()

		handler.Get(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if tenant, ok := resp["tenant"].(map[string]interface{}); ok {
			if tenant["id"] != tenantID {
				t.Errorf("expected ID %s, got %v", tenantID, tenant["id"])
			}
		} else {
			t.Error("response missing tenant")
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/tenants/invalid-uuid", nil)
		req.SetPathValue("id", "invalid-uuid")
		w := httptest.NewRecorder()

		handler.Get(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("non-existing tenant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/tenants/00000000-0000-0000-0000-000000000000", nil)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		handler.Get(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

