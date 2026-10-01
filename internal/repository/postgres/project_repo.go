package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/model"
)

// ProjectRepository is the Postgres-backed implementation of repository.ProjectRepository.
type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *model.Project) error {
	const q = `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, q, p.Name, p.Description, p.OwnerID).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("owner %d does not exist: %w", p.OwnerID, apperror.ErrInvalidInput)
		}
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	var p model.Project
	const q = `SELECT id, name, description, owner_id, created_at, updated_at FROM projects WHERE id = $1`
	if err := r.db.GetContext(ctx, &p, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("get project %d: %w", id, err)
	}
	return &p, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	const q = `SELECT id, name, description, owner_id, created_at, updated_at FROM projects ORDER BY id`
	if err := r.db.SelectContext(ctx, &projects, q); err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *model.Project) error {
	const q = `
		UPDATE projects
		SET name = $1, description = $2, updated_at = now()
		WHERE id = $3
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, q, p.Name, p.Description, p.ID).Scan(&p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("update project %d: %w", p.ID, err)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM projects WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
