package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/model"
	"github.com/mbeltran93/taskflow-go/internal/repository"
)

type CreateTaskInput struct {
	Title       string
	Description string
	ProjectID   int64
	AssigneeID  *int64
	DueDate     *model.Date
}

type UpdateTaskInput struct {
	Title       string
	Description string
	AssigneeID  *int64
	DueDate     *model.Date
}

type ListTasksInput struct {
	ProjectID *int64
	Status    *model.Status
}

// TaskService implements CRUD and status transitions for tasks.
type TaskService struct {
	tasks    repository.TaskRepository
	projects repository.ProjectRepository
	users    repository.UserRepository
}

func NewTaskService(tasks repository.TaskRepository, projects repository.ProjectRepository, users repository.UserRepository) *TaskService {
	return &TaskService{tasks: tasks, projects: projects, users: users}
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*model.Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || in.ProjectID == 0 {
		return nil, fmt.Errorf("title and projectId are required: %w", apperror.ErrInvalidInput)
	}

	if _, err := s.projects.GetByID(ctx, in.ProjectID); err != nil {
		return nil, fmt.Errorf("project %d: %w", in.ProjectID, err)
	}
	if in.AssigneeID != nil {
		if _, err := s.users.GetByID(ctx, *in.AssigneeID); err != nil {
			return nil, fmt.Errorf("assignee %d: %w", *in.AssigneeID, err)
		}
	}

	t := &model.Task{
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Status:      model.StatusTodo,
		ProjectID:   in.ProjectID,
		AssigneeID:  in.AssigneeID,
		DueDate:     in.DueDate,
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	return s.tasks.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, in ListTasksInput) ([]model.Task, error) {
	if in.Status != nil && !in.Status.Valid() {
		return nil, fmt.Errorf("status %q is not one of TODO, IN_PROGRESS, DONE: %w", *in.Status, apperror.ErrInvalidInput)
	}
	return s.tasks.List(ctx, repository.TaskFilter{ProjectID: in.ProjectID, Status: in.Status})
}

func (s *TaskService) Update(ctx context.Context, id int64, in UpdateTaskInput) (*model.Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required: %w", apperror.ErrInvalidInput)
	}
	if in.AssigneeID != nil {
		if _, err := s.users.GetByID(ctx, *in.AssigneeID); err != nil {
			return nil, fmt.Errorf("assignee %d: %w", *in.AssigneeID, err)
		}
	}

	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Title = title
	t.Description = strings.TrimSpace(in.Description)
	t.AssigneeID = in.AssigneeID
	t.DueDate = in.DueDate

	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) UpdateStatus(ctx context.Context, id int64, status model.Status) (*model.Task, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("status %q is not one of TODO, IN_PROGRESS, DONE: %w", status, apperror.ErrInvalidInput)
	}
	if err := s.tasks.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	return s.tasks.GetByID(ctx, id)
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	return s.tasks.Delete(ctx, id)
}
