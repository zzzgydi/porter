package manifests

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upsert(ctx context.Context, repoID, digest, mediaType string, sizeBytes int64, contentJSON []byte) error {
	m := &Manifest{
		Digest:       digest,
		RepositoryID: repoID,
		MediaType:    mediaType,
		SizeBytes:    sizeBytes,
		ContentJSON:  contentJSON,
	}
	if err := s.repo.Upsert(ctx, m); err != nil {
		return fmt.Errorf("upsert manifest: %w", err)
	}
	return nil
}
