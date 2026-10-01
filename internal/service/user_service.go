// Package service holds the business logic of taskflow-go. Services depend
// only on repository interfaces (never on *sqlx.DB directly), which is what
// lets the unit tests in this package run against in-memory fakes instead of
// a real database.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/auth"
	"github.com/mbeltran93/taskflow-go/internal/model"
	"github.com/mbeltran93/taskflow-go/internal/repository"
)

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

// UserService implements registration and read access to user accounts.
//
// POST /api/users (registration) is not part of the portfolio's original
// endpoint list, but without it there would be no way to create an account
// to log in with, so it is public on purpose - same convention used by the
// other backends in this portfolio.
type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Register(ctx context.Context, in RegisterUserInput) (*model.User, error) {
	name := strings.TrimSpace(in.Name)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	if name == "" || email == "" || len(in.Password) < 8 {
		return nil, fmt.Errorf("name and email are required, password must be at least 8 characters: %w", apperror.ErrInvalidInput)
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{Name: name, Email: email, PasswordHash: hash}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.users.List(ctx)
}
