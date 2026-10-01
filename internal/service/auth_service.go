package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/auth"
	"github.com/mbeltran93/taskflow-go/internal/repository"
)

type LoginInput struct {
	Email    string
	Password string
}

// AuthService issues JWTs for valid email/password combinations.
type AuthService struct {
	users repository.UserRepository
	jwt   *auth.JWTManager
}

func NewAuthService(users repository.UserRepository, jwt *auth.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

// Login verifies the credentials and returns a signed JWT on success.
// Invalid email and invalid password both fail with apperror.ErrUnauthorized,
// so the response never reveals whether the email exists.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (string, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return "", fmt.Errorf("invalid credentials: %w", apperror.ErrUnauthorized)
		}
		return "", err
	}

	if !auth.CheckPassword(u.PasswordHash, in.Password) {
		return "", fmt.Errorf("invalid credentials: %w", apperror.ErrUnauthorized)
	}

	token, err := s.jwt.Generate(u.ID, u.Email)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}
