package repositories

import (
	"context"
	"fmt"

	"github.com/porter/api/internal/projects"
)

type Service struct {
	repo        *Repo
	projectsSvc *projects.Service
}

func NewService(repo *Repo, projectsSvc *projects.Service) *Service {
	return &Service{repo: repo, projectsSvc: projectsSvc}
}

func (s *Service) Ensure(ctx context.Context, projectName, repoName string) (*Repository, error) {
	p, err := s.projectsSvc.GetByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	repo := &Repository{
		ProjectID: p.ID,
		Name:      repoName,
		FullName:  projectName + "/" + repoName,
	}
	if err := s.repo.Upsert(ctx, repo); err != nil {
		return nil, fmt.Errorf("upsert repository: %w", err)
	}
	return repo, nil
}

func (s *Service) GetByFullName(ctx context.Context, fullName string) (*Repository, error) {
	return s.repo.GetByFullName(ctx, fullName)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]Repository, error) {
	return s.repo.ListByProject(ctx, projectID)
}
