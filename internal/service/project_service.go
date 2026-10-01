package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/model"
	"github.com/mbeltran93/taskflow-go/internal/repository"
)

type CreateProjectInput struct {
	Name        string
	Description string
	OwnerID     int64
}

type UpdateProjectInput struct {
	Name        string
	Description string
}

// ProjectService implements CRUD for projects.
type ProjectService struct {
	projects repository.ProjectRepository
	users    repository.UserRepository
}

func NewProjectService(projects repository.ProjectRepository, users repository.UserRepository) *ProjectService {
	return &ProjectService{projects: projects, users: users}
}

func (s *ProjectService) Create(ctx context.Context, in CreateProjectInput) (*model.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || in.OwnerID == 0 {
		return nil, fmt.Errorf("name and ownerId are required: %w", apperror.ErrInvalidInput)
	}

	if _, err := s.users.GetByID(ctx, in.OwnerID); err != nil {
		return nil, fmt.Errorf("owner %d: %w", in.OwnerID, err)
	}

	p := &model.Project{Name: name, Description: strings.TrimSpace(in.Description), OwnerID: in.OwnerID}
	if err := s.projects.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	return s.projects.GetByID(ctx, id)
}

func (s *ProjectService) List(ctx context.Context) ([]model.Project, error) {
	return s.projects.List(ctx)
}

func (s *ProjectService) Update(ctx context.Context, id int64, in UpdateProjectInput) (*model.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required: %w", apperror.ErrInvalidInput)
	}

	p, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Name = name
	p.Description = strings.TrimSpace(in.Description)

	if err := s.projects.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	return s.projects.Delete(ctx, id)
}
