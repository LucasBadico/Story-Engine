package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.CharacterSkillRepository = (*CharacterSkillRepository)(nil)

// CharacterSkillRepository implements the character skill repository interface
type CharacterSkillRepository struct {
	db *DB
}

// NewCharacterSkillRepository creates a new character skill repository
func NewCharacterSkillRepository(db *DB) *CharacterSkillRepository {
	return &CharacterSkillRepository{db: db}
}

// Create creates a new character skill
func (r *CharacterSkillRepository) Create(ctx context.Context, characterSkill *rpg.CharacterSkill) error {
	// Get tenant_id from character
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM characters WHERE id = $1", characterSkill.CharacterID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO character_skills (id, tenant_id, character_id, skill_id, rank, xp_in_skill, is_active, acquired_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		characterSkill.ID, tenantID, characterSkill.CharacterID, characterSkill.SkillID,
		characterSkill.Rank, characterSkill.XPInSkill, characterSkill.IsActive,
		characterSkill.AcquiredAt)
	return err
}

// GetByID retrieves a character skill by ID
func (r *CharacterSkillRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.CharacterSkill, error) {
	query := `
		SELECT id, character_id, skill_id, rank, xp_in_skill, is_active, acquired_at
		FROM character_skills
		WHERE tenant_id = $1 AND id = $2
	`
	var cs rpg.CharacterSkill

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&cs.ID, &cs.CharacterID, &cs.SkillID, &cs.Rank, &cs.XPInSkill, &cs.IsActive, &cs.AcquiredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_skill",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &cs, nil
}

// GetByCharacterAndSkill retrieves a character skill by character and skill IDs
func (r *CharacterSkillRepository) GetByCharacterAndSkill(ctx context.Context, tenantID, characterID, skillID uuid.UUID) (*rpg.CharacterSkill, error) {
	query := `
		SELECT id, character_id, skill_id, rank, xp_in_skill, is_active, acquired_at
		FROM character_skills
		WHERE tenant_id = $1 AND character_id = $2 AND skill_id = $3
	`
	var cs rpg.CharacterSkill

	err := r.db.QueryRow(ctx, query, tenantID, characterID, skillID).Scan(
		&cs.ID, &cs.CharacterID, &cs.SkillID, &cs.Rank, &cs.XPInSkill, &cs.IsActive, &cs.AcquiredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_skill",
				ID:       characterID.String() + "/" + skillID.String(),
			}
		}
		return nil, err
	}

	return &cs, nil
}

// ListByCharacter lists all skills for a character
func (r *CharacterSkillRepository) ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*rpg.CharacterSkill, error) {
	query := `
		SELECT id, character_id, skill_id, rank, xp_in_skill, is_active, acquired_at
		FROM character_skills
		WHERE tenant_id = $1 AND character_id = $2
		ORDER BY acquired_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterSkills(rows)
}

// ListActiveByCharacter lists active skills for a character
func (r *CharacterSkillRepository) ListActiveByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*rpg.CharacterSkill, error) {
	query := `
		SELECT id, character_id, skill_id, rank, xp_in_skill, is_active, acquired_at
		FROM character_skills
		WHERE tenant_id = $1 AND character_id = $2 AND is_active = TRUE
		ORDER BY acquired_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterSkills(rows)
}

// Update updates a character skill
func (r *CharacterSkillRepository) Update(ctx context.Context, characterSkill *rpg.CharacterSkill) error {
	// Get tenant_id from character
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM characters WHERE id = $1", characterSkill.CharacterID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		UPDATE character_skills
		SET rank = $2, xp_in_skill = $3, is_active = $4
		WHERE tenant_id = $5 AND id = $1
	`
	_, err := r.db.Exec(ctx, query,
		characterSkill.ID, characterSkill.Rank, characterSkill.XPInSkill, characterSkill.IsActive, tenantID)
	return err
}

// Delete deletes a character skill
func (r *CharacterSkillRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_skills WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "character_skill",
			ID:       id.String(),
		}
	}
	return nil
}

// DeleteByCharacter deletes all skills for a character
func (r *CharacterSkillRepository) DeleteByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) error {
	query := `DELETE FROM character_skills WHERE tenant_id = $1 AND character_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, characterID)
	return err
}

func (r *CharacterSkillRepository) scanCharacterSkills(rows pgx.Rows) ([]*rpg.CharacterSkill, error) {
	skills := make([]*rpg.CharacterSkill, 0)
	for rows.Next() {
		var cs rpg.CharacterSkill

		err := rows.Scan(
			&cs.ID, &cs.CharacterID, &cs.SkillID, &cs.Rank, &cs.XPInSkill, &cs.IsActive, &cs.AcquiredAt)
		if err != nil {
			return nil, err
		}

		skills = append(skills, &cs)
	}
	return skills, rows.Err()
}


