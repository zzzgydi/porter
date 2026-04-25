package audit

import (
	"context"
	"encoding/json"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, actorType, actorID, action, target string, metadata map[string]any, ip, userAgent string) {
	var meta []byte
	if metadata != nil {
		meta, _ = json.Marshal(metadata)
	}
	l := &Log{
		ActorType: actorType,
		ActorID:   actorID,
		Action:    action,
		Target:    target,
		Metadata:  meta,
		IP:        ip,
		UserAgent: userAgent,
	}
	_ = s.repo.Create(ctx, l)
}

func (s *Service) LogEvent(ctx context.Context, action, target string, metadata map[string]any) {
	s.Log(ctx, "system", "", action, target, metadata, "", "")
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Log, error) {
	return s.repo.List(ctx, limit, offset)
}
