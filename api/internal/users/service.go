package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/zzzgydi/porter/api/internal/password"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrLastAdmin    = errors.New("cannot demote the last admin")
	ErrInvalidRole  = errors.New("invalid role")
)

// dummyHash is a precomputed bcrypt hash used to equalize response timing
// when the account does not exist (prevents user enumeration).
var dummyHash = "$2a$12$VA/r6h.K3IIk6K5.D1766eOoTc21dLyKlJKNE3Cf9iASrLTL79Qo6"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// normalizeEmail makes account lookups case-insensitive.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) Create(ctx context.Context, email, name, plainPassword, role string) (*User, error) {
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &User{
		Email:        normalizeEmail(email),
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
	u, err := s.repo.GetByEmail(ctx, normalizeEmail(email))
	if err != nil {
		// Spend a bcrypt round anyway so unknown emails are indistinguishable.
		password.Check(plainPassword, dummyHash)
		return nil, fmt.Errorf("invalid credentials")
	}
	if !password.Check(plainPassword, u.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}
	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
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

// Update changes a user's mutable fields. name == nil leaves the current
// value untouched; role == "" leaves it untouched.
func (s *Service) Update(ctx context.Context, id string, name *string, role string) error {
	if role != "" && !ValidUserRoles[role] {
		return ErrInvalidRole
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	// Prevent demoting the last platform admin
	if role != "" && role != "platform_admin" && u.Role == "platform_admin" {
		admins, err := s.repo.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}
	if name != nil {
		u.Name = *name
	}
	if role != "" {
		u.Role = role
	}
	return s.repo.Update(ctx, u)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
