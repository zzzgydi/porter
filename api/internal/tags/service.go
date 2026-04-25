package tags

import (
	"context"
	"database/sql"
	"fmt"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upsert(ctx context.Context, repoID, name, digest, mediaType string, sizeBytes int64, pushedBy string) (*Tag, error) {
	t := &Tag{
		RepositoryID: repoID,
		Name:         name,
		Digest:       digest,
		MediaType:    mediaType,
		SizeBytes:    sizeBytes,
	}
	if pushedBy != "" {
		t.PushedBy = sql.NullString{String: pushedBy, Valid: true}
	}
	if err := s.repo.Upsert(ctx, t); err != nil {
		return nil, fmt.Errorf("upsert tag: %w", err)
	}
	return t, nil
}

func (s *Service) ListByRepo(ctx context.Context, repoID string) ([]Tag, error) {
	return s.repo.ListByRepo(ctx, repoID)
}

func (s *Service) Get(ctx context.Context, repoID, name string) (*Tag, error) {
	return s.repo.Get(ctx, repoID, name)
}

func (s *Service) MarkDeleted(ctx context.Context, repoID, name string) error {
	return s.repo.MarkDeleted(ctx, repoID, name)
}

func (s *Service) MarkDeletedByDigest(ctx context.Context, repoID, digest string) error {
	return s.repo.MarkDeletedByDigest(ctx, repoID, digest)
}
