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

var _ repositories.RPGClassSkillRepository = (*RPGClassSkillRepository)(nil)

// RPGClassSkillRepository implements the RPG class skill repository interface
type RPGClassSkillRepository struct {
	db *DB
}

// NewRPGClassSkillRepository creates a new RPG class skill repository
func NewRPGClassSkillRepository(db *DB) *RPGClassSkillRepository {
	return &RPGClassSkillRepository{db: db}
}

// Create creates a new RPG class skill
func (r *RPGClassSkillRepository) Create(ctx context.Context, classSkill *rpg.RPGClassSkill) error {
	// Get tenant_id from class
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM rpg_classes WHERE id = $1", classSkill.ClassID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO rpg_class_skills (id, tenant_id, class_id, skill_id, unlock_level, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		classSkill.ID, tenantID, classSkill.ClassID, classSkill.SkillID, classSkill.UnlockLevel, classSkill.CreatedAt)
	return err
}

// GetByID retrieves an RPG class skill by ID
func (r *RPGClassSkillRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.RPGClassSkill, error) {
	query := `
		SELECT id, class_id, skill_id, unlock_level, created_at
		FROM rpg_class_skills
		WHERE tenant_id = $1 AND id = $2
	`
	var cs rpg.RPGClassSkill

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&cs.ID, &cs.ClassID, &cs.SkillID, &cs.UnlockLevel, &cs.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "rpg_class_skill",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &cs, nil
}

// GetByClassAndSkill retrieves an RPG class skill by class and skill IDs
func (r *RPGClassSkillRepository) GetByClassAndSkill(ctx context.Context, tenantID, classID, skillID uuid.UUID) (*rpg.RPGClassSkill, error) {
	query := `
		SELECT id, class_id, skill_id, unlock_level, created_at
		FROM rpg_class_skills
		WHERE tenant_id = $1 AND class_id = $2 AND skill_id = $3
	`
	var cs rpg.RPGClassSkill

	err := r.db.QueryRow(ctx, query, tenantID, classID, skillID).Scan(
		&cs.ID, &cs.ClassID, &cs.SkillID, &cs.UnlockLevel, &cs.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "rpg_class_skill",
				ID:       classID.String() + "/" + skillID.String(),
			}
		}
		return nil, err
	}

	return &cs, nil
}

// ListByClass lists all skills for a class
func (r *RPGClassSkillRepository) ListByClass(ctx context.Context, tenantID, classID uuid.UUID) ([]*rpg.RPGClassSkill, error) {
	query := `
		SELECT id, class_id, skill_id, unlock_level, created_at
		FROM rpg_class_skills
		WHERE tenant_id = $1 AND class_id = $2
		ORDER BY unlock_level ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRPGClassSkills(rows)
}

// Update updates an RPG class skill
func (r *RPGClassSkillRepository) Update(ctx context.Context, classSkill *rpg.RPGClassSkill) error {
	// Get tenant_id from class
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM rpg_classes WHERE id = $1", classSkill.ClassID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		UPDATE rpg_class_skills
		SET unlock_level = $2
		WHERE tenant_id = $3 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, classSkill.ID, classSkill.UnlockLevel, tenantID)
	return err
}

// Delete deletes an RPG class skill
func (r *RPGClassSkillRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM rpg_class_skills WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByClass deletes all skills for a class
func (r *RPGClassSkillRepository) DeleteByClass(ctx context.Context, tenantID, classID uuid.UUID) error {
	query := `DELETE FROM rpg_class_skills WHERE tenant_id = $1 AND class_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, classID)
	return err
}

func (r *RPGClassSkillRepository) scanRPGClassSkills(rows pgx.Rows) ([]*rpg.RPGClassSkill, error) {
	skills := make([]*rpg.RPGClassSkill, 0)
	for rows.Next() {
		var cs rpg.RPGClassSkill

		err := rows.Scan(
			&cs.ID, &cs.ClassID, &cs.SkillID, &cs.UnlockLevel, &cs.CreatedAt)
		if err != nil {
			return nil, err
		}

		skills = append(skills, &cs)
	}
	return skills, rows.Err()
}
