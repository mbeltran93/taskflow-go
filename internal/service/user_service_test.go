package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/auth"
)

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name    string
		input   RegisterUserInput
		seed    *RegisterUserInput // pre-existing user to force a conflict
		wantErr error
	}{
		{
			name:  "valid input creates a user with a hashed password",
			input: RegisterUserInput{Name: "Ada Lovelace", Email: "Ada@Example.com", Password: "supersecret"},
		},
		{
			name:    "blank name is rejected",
			input:   RegisterUserInput{Name: "  ", Email: "a@example.com", Password: "supersecret"},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			name:    "short password is rejected",
			input:   RegisterUserInput{Name: "Ada", Email: "a@example.com", Password: "short"},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			name:    "duplicate email is a conflict",
			seed:    &RegisterUserInput{Name: "First", Email: "dup@example.com", Password: "supersecret"},
			input:   RegisterUserInput{Name: "Second", Email: "dup@example.com", Password: "supersecret"},
			wantErr: apperror.ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			svc := NewUserService(repo)
			ctx := context.Background()

			if tc.seed != nil {
				if _, err := svc.Register(ctx, *tc.seed); err != nil {
					t.Fatalf("seed register failed: %v", err)
				}
			}

			u, err := svc.Register(ctx, tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Register() error = %v, want wrapping %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Register() unexpected error: %v", err)
			}
			if u.ID == 0 {
				t.Errorf("expected a generated ID, got 0")
			}
			if u.Email != "ada@example.com" {
				t.Errorf("email should be normalized to lowercase, got %q", u.Email)
			}
			if u.PasswordHash == tc.input.Password {
				t.Errorf("password must be hashed, not stored as plaintext")
			}
			if !auth.CheckPassword(u.PasswordHash, tc.input.Password) {
				t.Errorf("stored hash does not verify against the original password")
			}
		})
	}
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())

	_, err := svc.GetByID(context.Background(), 999)

	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, apperror.ErrNotFound)
	}
}
