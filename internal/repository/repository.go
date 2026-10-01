// Package repository declares the persistence interfaces the service layer
// depends on. Concrete implementations (internal/repository/postgres) are
// wired in by main; tests can swap in in-memory fakes instead, which is how
// Go gets dependency injection without a DI framework or runtime reflection.
package repository

import (
	"context"

	"github.com/mbeltran93/taskflow-go/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
}

type ProjectRepository interface {
	Create(ctx context.Context, p *model.Project) error
	GetByID(ctx context.Context, id int64) (*model.Project, error)
	List(ctx context.Context) ([]model.Project, error)
	Update(ctx context.Context, p *model.Project) error
	Delete(ctx context.Context, id int64) error
}

// TaskFilter narrows a Task listing. A nil field means "don't filter on this".
type TaskFilter struct {
	ProjectID *int64
	Status    *model.Status
}

type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	List(ctx context.Context, filter TaskFilter) ([]model.Task, error)
	Update(ctx context.Context, t *model.Task) error
	UpdateStatus(ctx context.Context, id int64, status model.Status) error
	Delete(ctx context.Context, id int64) error
}
