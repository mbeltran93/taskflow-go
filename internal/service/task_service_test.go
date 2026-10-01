package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/model"
)

func newTaskServiceFixture(t *testing.T) (*TaskService, int64, int64) {
	t.Helper()
	users := newFakeUserRepo()
	projects := newFakeProjectRepo()
	ctx := context.Background()

	owner, err := NewUserService(users).Register(ctx, RegisterUserInput{
		Name: "Owner", Email: "owner@example.com", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("seed owner failed: %v", err)
	}
	project, err := NewProjectService(projects, users).Create(ctx, CreateProjectInput{Name: "Launch", OwnerID: owner.ID})
	if err != nil {
		t.Fatalf("seed project failed: %v", err)
	}

	return NewTaskService(newFakeTaskRepo(), projects, users), project.ID, owner.ID
}

func TestTaskService_Create(t *testing.T) {
	svc, projectID, assigneeID := newTaskServiceFixture(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		buildInput func() CreateTaskInput
		wantErr    error
		wantStatus model.Status
	}{
		{
			name: "valid task defaults to TODO",
			buildInput: func() CreateTaskInput {
				return CreateTaskInput{Title: "Write docs", ProjectID: projectID, AssigneeID: &assigneeID}
			},
			wantStatus: model.StatusTodo,
		},
		{
			name: "blank title is rejected",
			buildInput: func() CreateTaskInput {
				return CreateTaskInput{Title: "  ", ProjectID: projectID}
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			name: "unknown project fails as not found",
			buildInput: func() CreateTaskInput {
				return CreateTaskInput{Title: "Write docs", ProjectID: 999}
			},
			wantErr: apperror.ErrNotFound,
		},
		{
			name: "unknown assignee fails as not found",
			buildInput: func() CreateTaskInput {
				ghost := int64(999)
				return CreateTaskInput{Title: "Write docs", ProjectID: projectID, AssigneeID: &ghost}
			},
			wantErr: apperror.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task, err := svc.Create(ctx, tc.buildInput())

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Create() error = %v, want wrapping %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if task.Status != tc.wantStatus {
				t.Errorf("Status = %q, want %q", task.Status, tc.wantStatus)
			}
		})
	}
}

func TestTaskService_UpdateStatus(t *testing.T) {
	svc, projectID, _ := newTaskServiceFixture(t)
	ctx := context.Background()

	task, err := svc.Create(ctx, CreateTaskInput{Title: "Write docs", ProjectID: projectID})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		status  model.Status
		wantErr error
	}{
		{name: "moving to IN_PROGRESS succeeds", status: model.StatusInProgress},
		{name: "moving to DONE succeeds", status: model.StatusDone},
		{name: "invalid status is rejected", status: "BOGUS", wantErr: apperror.ErrInvalidInput},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updated, err := svc.UpdateStatus(ctx, task.ID, tc.status)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("UpdateStatus() error = %v, want wrapping %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateStatus() unexpected error: %v", err)
			}
			if updated.Status != tc.status {
				t.Errorf("Status = %q, want %q", updated.Status, tc.status)
			}
		})
	}
}

func TestTaskService_List_FiltersByProjectAndStatus(t *testing.T) {
	svc, projectID, _ := newTaskServiceFixture(t)
	ctx := context.Background()

	a, _ := svc.Create(ctx, CreateTaskInput{Title: "A", ProjectID: projectID})
	b, _ := svc.Create(ctx, CreateTaskInput{Title: "B", ProjectID: projectID})
	if _, err := svc.UpdateStatus(ctx, b.ID, model.StatusDone); err != nil {
		t.Fatalf("UpdateStatus() unexpected error: %v", err)
	}

	done := model.StatusDone
	results, err := svc.List(ctx, ListTasksInput{ProjectID: &projectID, Status: &done})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].ID != b.ID {
		t.Fatalf("List() with status=DONE = %+v, want only task %d", results, b.ID)
	}

	all, err := svc.List(ctx, ListTasksInput{ProjectID: &projectID})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("List() without status filter returned %d tasks, want 2", len(all))
	}
	_ = a
}
