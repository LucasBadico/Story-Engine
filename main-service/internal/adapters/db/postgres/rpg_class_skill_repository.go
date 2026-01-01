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
	query := `
		INSERT INTO rpg_class_skills (id, class_id, skill_id, unlock_level, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		classSkill.ID, classSkill.ClassID, classSkill.SkillID, classSkill.UnlockLevel, classSkill.CreatedAt)
	return err
}

// GetByID retrieves an RPG class skill by ID
func (r *RPGClassSkillRepository) GetByID(ctx context.Context, id uuid.UUID) (*rpg.RPGClassSkill, error) {
	query := `
		SELECT id, class_id, skill_id, unlock_level, created_at
		FROM rpg_class_skills
		WHERE id = $1
	`
	var cs rpg.RPGClassSkill

	err := r.db.QueryRow(ctx, query, id).Scan(
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
func (r *RPGClassSkillRepository) GetByClassAndSkill(ctx context.Context, classID, skillID uuid.UUID) (*rpg.RPGClassSkill, error) {
	query := `
		SELECT id, class_id, skill_id, unlock_level, created_at
		FROM rpg_class_skills
		WHERE class_id = $1 AND skill_id = $2
	`
	var cs rpg.RPGClassSkill

	err := r.db.QueryRow(ctx, query, classID, skillID).Scan(
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
func (r *RPGClassSkillRepository) ListByClass(ctx context.Context, classID uuid.UUID) ([]*rpg.RPGClassSkill, error) {
	query := `
		SELECT id, class_id, skill_id, unlock_level, created_at
		FROM rpg_class_skills
		WHERE class_id = $1
		ORDER BY unlock_level ASC
	`
	rows, err := r.db.Query(ctx, query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRPGClassSkills(rows)
}

// Update updates an RPG class skill
func (r *RPGClassSkillRepository) Update(ctx context.Context, classSkill *rpg.RPGClassSkill) error {
	query := `
		UPDATE rpg_class_skills
		SET unlock_level = $2
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, classSkill.ID, classSkill.UnlockLevel)
	return err
}

// Delete deletes an RPG class skill
func (r *RPGClassSkillRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM rpg_class_skills WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByClass deletes all skills for a class
func (r *RPGClassSkillRepository) DeleteByClass(ctx context.Context, classID uuid.UUID) error {
	query := `DELETE FROM rpg_class_skills WHERE class_id = $1`
	_, err := r.db.Exec(ctx, query, classID)
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

