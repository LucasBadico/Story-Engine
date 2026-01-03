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
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestSkillHandler_Create(t *testing.T) {
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

	skillRepo := postgres.NewSkillRepository(db)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	handler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Skill", "category": "combat", "type": "active"}`
		req := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(body))
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

		if skill, ok := resp["skill"].(map[string]interface{}); ok {
			if skill["name"] != "Test Skill" {
				t.Errorf("expected name 'Test Skill', got %v", skill["name"])
			}
		} else {
			t.Error("response missing skill")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test Skill"}`
		req := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", rpgSystemID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestSkillHandler_Get(t *testing.T) {
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

	skillRepo := postgres.NewSkillRepository(db)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	handler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	// Create skill
	skillBody := `{"name": "Get Test Skill"}`
	skillReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(skillBody))
	skillReq.Header.Set("Content-Type", "application/json")
	skillReq.Header.Set("X-Tenant-ID", tenantID)
	skillReq.SetPathValue("id", rpgSystemID)
	skillW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(skillW, skillReq)

	if skillW.Code != http.StatusCreated {
		t.Fatalf("failed to create skill: status %d, body: %s", skillW.Code, skillW.Body.String())
	}

	var skillResp map[string]interface{}
	if err := json.NewDecoder(skillW.Body).Decode(&skillResp); err != nil {
		t.Fatalf("failed to decode skill response: %v", err)
	}

	skillObj, ok := skillResp["skill"].(map[string]interface{})
	if !ok {
		t.Fatalf("skill response missing skill object: %v", skillResp)
	}

	skillID, ok := skillObj["id"].(string)
	if !ok {
		t.Fatalf("skill response missing id: %v", skillObj)
	}

	t.Run("existing skill", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-skills/"+skillID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", skillID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if skill, ok := resp["skill"].(map[string]interface{}); ok {
			if skill["id"] != skillID {
				t.Errorf("expected ID %s, got %v", skillID, skill["id"])
			}
		} else {
			t.Error("response missing skill")
		}
	})

	t.Run("non-existing skill", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-skills/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestSkillHandler_List(t *testing.T) {
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

	skillRepo := postgres.NewSkillRepository(db)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	handler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	// Create multiple skills
	for i := 1; i <= 3; i++ {
		skillBody := `{"name": "Skill ` + strconv.Itoa(i) + `"}`
		skillReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(skillBody))
		skillReq.Header.Set("Content-Type", "application/json")
		skillReq.Header.Set("X-Tenant-ID", tenantID)
		skillReq.SetPathValue("id", rpgSystemID)
		skillW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(skillW, skillReq)

		if skillW.Code != http.StatusCreated {
			t.Fatalf("failed to create skill %d: status %d, body: %s", i, skillW.Code, skillW.Body.String())
		}
	}

	t.Run("list skills", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", nil)
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

		if skills, ok := resp["skills"].([]interface{}); ok {
			if len(skills) < 3 {
				t.Errorf("expected at least 3 skills, got %d", len(skills))
			}
		} else {
			t.Error("response missing skills")
		}
	})
}

func TestSkillHandler_Update(t *testing.T) {
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

	skillRepo := postgres.NewSkillRepository(db)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	handler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	// Create skill
	skillBody := `{"name": "Original Skill"}`
	skillReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(skillBody))
	skillReq.Header.Set("Content-Type", "application/json")
	skillReq.Header.Set("X-Tenant-ID", tenantID)
	skillReq.SetPathValue("id", rpgSystemID)
	skillW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(skillW, skillReq)

	if skillW.Code != http.StatusCreated {
		t.Fatalf("failed to create skill: status %d, body: %s", skillW.Code, skillW.Body.String())
	}

	var skillResp map[string]interface{}
	if err := json.NewDecoder(skillW.Body).Decode(&skillResp); err != nil {
		t.Fatalf("failed to decode skill response: %v", err)
	}

	skillObj, ok := skillResp["skill"].(map[string]interface{})
	if !ok {
		t.Fatalf("skill response missing skill object: %v", skillResp)
	}

	skillID, ok := skillObj["id"].(string)
	if !ok {
		t.Fatalf("skill response missing id: %v", skillObj)
	}

	t.Run("update skill name", func(t *testing.T) {
		body := `{"name": "Updated Skill"}`
		req := httptest.NewRequest("PUT", "/api/v1/rpg-skills/"+skillID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", skillID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if skill, ok := resp["skill"].(map[string]interface{}); ok {
			if skill["name"] != "Updated Skill" {
				t.Errorf("expected name 'Updated Skill', got %v", skill["name"])
			}
		} else {
			t.Error("response missing skill")
		}
	})

	t.Run("non-existing skill", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/rpg-skills/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestSkillHandler_Delete(t *testing.T) {
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

	skillRepo := postgres.NewSkillRepository(db)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	handler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	// Create skill
	skillBody := `{"name": "Skill to Delete"}`
	skillReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(skillBody))
	skillReq.Header.Set("Content-Type", "application/json")
	skillReq.Header.Set("X-Tenant-ID", tenantID)
	skillReq.SetPathValue("id", rpgSystemID)
	skillW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(skillW, skillReq)

	if skillW.Code != http.StatusCreated {
		t.Fatalf("failed to create skill: status %d, body: %s", skillW.Code, skillW.Body.String())
	}

	var skillResp map[string]interface{}
	if err := json.NewDecoder(skillW.Body).Decode(&skillResp); err != nil {
		t.Fatalf("failed to decode skill response: %v", err)
	}

	skillObj, ok := skillResp["skill"].(map[string]interface{})
	if !ok {
		t.Fatalf("skill response missing skill object: %v", skillResp)
	}

	skillID, ok := skillObj["id"].(string)
	if !ok {
		t.Fatalf("skill response missing id: %v", skillObj)
	}

	t.Run("delete existing skill", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/rpg-skills/"+skillID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", skillID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify skill is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/rpg-skills/"+skillID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", skillID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted skill, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing skill", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/rpg-skills/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

