//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCharacterSkillHandler_Learn(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, postgres.NewCharacterTraitRepository(db), worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), postgres.NewCharacterTraitRepository(db), log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(postgres.NewCharacterTraitRepository(db), log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(postgres.NewCharacterTraitRepository(db), log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(postgres.NewCharacterTraitRepository(db), log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, postgres.NewRPGClassRepository(db), auditLogRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, postgres.NewRPGClassRepository(db), log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Test Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(characterHandler.Create).ServeHTTP(characterW, characterReq)

	if characterW.Code != http.StatusCreated {
		t.Fatalf("failed to create character: status %d, body: %s", characterW.Code, characterW.Body.String())
	}

	var characterResp map[string]interface{}
	if err := json.NewDecoder(characterW.Body).Decode(&characterResp); err != nil {
		t.Fatalf("failed to decode character response: %v", err)
	}

	characterObj, ok := characterResp["character"].(map[string]interface{})
	if !ok {
		t.Fatalf("character response missing character object: %v", characterResp)
	}

	characterID, ok := characterObj["id"].(string)
	if !ok {
		t.Fatalf("character response missing id: %v", characterObj)
	}

	// Create RPG system and skill
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

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
	skillHandler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	skillBody := `{"name": "Test Skill"}`
	skillReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/skills", strings.NewReader(skillBody))
	skillReq.Header.Set("Content-Type", "application/json")
	skillReq.Header.Set("X-Tenant-ID", tenantID)
	skillReq.SetPathValue("id", rpgSystemID)
	skillW := httptest.NewRecorder()
	withTenantMiddleware(skillHandler.Create).ServeHTTP(skillW, skillReq)

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

	characterSkillRepo := postgres.NewCharacterSkillRepository(db)
	learnSkillUseCase := characterskillapp.NewLearnSkillUseCase(characterSkillRepo, characterRepo, skillRepo, log)
	listSkillsUseCase := characterskillapp.NewListCharacterSkillsUseCase(characterSkillRepo, log)
	updateSkillUseCase := characterskillapp.NewUpdateCharacterSkillUseCase(characterSkillRepo, skillRepo, log)
	deleteSkillUseCase := characterskillapp.NewDeleteCharacterSkillUseCase(characterSkillRepo, log)
	handler := NewCharacterSkillHandler(learnSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	t.Run("successful learning", func(t *testing.T) {
		body := `{"skill_id": "` + skillID + `"}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/skills", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Learn).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if characterSkill, ok := resp["character_skill"].(map[string]interface{}); ok {
			if characterSkill["skill_id"] != skillID {
				t.Errorf("expected skill_id %s, got %v", skillID, characterSkill["skill_id"])
			}
		} else {
			t.Error("response missing character_skill")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"skill_id": "` + skillID + `"}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/skills", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Learn).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCharacterSkillHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, postgres.NewCharacterTraitRepository(db), worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), postgres.NewCharacterTraitRepository(db), log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(postgres.NewCharacterTraitRepository(db), log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(postgres.NewCharacterTraitRepository(db), log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(postgres.NewCharacterTraitRepository(db), log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, postgres.NewRPGClassRepository(db), auditLogRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, postgres.NewRPGClassRepository(db), log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Test Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(characterHandler.Create).ServeHTTP(characterW, characterReq)

	if characterW.Code != http.StatusCreated {
		t.Fatalf("failed to create character: status %d, body: %s", characterW.Code, characterW.Body.String())
	}

	var characterResp map[string]interface{}
	if err := json.NewDecoder(characterW.Body).Decode(&characterResp); err != nil {
		t.Fatalf("failed to decode character response: %v", err)
	}

	characterObj, ok := characterResp["character"].(map[string]interface{})
	if !ok {
		t.Fatalf("character response missing character object: %v", characterResp)
	}

	characterID, ok := characterObj["id"].(string)
	if !ok {
		t.Fatalf("character response missing id: %v", characterObj)
	}

	characterSkillRepo := postgres.NewCharacterSkillRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	learnSkillUseCase := characterskillapp.NewLearnSkillUseCase(characterSkillRepo, characterRepo, skillRepo, log)
	listSkillsUseCase := characterskillapp.NewListCharacterSkillsUseCase(characterSkillRepo, log)
	updateSkillUseCase := characterskillapp.NewUpdateCharacterSkillUseCase(characterSkillRepo, skillRepo, log)
	deleteSkillUseCase := characterskillapp.NewDeleteCharacterSkillUseCase(characterSkillRepo, log)
	handler := NewCharacterSkillHandler(learnSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	t.Run("list skills", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/"+characterID+"/skills", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := resp["skills"]; !ok {
			t.Error("response missing skills")
		}
	})
}

