package audit

import (
	"context"
	"encoding/json"
	"log/slog"
)

type Service struct {
	repo   *Repo
	logger *slog.Logger
}

func NewService(repo *Repo, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
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
	if err := s.repo.Create(ctx, l); err != nil {
		if s.logger != nil {
			s.logger.Warn("audit log creation failed", "action", action, "target", target, "error", err)
		}
	}
}

func (s *Service) LogEvent(ctx context.Context, action, target string, metadata map[string]any) {
	s.Log(ctx, "system", "", action, target, metadata, "", "")
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Log, error) {
	return s.repo.List(ctx, limit, offset)
}
