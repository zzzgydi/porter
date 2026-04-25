package users

import (
	"context"
	"fmt"

	"github.com/porter/api/internal/password"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, email, name, plainPassword, role string) (*User, error) {
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &User{
		Email:        email,
		Name:         name,
		PasswordHash: hash,
		Role:         role,
	}
	if u.Role == "" {
		u.Role = "user"
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *Service) Authenticate(ctx context.Context, email, plainPassword string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !password.Check(plainPassword, u.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}
	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *Service) CountAdmins(ctx context.Context) (int64, error) {
	return s.repo.CountAdmins(ctx)
}

var ValidUserRoles = map[string]bool{"user": true, "platform_admin": true}

func (s *Service) Update(ctx context.Context, id, name, role string) error {
	if role != "" && !ValidUserRoles[role] {
		return fmt.Errorf("invalid role")
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	u.Name = name
	if role != "" {
		u.Role = role
	}
	return s.repo.Update(ctx, u)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
