package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
)

func newProjectServiceWithOwner(t *testing.T) (*ProjectService, int64) {
	t.Helper()
	users := newFakeUserRepo()
	owner, err := NewUserService(users).Register(context.Background(), RegisterUserInput{
		Name: "Owner", Email: "owner@example.com", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("seed owner failed: %v", err)
	}
	return NewProjectService(newFakeProjectRepo(), users), owner.ID
}

func TestProjectService_Create(t *testing.T) {
	tests := []struct {
		name       string
		buildInput func(ownerID int64) CreateProjectInput
		wantErr    error
	}{
		{
			name: "valid project is created",
			buildInput: func(ownerID int64) CreateProjectInput {
				return CreateProjectInput{Name: "Launch", Description: "Q4 launch", OwnerID: ownerID}
			},
		},
		{
			name: "blank name is rejected",
			buildInput: func(ownerID int64) CreateProjectInput {
				return CreateProjectInput{Name: "  ", OwnerID: ownerID}
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			name: "missing owner is rejected",
			buildInput: func(int64) CreateProjectInput {
				return CreateProjectInput{Name: "Launch"}
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			name: "unknown owner fails as not found",
			buildInput: func(int64) CreateProjectInput {
				return CreateProjectInput{Name: "Launch", OwnerID: 999}
			},
			wantErr: apperror.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, ownerID := newProjectServiceWithOwner(t)

			p, err := svc.Create(context.Background(), tc.buildInput(ownerID))

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Create() error = %v, want wrapping %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if p.ID == 0 {
				t.Errorf("expected a generated ID, got 0")
			}
		})
	}
}

func TestProjectService_UpdateAndDelete(t *testing.T) {
	svc, ownerID := newProjectServiceWithOwner(t)
	ctx := context.Background()

	p, err := svc.Create(ctx, CreateProjectInput{Name: "Launch", OwnerID: ownerID})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	updated, err := svc.Update(ctx, p.ID, UpdateProjectInput{Name: "Launch v2", Description: "renamed"})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if updated.Name != "Launch v2" {
		t.Errorf("Name = %q, want %q", updated.Name, "Launch v2")
	}

	if err := svc.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	if _, err := svc.GetByID(ctx, p.ID); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want %v", err, apperror.ErrNotFound)
	}

	if _, err := svc.Update(ctx, p.ID, UpdateProjectInput{Name: "irrelevant"}); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("Update() on deleted project error = %v, want %v", err, apperror.ErrNotFound)
	}

	if err := svc.Delete(ctx, p.ID); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("Delete() twice error = %v, want %v", err, apperror.ErrNotFound)
	}
}
