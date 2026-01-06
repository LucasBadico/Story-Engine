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
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestArchetypeHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Warrior", "description": "A warrior archetype"}`
		req := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(body))
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

		if archetype, ok := resp["archetype"].(map[string]interface{}); ok {
			if archetype["name"] != "Warrior" {
				t.Errorf("expected name 'Warrior', got %v", archetype["name"])
			}
		} else {
			t.Error("response missing archetype")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Warrior"}`
		req := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestArchetypeHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	// Create archetype
	archetypeBody := `{"name": "Get Test Archetype"}`
	archetypeReq := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(archetypeBody))
	archetypeReq.Header.Set("Content-Type", "application/json")
	archetypeReq.Header.Set("X-Tenant-ID", tenantID)
	archetypeW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(archetypeW, archetypeReq)

	if archetypeW.Code != http.StatusCreated {
		t.Fatalf("failed to create archetype: status %d, body: %s", archetypeW.Code, archetypeW.Body.String())
	}

	var archetypeResp map[string]interface{}
	if err := json.NewDecoder(archetypeW.Body).Decode(&archetypeResp); err != nil {
		t.Fatalf("failed to decode archetype response: %v", err)
	}

	archetypeObj, ok := archetypeResp["archetype"].(map[string]interface{})
	if !ok {
		t.Fatalf("archetype response missing archetype object: %v", archetypeResp)
	}

	archetypeID, ok := archetypeObj["id"].(string)
	if !ok {
		t.Fatalf("archetype response missing id: %v", archetypeObj)
	}

	t.Run("existing archetype", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/archetypes/"+archetypeID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", archetypeID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if archetype, ok := resp["archetype"].(map[string]interface{}); ok {
			if archetype["id"] != archetypeID {
				t.Errorf("expected ID %s, got %v", archetypeID, archetype["id"])
			}
		} else {
			t.Error("response missing archetype")
		}
	})

	t.Run("non-existing archetype", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/archetypes/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestArchetypeHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	// Create multiple archetypes
	for i := 1; i <= 3; i++ {
		archetypeBody := `{"name": "Archetype ` + strconv.Itoa(i) + `"}`
		archetypeReq := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(archetypeBody))
		archetypeReq.Header.Set("Content-Type", "application/json")
		archetypeReq.Header.Set("X-Tenant-ID", tenantID)
		archetypeW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(archetypeW, archetypeReq)

		if archetypeW.Code != http.StatusCreated {
			t.Fatalf("failed to create archetype %d: status %d, body: %s", i, archetypeW.Code, archetypeW.Body.String())
		}
	}

	t.Run("list archetypes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/archetypes?limit=10&offset=0", nil)
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

		if archetypes, ok := resp["archetypes"].([]interface{}); ok {
			if len(archetypes) < 3 {
				t.Errorf("expected at least 3 archetypes, got %d", len(archetypes))
			}
		} else {
			t.Error("response missing archetypes")
		}
	})
}

func TestArchetypeHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	// Create archetype
	archetypeBody := `{"name": "Original Archetype"}`
	archetypeReq := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(archetypeBody))
	archetypeReq.Header.Set("Content-Type", "application/json")
	archetypeReq.Header.Set("X-Tenant-ID", tenantID)
	archetypeW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(archetypeW, archetypeReq)

	if archetypeW.Code != http.StatusCreated {
		t.Fatalf("failed to create archetype: status %d, body: %s", archetypeW.Code, archetypeW.Body.String())
	}

	var archetypeResp map[string]interface{}
	if err := json.NewDecoder(archetypeW.Body).Decode(&archetypeResp); err != nil {
		t.Fatalf("failed to decode archetype response: %v", err)
	}

	archetypeObj, ok := archetypeResp["archetype"].(map[string]interface{})
	if !ok {
		t.Fatalf("archetype response missing archetype object: %v", archetypeResp)
	}

	archetypeID, ok := archetypeObj["id"].(string)
	if !ok {
		t.Fatalf("archetype response missing id: %v", archetypeObj)
	}

	t.Run("update archetype", func(t *testing.T) {
		body := `{"name": "Updated Archetype"}`
		req := httptest.NewRequest("PUT", "/api/v1/archetypes/"+archetypeID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", archetypeID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if archetype, ok := resp["archetype"].(map[string]interface{}); ok {
			if archetype["name"] != "Updated Archetype" {
				t.Errorf("expected name 'Updated Archetype', got %v", archetype["name"])
			}
		} else {
			t.Error("response missing archetype")
		}
	})
}

func TestArchetypeHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	// Create archetype
	archetypeBody := `{"name": "Archetype to Delete"}`
	archetypeReq := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(archetypeBody))
	archetypeReq.Header.Set("Content-Type", "application/json")
	archetypeReq.Header.Set("X-Tenant-ID", tenantID)
	archetypeW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(archetypeW, archetypeReq)

	if archetypeW.Code != http.StatusCreated {
		t.Fatalf("failed to create archetype: status %d, body: %s", archetypeW.Code, archetypeW.Body.String())
	}

	var archetypeResp map[string]interface{}
	if err := json.NewDecoder(archetypeW.Body).Decode(&archetypeResp); err != nil {
		t.Fatalf("failed to decode archetype response: %v", err)
	}

	archetypeObj, ok := archetypeResp["archetype"].(map[string]interface{})
	if !ok {
		t.Fatalf("archetype response missing archetype object: %v", archetypeResp)
	}

	archetypeID, ok := archetypeObj["id"].(string)
	if !ok {
		t.Fatalf("archetype response missing id: %v", archetypeObj)
	}

	t.Run("delete archetype", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/archetypes/"+archetypeID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", archetypeID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})
}

func TestArchetypeHandler_AddTrait(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	// Create trait handler for trait creation
	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)

	// Create archetype
	archetypeBody := `{"name": "Warrior"}`
	archetypeReq := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(archetypeBody))
	archetypeReq.Header.Set("Content-Type", "application/json")
	archetypeReq.Header.Set("X-Tenant-ID", tenantID)
	archetypeW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(archetypeW, archetypeReq)

	if archetypeW.Code != http.StatusCreated {
		t.Fatalf("failed to create archetype: status %d, body: %s", archetypeW.Code, archetypeW.Body.String())
	}

	var archetypeResp map[string]interface{}
	if err := json.NewDecoder(archetypeW.Body).Decode(&archetypeResp); err != nil {
		t.Fatalf("failed to decode archetype response: %v", err)
	}

	archetypeObj, ok := archetypeResp["archetype"].(map[string]interface{})
	if !ok {
		t.Fatalf("archetype response missing archetype object: %v", archetypeResp)
	}

	archetypeID, ok := archetypeObj["id"].(string)
	if !ok {
		t.Fatalf("archetype response missing id: %v", archetypeObj)
	}

	// Create trait
	traitBody := `{"name": "Brave"}`
	traitReq := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(traitBody))
	traitReq.Header.Set("Content-Type", "application/json")
	traitReq.Header.Set("X-Tenant-ID", tenantID)
	traitW := httptest.NewRecorder()
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

	t.Run("add trait to archetype", func(t *testing.T) {
		body := `{"trait_id": "` + traitID + `", "default_value": "high"}`
		req := httptest.NewRequest("POST", "/api/v1/archetypes/"+archetypeID+"/traits", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", archetypeID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.AddTrait).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}
	})
}

func TestArchetypeHandler_RemoveTrait(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	handler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitUseCase, removeTraitUseCase, getTraitsUseCase, log)

	// Create trait handler for trait creation
	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)

	// Create archetype
	archetypeBody := `{"name": "Warrior"}`
	archetypeReq := httptest.NewRequest("POST", "/api/v1/archetypes", strings.NewReader(archetypeBody))
	archetypeReq.Header.Set("Content-Type", "application/json")
	archetypeReq.Header.Set("X-Tenant-ID", tenantID)
	archetypeW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(archetypeW, archetypeReq)

	if archetypeW.Code != http.StatusCreated {
		t.Fatalf("failed to create archetype: status %d, body: %s", archetypeW.Code, archetypeW.Body.String())
	}

	var archetypeResp map[string]interface{}
	if err := json.NewDecoder(archetypeW.Body).Decode(&archetypeResp); err != nil {
		t.Fatalf("failed to decode archetype response: %v", err)
	}

	archetypeObj, ok := archetypeResp["archetype"].(map[string]interface{})
	if !ok {
		t.Fatalf("archetype response missing archetype object: %v", archetypeResp)
	}

	archetypeID, ok := archetypeObj["id"].(string)
	if !ok {
		t.Fatalf("archetype response missing id: %v", archetypeObj)
	}

	// Create trait
	traitBody := `{"name": "Brave"}`
	traitReq := httptest.NewRequest("POST", "/api/v1/traits", strings.NewReader(traitBody))
	traitReq.Header.Set("Content-Type", "application/json")
	traitReq.Header.Set("X-Tenant-ID", tenantID)
	traitW := httptest.NewRecorder()
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

	// Add trait to archetype first
	addBody := `{"trait_id": "` + traitID + `", "default_value": "high"}`
	addReq := httptest.NewRequest("POST", "/api/v1/archetypes/"+archetypeID+"/traits", strings.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.Header.Set("X-Tenant-ID", tenantID)
	addReq.SetPathValue("id", archetypeID)
	addW := httptest.NewRecorder()
	withTenantMiddleware(handler.AddTrait).ServeHTTP(addW, addReq)

	if addW.Code != http.StatusCreated {
		t.Fatalf("failed to add trait: status %d, body: %s", addW.Code, addW.Body.String())
	}

	t.Run("remove trait from archetype", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/archetypes/"+archetypeID+"/traits/"+traitID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", archetypeID)
		req.SetPathValue("trait_id", traitID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.RemoveTrait).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})
}

