package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/porter/api/internal/audit"
	"github.com/porter/api/internal/manifests"
	"github.com/porter/api/internal/redisx"
	"github.com/porter/api/internal/repositories"
	"github.com/porter/api/internal/tags"
)

type EventPayload struct {
	Events []Event `json:"events"`
}

type Event struct {
	ID        string      `json:"id"`
	Timestamp string      `json:"timestamp"`
	Action    string      `json:"action"`
	Target    EventTarget `json:"target"`
	Request   EventRequest `json:"request"`
}

type EventTarget struct {
	MediaType  string `json:"mediaType"`
	Digest     string `json:"digest"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       int64  `json:"size"`
}

type EventRequest struct {
	Addr      string `json:"addr"`
	UserAgent string `json:"useragent"`
}

type EventHandler struct {
	redis         *redisx.Client
	repoSvc       *repositories.Service
	manifestSvc   *manifests.Service
	tagSvc        *tags.Service
	auditSvc      *audit.Service
	webhookSecret string
	logger        *slog.Logger
}

func NewEventHandler(redis *redisx.Client, repoSvc *repositories.Service, manifestSvc *manifests.Service, tagSvc *tags.Service, auditSvc *audit.Service, secret string, logger *slog.Logger) *EventHandler {
	return &EventHandler{
		redis:         redis,
		repoSvc:       repoSvc,
		manifestSvc:   manifestSvc,
		tagSvc:        tagSvc,
		auditSvc:      auditSvc,
		webhookSecret: secret,
		logger:        logger,
	}
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	expected := "Bearer " + h.webhookSecret
	if authHeader != expected {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload EventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	for _, ev := range payload.Events {
		if err := h.handleEvent(r.Context(), ev); err != nil {
			h.logger.Warn("webhook event failed", "event_id", ev.ID, "action", ev.Action, "error", err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *EventHandler) handleEvent(ctx context.Context, ev Event) error {
	// Deduplicate via Redis
	ok, err := h.redis.MarkWebhookEvent(ctx, ev.ID, 24*time.Hour)
	if err != nil || !ok {
		return nil // duplicate or error
	}

	parts := strings.SplitN(ev.Target.Repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name: %s", ev.Target.Repository)
	}
	projectName := parts[0]
	repoName := parts[1]

	repo, err := h.repoSvc.Ensure(ctx, projectName, repoName)
	if err != nil {
		return fmt.Errorf("ensure repository: %w", err)
	}

	switch ev.Action {
	case "push":
		// Upsert manifest first (tags FK references manifests)
		if err := h.manifestSvc.Upsert(ctx, repo.ID, ev.Target.Digest, ev.Target.MediaType, ev.Target.Size, nil); err != nil {
			return fmt.Errorf("upsert manifest: %w", err)
		}
		if ev.Target.Tag != "" {
			_, err = h.tagSvc.Upsert(ctx, repo.ID, ev.Target.Tag, ev.Target.Digest, ev.Target.MediaType, ev.Target.Size, "")
			if err != nil {
				return fmt.Errorf("upsert tag: %w", err)
			}
		}
		h.auditSvc.LogEvent(ctx, "registry.push", ev.Target.Repository+":"+ev.Target.Tag, map[string]any{
			"digest": ev.Target.Digest,
			"size":   ev.Target.Size,
		})

	case "delete":
		if ev.Target.Tag != "" {
			_ = h.tagSvc.MarkDeleted(ctx, repo.ID, ev.Target.Tag)
		} else {
			_ = h.tagSvc.MarkDeletedByDigest(ctx, repo.ID, ev.Target.Digest)
		}
		h.auditSvc.LogEvent(ctx, "registry.delete", ev.Target.Repository, map[string]any{
			"digest": ev.Target.Digest,
		})
	}

	return nil
}
