package members

import (
	"context"
	"fmt"

	"github.com/porter/api/internal/users"
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

func (s *Service) Add(ctx context.Context, projectID, email, role string) (*Member, error) {
	u, err := s.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
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
	return s.repo.Remove(ctx, projectID, userID)
}

func (s *Service) List(ctx context.Context, projectID string) ([]Member, error) {
	return s.repo.ListByProject(ctx, projectID)
}
