package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/model"
	"github.com/mbeltran93/taskflow-go/internal/repository"
)

// TaskRepository is the Postgres-backed implementation of repository.TaskRepository.
type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t *model.Task) error {
	const q = `
		INSERT INTO tasks (title, description, status, project_id, assignee_id, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, q, t.Title, t.Description, t.Status, t.ProjectID, t.AssigneeID, t.DueDate).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("project or assignee does not exist: %w", apperror.ErrInvalidInput)
		}
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	var t model.Task
	const q = `
		SELECT id, title, description, status, project_id, assignee_id, due_date, created_at, updated_at
		FROM tasks WHERE id = $1`
	if err := r.db.GetContext(ctx, &t, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("get task %d: %w", id, err)
	}
	return &t, nil
}

func (r *TaskRepository) List(ctx context.Context, filter repository.TaskFilter) ([]model.Task, error) {
	q := strings.Builder{}
	q.WriteString(`SELECT id, title, description, status, project_id, assignee_id, due_date, created_at, updated_at FROM tasks WHERE 1=1`)

	args := []any{}
	if filter.ProjectID != nil {
		args = append(args, *filter.ProjectID)
		fmt.Fprintf(&q, " AND project_id = $%d", len(args))
	}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		fmt.Fprintf(&q, " AND status = $%d", len(args))
	}
	q.WriteString(" ORDER BY id")

	var tasks []model.Task
	if err := r.db.SelectContext(ctx, &tasks, q.String(), args...); err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, t *model.Task) error {
	const q = `
		UPDATE tasks
		SET title = $1, description = $2, assignee_id = $3, due_date = $4, updated_at = now()
		WHERE id = $5
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, q, t.Title, t.Description, t.AssigneeID, t.DueDate, t.ID).Scan(&t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrNotFound
		}
		if isForeignKeyViolation(err) {
			return fmt.Errorf("assignee does not exist: %w", apperror.ErrInvalidInput)
		}
		return fmt.Errorf("update task %d: %w", t.ID, err)
	}
	return nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id int64, status model.Status) error {
	const q = `UPDATE tasks SET status = $1, updated_at = now() WHERE id = $2`
	res, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update task %d status: %w", id, err)
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

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM tasks WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete task %d: %w", id, err)
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
