package projects

import (
	"context"
	"fmt"
	"strings"

	"github.com/porter/api/internal/members"
)

type Service struct {
	repo    *Repo
	memRepo *members.Repo
}

func NewService(repo *Repo, memRepo *members.Repo) *Service {
	return &Service{repo: repo, memRepo: memRepo}
}

func (s *Service) Create(ctx context.Context, name, displayName, visibility, creatorUserID string) (*Project, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	p := &Project{
		Name:        name,
		DisplayName: displayName,
		Visibility:  visibility,
	}
	if p.Visibility == "" {
		p.Visibility = "private"
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	// creator becomes owner
	if err := s.memRepo.Add(ctx, p.ID, creatorUserID, "owner"); err != nil {
		return nil, fmt.Errorf("add owner: %w", err)
	}
	return p, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (*Project, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Project, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]Project, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Update(ctx context.Context, id, displayName, visibility string) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	p.DisplayName = displayName
	if visibility != "" {
		p.Visibility = visibility
	}
	return s.repo.Update(ctx, p)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
