package service

import (
	"context"
	"time"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/model"
	"github.com/mbeltran93/taskflow-go/internal/repository"
)

// The fakes below are plain in-memory structs satisfying the repository
// interfaces. This is the whole "mocking" story needed here: no mocking
// framework, no generated code, just a small type that implements an
// interface - which is idiomatic Go dependency injection for tests.

type fakeUserRepo struct {
	byID    map[int64]*model.User
	byEmail map[string]*model.User
	nextID  int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[int64]*model.User{}, byEmail: map[string]*model.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u *model.User) error {
	if _, exists := f.byEmail[u.Email]; exists {
		return apperror.ErrConflict
	}
	f.nextID++
	u.ID = f.nextID
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	cp := *u
	f.byID[u.ID] = &cp
	f.byEmail[u.Email] = &cp
	return nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) List(_ context.Context) ([]model.User, error) {
	out := make([]model.User, 0, len(f.byID))
	for _, u := range f.byID {
		out = append(out, *u)
	}
	return out, nil
}

type fakeProjectRepo struct {
	byID   map[int64]*model.Project
	nextID int64
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{byID: map[int64]*model.Project{}}
}

func (f *fakeProjectRepo) Create(_ context.Context, p *model.Project) error {
	f.nextID++
	p.ID = f.nextID
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	cp := *p
	f.byID[p.ID] = &cp
	return nil
}

func (f *fakeProjectRepo) GetByID(_ context.Context, id int64) (*model.Project, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (f *fakeProjectRepo) List(_ context.Context) ([]model.Project, error) {
	out := make([]model.Project, 0, len(f.byID))
	for _, p := range f.byID {
		out = append(out, *p)
	}
	return out, nil
}

func (f *fakeProjectRepo) Update(_ context.Context, p *model.Project) error {
	if _, ok := f.byID[p.ID]; !ok {
		return apperror.ErrNotFound
	}
	p.UpdatedAt = time.Now()
	cp := *p
	f.byID[p.ID] = &cp
	return nil
}

func (f *fakeProjectRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.byID[id]; !ok {
		return apperror.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}

type fakeTaskRepo struct {
	byID   map[int64]*model.Task
	nextID int64
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{byID: map[int64]*model.Task{}}
}

func (f *fakeTaskRepo) Create(_ context.Context, t *model.Task) error {
	f.nextID++
	t.ID = f.nextID
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	cp := *t
	f.byID[t.ID] = &cp
	return nil
}

func (f *fakeTaskRepo) GetByID(_ context.Context, id int64) (*model.Task, error) {
	t, ok := f.byID[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (f *fakeTaskRepo) List(_ context.Context, filter repository.TaskFilter) ([]model.Task, error) {
	out := []model.Task{}
	for _, t := range f.byID {
		if filter.ProjectID != nil && t.ProjectID != *filter.ProjectID {
			continue
		}
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		out = append(out, *t)
	}
	return out, nil
}

func (f *fakeTaskRepo) Update(_ context.Context, t *model.Task) error {
	if _, ok := f.byID[t.ID]; !ok {
		return apperror.ErrNotFound
	}
	t.UpdatedAt = time.Now()
	cp := *t
	f.byID[t.ID] = &cp
	return nil
}

func (f *fakeTaskRepo) UpdateStatus(_ context.Context, id int64, status model.Status) error {
	t, ok := f.byID[id]
	if !ok {
		return apperror.ErrNotFound
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

func (f *fakeTaskRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.byID[id]; !ok {
		return apperror.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}
