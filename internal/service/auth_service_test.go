package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/auth"
)

func TestAuthService_Login(t *testing.T) {
	users := newFakeUserRepo()
	userSvc := NewUserService(users)
	if _, err := userSvc.Register(context.Background(), RegisterUserInput{
		Name: "Grace Hopper", Email: "grace@example.com", Password: "supersecret",
	}); err != nil {
		t.Fatalf("seed register failed: %v", err)
	}

	jwtManager := auth.NewJWTManager("test-secret", time.Hour)
	authSvc := NewAuthService(users, jwtManager)

	tests := []struct {
		name    string
		input   LoginInput
		wantErr error
	}{
		{
			name:  "correct credentials return a verifiable token",
			input: LoginInput{Email: "grace@example.com", Password: "supersecret"},
		},
		{
			name:  "email is matched case-insensitively",
			input: LoginInput{Email: "GRACE@example.com", Password: "supersecret"},
		},
		{
			name:    "wrong password is unauthorized",
			input:   LoginInput{Email: "grace@example.com", Password: "wrongpassword"},
			wantErr: apperror.ErrUnauthorized,
		},
		{
			name:    "unknown email is unauthorized, not not-found",
			input:   LoginInput{Email: "ghost@example.com", Password: "supersecret"},
			wantErr: apperror.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := authSvc.Login(context.Background(), tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Login() error = %v, want wrapping %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Login() unexpected error: %v", err)
			}
			if token == "" {
				t.Fatalf("expected a non-empty token")
			}
			claims, err := jwtManager.Verify(token)
			if err != nil {
				t.Fatalf("issued token did not verify: %v", err)
			}
			if claims.Email != "grace@example.com" {
				t.Errorf("claims.Email = %q, want grace@example.com", claims.Email)
			}
		})
	}
}
