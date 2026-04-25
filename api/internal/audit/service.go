package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

type Service struct {
	repo   *Repo
	logger *slog.Logger
	queue  chan *Log
}

func NewService(repo *Repo, logger *slog.Logger) *Service {
	s := &Service{
		repo:   repo,
		logger: logger,
		queue:  make(chan *Log, 1000),
	}
	go s.worker()
	return s
}

func (s *Service) worker() {
	for l := range s.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.repo.Create(ctx, l); err != nil {
			if s.logger != nil {
				s.logger.Warn("audit log creation failed", "action", l.Action, "target", l.Target, "error", err)
			}
		}
		cancel()
	}
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
	select {
	case s.queue <- l:
	default:
		if s.logger != nil {
			s.logger.Warn("audit log queue full, dropped", "action", action, "target", target)
		}
	}
}

func (s *Service) LogEvent(ctx context.Context, action, target string, metadata map[string]any) {
	s.Log(ctx, "system", "", action, target, metadata, "", "")
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Log, error) {
	return s.repo.List(ctx, limit, offset)
}
