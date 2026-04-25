package tags

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/httpx"
	"github.com/porter/api/internal/projects"
	"github.com/porter/api/internal/repositories"
)

type registryClient interface {
	DeleteManifest(ctx context.Context, repoName, digest string) error
}

type Handler struct {
	service     *Service
	repoSvc     *repositories.Service
	projectsSvc *projects.Service
	registry    registryClient
	sessionMgr  *session.Manager
}

func NewHandler(
	service *Service,
	repoSvc *repositories.Service,
	projectsSvc *projects.Service,
	registry registryClient,
	sessionMgr *session.Manager,
) *Handler {
	return &Handler{
		service:     service,
		repoSvc:     repoSvc,
		projectsSvc: projectsSvc,
		registry:    registry,
		sessionMgr:  sessionMgr,
	}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Delete("/{tag}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	project := chi.URLParam(r, "project")
	repoName := chi.URLParam(r, "repo")
	fullName := project + "/" + repoName
	repo, err := h.repoSvc.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	list, err := h.service.ListByRepo(r.Context(), repo.ID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	project := chi.URLParam(r, "project")
	repoName := chi.URLParam(r, "repo")
	tagName := chi.URLParam(r, "tag")
	fullName := project + "/" + repoName

	repo, err := h.repoSvc.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}

	tag, err := h.service.Get(r.Context(), repo.ID, tagName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("tag not found"))
		return
	}

	if err := h.registry.DeleteManifest(r.Context(), fullName, tag.Digest); err != nil {
		httpx.JSONError(w, httpx.Internal("registry delete failed: "+err.Error()))
		return
	}

	if err := h.service.MarkDeleted(r.Context(), repo.ID, tagName); err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
