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
	rpgclassapp "github.com/story-engine/main-service/internal/application/rpg/rpg_class"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestRPGClassHandler_Create(t *testing.T) {
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
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

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

	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	createClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	handler := NewRPGClassHandler(createClassUseCase, getClassUseCase, listClassesUseCase, updateClassUseCase, deleteClassUseCase, addSkillUseCase, listClassSkillsUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Class", "tier": 1, "description": "A test class"}`
		req := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgClass, ok := resp["rpg_class"].(map[string]interface{}); ok {
			if rpgClass["name"] != "Test Class" {
				t.Errorf("expected name 'Test Class', got %v", rpgClass["name"])
			}
		} else {
			t.Error("response missing rpg_class")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test Class"}`
		req := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestRPGClassHandler_Get(t *testing.T) {
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
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

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

	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	createClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	handler := NewRPGClassHandler(createClassUseCase, getClassUseCase, listClassesUseCase, updateClassUseCase, deleteClassUseCase, addSkillUseCase, listClassSkillsUseCase, log)

	// Create RPG class
	rpgClassBody := `{"name": "Get Test Class"}`
	rpgClassReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", strings.NewReader(rpgClassBody))
	rpgClassReq.Header.Set("Content-Type", "application/json")
	rpgClassReq.Header.Set("X-Tenant-ID", tenantID)
	rpgClassReq.SetPathValue("id", rpgSystemID)
	rpgClassW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(rpgClassW, rpgClassReq)

	if rpgClassW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG class: status %d, body: %s", rpgClassW.Code, rpgClassW.Body.String())
	}

	var rpgClassResp map[string]interface{}
	if err := json.NewDecoder(rpgClassW.Body).Decode(&rpgClassResp); err != nil {
		t.Fatalf("failed to decode RPG class response: %v", err)
	}

	rpgClassObj, ok := rpgClassResp["rpg_class"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG class response missing rpg_class object: %v", rpgClassResp)
	}

	rpgClassID, ok := rpgClassObj["id"].(string)
	if !ok {
		t.Fatalf("RPG class response missing id: %v", rpgClassObj)
	}

	t.Run("existing rpg_class", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-classes/"+rpgClassID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgClassID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgClass, ok := resp["rpg_class"].(map[string]interface{}); ok {
			if rpgClass["id"] != rpgClassID {
				t.Errorf("expected ID %s, got %v", rpgClassID, rpgClass["id"])
			}
		} else {
			t.Error("response missing rpg_class")
		}
	})

	t.Run("non-existing rpg_class", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-classes/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestRPGClassHandler_List(t *testing.T) {
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
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

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

	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	createClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	handler := NewRPGClassHandler(createClassUseCase, getClassUseCase, listClassesUseCase, updateClassUseCase, deleteClassUseCase, addSkillUseCase, listClassSkillsUseCase, log)

	// Create multiple RPG classes
	for i := 1; i <= 3; i++ {
		rpgClassBody := `{"name": "Class ` + strconv.Itoa(i) + `"}`
		rpgClassReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", strings.NewReader(rpgClassBody))
		rpgClassReq.Header.Set("Content-Type", "application/json")
		rpgClassReq.Header.Set("X-Tenant-ID", tenantID)
		rpgClassReq.SetPathValue("id", rpgSystemID)
		rpgClassW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(rpgClassW, rpgClassReq)

		if rpgClassW.Code != http.StatusCreated {
			t.Fatalf("failed to create RPG class %d: status %d, body: %s", i, rpgClassW.Code, rpgClassW.Body.String())
		}
	}

	t.Run("list rpg_classes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgClasses, ok := resp["rpg_classes"].([]interface{}); ok {
			if len(rpgClasses) < 3 {
				t.Errorf("expected at least 3 RPG classes, got %d", len(rpgClasses))
			}
		} else {
			t.Error("response missing rpg_classes")
		}
	})
}

func TestRPGClassHandler_Update(t *testing.T) {
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
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

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

	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	createClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	handler := NewRPGClassHandler(createClassUseCase, getClassUseCase, listClassesUseCase, updateClassUseCase, deleteClassUseCase, addSkillUseCase, listClassSkillsUseCase, log)

	// Create RPG class
	rpgClassBody := `{"name": "Original Class"}`
	rpgClassReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", strings.NewReader(rpgClassBody))
	rpgClassReq.Header.Set("Content-Type", "application/json")
	rpgClassReq.Header.Set("X-Tenant-ID", tenantID)
	rpgClassReq.SetPathValue("id", rpgSystemID)
	rpgClassW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(rpgClassW, rpgClassReq)

	if rpgClassW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG class: status %d, body: %s", rpgClassW.Code, rpgClassW.Body.String())
	}

	var rpgClassResp map[string]interface{}
	if err := json.NewDecoder(rpgClassW.Body).Decode(&rpgClassResp); err != nil {
		t.Fatalf("failed to decode RPG class response: %v", err)
	}

	rpgClassObj, ok := rpgClassResp["rpg_class"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG class response missing rpg_class object: %v", rpgClassResp)
	}

	rpgClassID, ok := rpgClassObj["id"].(string)
	if !ok {
		t.Fatalf("RPG class response missing id: %v", rpgClassObj)
	}

	t.Run("update rpg_class name", func(t *testing.T) {
		body := `{"name": "Updated Class"}`
		req := httptest.NewRequest("PUT", "/api/v1/rpg-classes/"+rpgClassID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgClassID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if rpgClass, ok := resp["rpg_class"].(map[string]interface{}); ok {
			if rpgClass["name"] != "Updated Class" {
				t.Errorf("expected name 'Updated Class', got %v", rpgClass["name"])
			}
		} else {
			t.Error("response missing rpg_class")
		}
	})

	t.Run("non-existing rpg_class", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/rpg-classes/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestRPGClassHandler_Delete(t *testing.T) {
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
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	// Create RPG system
	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

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

	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	createClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	handler := NewRPGClassHandler(createClassUseCase, getClassUseCase, listClassesUseCase, updateClassUseCase, deleteClassUseCase, addSkillUseCase, listClassSkillsUseCase, log)

	// Create RPG class
	rpgClassBody := `{"name": "Class to Delete"}`
	rpgClassReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/classes", strings.NewReader(rpgClassBody))
	rpgClassReq.Header.Set("Content-Type", "application/json")
	rpgClassReq.Header.Set("X-Tenant-ID", tenantID)
	rpgClassReq.SetPathValue("id", rpgSystemID)
	rpgClassW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(rpgClassW, rpgClassReq)

	if rpgClassW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG class: status %d, body: %s", rpgClassW.Code, rpgClassW.Body.String())
	}

	var rpgClassResp map[string]interface{}
	if err := json.NewDecoder(rpgClassW.Body).Decode(&rpgClassResp); err != nil {
		t.Fatalf("failed to decode RPG class response: %v", err)
	}

	rpgClassObj, ok := rpgClassResp["rpg_class"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG class response missing rpg_class object: %v", rpgClassResp)
	}

	rpgClassID, ok := rpgClassObj["id"].(string)
	if !ok {
		t.Fatalf("RPG class response missing id: %v", rpgClassObj)
	}

	t.Run("delete existing rpg_class", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/rpg-classes/"+rpgClassID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", rpgClassID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify RPG class is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/rpg-classes/"+rpgClassID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", rpgClassID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted RPG class, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing rpg_class", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/rpg-classes/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

