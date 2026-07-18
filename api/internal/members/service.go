package members

import (
	"context"
	"errors"
	"fmt"

	"github.com/zzzgydi/porter/api/internal/users"
)

var (
	ErrInvalidRole   = errors.New("invalid role")
	ErrUserNotFound  = errors.New("user not found")
	ErrLastOwner     = errors.New("cannot remove or downgrade the last owner")
	ErrAlreadyMember = errors.New("user is already a member with that role")
)

type Service struct {
	repo      *Repo
	usersRepo *users.Repo
}

func NewService(repo *Repo, usersRepo *users.Repo) *Service {
	return &Service{repo: repo, usersRepo: usersRepo}
}

func (s *Service) Repo() *Repo {
	return s.repo
}

var validRoles = map[string]bool{"owner": true, "developer": true, "guest": true}

func (s *Service) Add(ctx context.Context, projectID, email, role string) (*Member, error) {
	if !validRoles[role] {
		return nil, ErrInvalidRole
	}
	u, err := s.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	// Prevent downgrading the last owner
	existingRole, _ := s.repo.GetRole(ctx, projectID, u.ID)
	if existingRole == role {
		return nil, ErrAlreadyMember
	}
	if existingRole == "owner" && role != "owner" {
		owners, err := s.repo.CountOwners(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("count owners: %w", err)
		}
		if owners <= 1 {
			return nil, ErrLastOwner
		}
	}
	if err := s.repo.Add(ctx, projectID, u.ID, role); err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}
	return &Member{
		ProjectID: projectID,
		UserID:    u.ID,
		Role:      role,
		Email:     u.Email,
		Name:      u.Name,
	}, nil
}

func (s *Service) Remove(ctx context.Context, projectID, userID string) error {
	role, err := s.repo.GetRole(ctx, projectID, userID)
	if err != nil {
		if IsNotMember(err) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	if role == "owner" {
		owners, err := s.repo.CountOwners(ctx, projectID)
		if err != nil {
			return fmt.Errorf("count owners: %w", err)
		}
		if owners <= 1 {
			return ErrLastOwner
		}
	}
	return s.repo.Remove(ctx, projectID, userID)
}

func (s *Service) List(ctx context.Context, projectID string) ([]Member, error) {
	return s.repo.ListByProject(ctx, projectID)
}
