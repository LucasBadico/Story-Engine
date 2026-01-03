package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.SkillRepository = (*SkillRepository)(nil)

// SkillRepository implements the skill repository interface
type SkillRepository struct {
	db *DB
}

// NewSkillRepository creates a new skill repository
func NewSkillRepository(db *DB) *SkillRepository {
	return &SkillRepository{db: db}
}

// Create creates a new skill
func (r *SkillRepository) Create(ctx context.Context, skill *rpg.Skill) error {
	query := `
		INSERT INTO rpg_skills (id, tenant_id, rpg_system_id, name, category, type, description, prerequisites, max_rank, effects_schema, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	var category *string
	var skillType *string
	if skill.Category != nil {
		categoryStr := string(*skill.Category)
		category = &categoryStr
	}
	if skill.Type != nil {
		typeStr := string(*skill.Type)
		skillType = &typeStr
	}

	_, err := r.db.Exec(ctx, query,
		skill.ID, skill.TenantID, skill.RPGSystemID, skill.Name, category, skillType, skill.Description,
		skill.Prerequisites, skill.MaxRank, skill.EffectsSchema,
		skill.CreatedAt, skill.UpdatedAt)
	return err
}

// GetByID retrieves a skill by ID
func (r *SkillRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.Skill, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, name, category, type, description, prerequisites, max_rank, effects_schema, created_at, updated_at
		FROM rpg_skills
		WHERE tenant_id = $1 AND id = $2
	`
	var skill rpg.Skill
	var category sql.NullString
	var skillType sql.NullString
	var description sql.NullString
	var prerequisites sql.NullString
	var effectsSchema sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&skill.ID, &skill.TenantID, &skill.RPGSystemID, &skill.Name, &category, &skillType, &description,
		&prerequisites, &skill.MaxRank, &effectsSchema,
		&skill.CreatedAt, &skill.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "rpg_skill",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if category.Valid {
		cat := rpg.SkillCategory(category.String)
		skill.Category = &cat
	}
	if skillType.Valid {
		st := rpg.SkillType(skillType.String)
		skill.Type = &st
	}
	if description.Valid {
		skill.Description = &description.String
	}
	if prerequisites.Valid {
		prereq := json.RawMessage(prerequisites.String)
		skill.Prerequisites = &prereq
	}
	if effectsSchema.Valid {
		effects := json.RawMessage(effectsSchema.String)
		skill.EffectsSchema = &effects
	}

	return &skill, nil
}

// ListBySystem lists skills for an RPG system
func (r *SkillRepository) ListBySystem(ctx context.Context, tenantID, rpgSystemID uuid.UUID) ([]*rpg.Skill, error) {
	query := `
		SELECT id, tenant_id, rpg_system_id, name, category, type, description, prerequisites, max_rank, effects_schema, created_at, updated_at
		FROM rpg_skills
		WHERE tenant_id = $1 AND rpg_system_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, rpgSystemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSkills(rows)
}

// Update updates a skill
func (r *SkillRepository) Update(ctx context.Context, skill *rpg.Skill) error {
	query := `
		UPDATE rpg_skills
		SET name = $2, category = $3, type = $4, description = $5, prerequisites = $6, max_rank = $7, effects_schema = $8, updated_at = $9
		WHERE tenant_id = $10 AND id = $1
	`
	var category *string
	var skillType *string
	if skill.Category != nil {
		categoryStr := string(*skill.Category)
		category = &categoryStr
	}
	if skill.Type != nil {
		typeStr := string(*skill.Type)
		skillType = &typeStr
	}

	_, err := r.db.Exec(ctx, query,
		skill.ID, skill.Name, category, skillType, skill.Description,
		skill.Prerequisites, skill.MaxRank, skill.EffectsSchema,
		skill.UpdatedAt, skill.TenantID)
	return err
}

// Delete deletes a skill
func (r *SkillRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM rpg_skills WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "skill",
			ID:       id.String(),
		}
	}
	return nil
}

func (r *SkillRepository) scanSkills(rows pgx.Rows) ([]*rpg.Skill, error) {
	skills := make([]*rpg.Skill, 0)
	for rows.Next() {
		var skill rpg.Skill
		var category sql.NullString
		var skillType sql.NullString
		var description sql.NullString
		var prerequisites sql.NullString
		var effectsSchema sql.NullString

		err := rows.Scan(
			&skill.ID, &skill.TenantID, &skill.RPGSystemID, &skill.Name, &category, &skillType, &description,
			&prerequisites, &skill.MaxRank, &effectsSchema,
			&skill.CreatedAt, &skill.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if category.Valid {
			cat := rpg.SkillCategory(category.String)
			skill.Category = &cat
		}
		if skillType.Valid {
			st := rpg.SkillType(skillType.String)
			skill.Type = &st
		}
		if description.Valid {
			skill.Description = &description.String
		}
		if prerequisites.Valid {
			prereq := json.RawMessage(prerequisites.String)
			skill.Prerequisites = &prereq
		}
		if effectsSchema.Valid {
			effects := json.RawMessage(effectsSchema.String)
			skill.EffectsSchema = &effects
		}

		skills = append(skills, &skill)
	}
	return skills, rows.Err()
}


