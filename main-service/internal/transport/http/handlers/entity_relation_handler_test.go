//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestEntityRelationHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	log := logger.New()
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	summaryGenerator := relationapp.NewSummaryGenerator()

	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	listRelationsByWorldUseCase := relationapp.NewListRelationsByWorldUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)

	handler := NewEntityRelationHandler(
		createRelationUseCase,
		getRelationUseCase,
		listRelationsBySourceUseCase,
		listRelationsByTargetUseCase,
		listRelationsByWorldUseCase,
		updateRelationUseCase,
		deleteRelationUseCase,
		log,
	)

	sourceID := createTestCharacter(t, db, tenantID, worldID, "Source")
	targetID := createTestCharacter(t, db, tenantID, worldID, "Target")
	otherTargetID := createTestCharacter(t, db, tenantID, worldID, "Other Target")

	t.Run("mirror true with known type", func(t *testing.T) {
		body := `{"world_id":"` + worldID + `","source_type":"character","source_id":"` + sourceID + `","target_type":"character","target_id":"` + targetID + `","relation_type":"ally_of","create_mirror":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/relations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := resp["mirror"]; !ok {
			t.Fatalf("expected mirror in response")
		}
	})

	t.Run("mirror true with custom type", func(t *testing.T) {
		body := `{"world_id":"` + worldID + `","source_type":"character","source_id":"` + sourceID + `","target_type":"character","target_id":"` + targetID + `","relation_type":"custom:bonded_to","create_mirror":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/relations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := resp["mirror"]; ok {
			t.Fatalf("did not expect mirror for custom relation")
		}
	})

	t.Run("mirror false", func(t *testing.T) {
		body := `{"world_id":"` + worldID + `","source_type":"character","source_id":"` + sourceID + `","target_type":"character","target_id":"` + otherTargetID + `","relation_type":"ally_of","create_mirror":false}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/relations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := resp["mirror"]; ok {
			t.Fatalf("did not expect mirror when create_mirror is false")
		}
	})
}

func createTestCharacter(t *testing.T, db *postgres.DB, tenantID, worldID, name string) string {
	t.Helper()

	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	tenantUUID := uuid.MustParse(tenantID)
	worldUUID := uuid.MustParse(worldID)

	useCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, logger.New())
	output, err := useCase.Execute(context.Background(), characterapp.CreateCharacterInput{
		TenantID:    tenantUUID,
		WorldID:     worldUUID,
		Name:        name,
		Description: "",
	})
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	return output.Character.ID.String()
}
