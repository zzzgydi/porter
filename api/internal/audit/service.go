package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	repo   *Repo
	logger *slog.Logger
	queue  chan *Log
	wg     sync.WaitGroup
}

func NewService(repo *Repo, logger *slog.Logger) *Service {
	s := &Service{
		repo:   repo,
		logger: logger,
		queue:  make(chan *Log, 1000),
	}
	s.wg.Add(1)
	go s.worker()
	return s
}

func (s *Service) worker() {
	defer s.wg.Done()
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

// Shutdown closes the queue and waits for buffered events to be flushed,
// bounded by ctx.
func (s *Service) Shutdown(ctx context.Context) {
	close(s.queue)
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		if s.logger != nil {
			s.logger.Warn("audit shutdown timed out, some events may be lost")
		}
	}
}

func (s *Service) Log(ctx context.Context, actorType, actorID, action, target string, metadata map[string]any, ip, userAgent string) {
	var meta json.RawMessage
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
