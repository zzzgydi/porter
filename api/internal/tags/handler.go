package tags

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/audit"
	"github.com/zzzgydi/porter/api/internal/authz"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/projects"
	"github.com/zzzgydi/porter/api/internal/repositories"
	"github.com/zzzgydi/porter/api/internal/session"
)

type registryClient interface {
	DeleteManifest(ctx context.Context, repoName, digest string) error
}

type Handler struct {
	service     *Service
	repoSvc     *repositories.Service
	projectsSvc *projects.Service
	membersSvc  *members.Service
	registry    registryClient
	auditSvc    *audit.Service
	sessionMgr  *session.Manager
}

func NewHandler(
	service *Service,
	repoSvc *repositories.Service,
	projectsSvc *projects.Service,
	membersSvc *members.Service,
	registry registryClient,
	auditSvc *audit.Service,
	sessionMgr *session.Manager,
) *Handler {
	return &Handler{
		service:     service,
		repoSvc:     repoSvc,
		projectsSvc: projectsSvc,
		membersSvc:  membersSvc,
		registry:    registry,
		auditSvc:    auditSvc,
		sessionMgr:  sessionMgr,
	}
}

func (h *Handler) requireProjectMember(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireMember(h.membersSvc.Repo(), h.sessionMgr, r, projectID)
}

func (h *Handler) requireProjectOwner(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireRole(h.membersSvc.Repo(), h.sessionMgr, r, projectID, "owner")
}

// resolve looks up the repository and checks project membership.
func (h *Handler) resolve(w http.ResponseWriter, r *http.Request, repoName string, ownerOnly bool) (*repositories.Repository, bool) {
	if _, err := authz.RequireSession(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, err)
		return nil, false
	}
	project := chi.URLParam(r, "project")
	fullName := project + "/" + repoName
	repo, err := h.repoSvc.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return nil, false
	}
	if ownerOnly {
		if _, err := h.requireProjectOwner(r, repo.ProjectID); err != nil {
			httpx.JSONError(w, err)
			return nil, false
		}
	} else {
		if _, err := h.requireProjectMember(r, repo.ProjectID); err != nil {
			httpx.JSONError(w, err)
			return nil, false
		}
	}
	return repo, true
}

// List handles GET /api/projects/{project}/repositories/{repo...}/tags
func (h *Handler) List(w http.ResponseWriter, r *http.Request, repoName string) {
	repo, ok := h.resolve(w, r, repoName, false)
	if !ok {
		return
	}
	list, err := h.service.ListByRepo(r.Context(), repo.ID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// Get handles GET /api/projects/{project}/repositories/{repo...}/tags/{tag}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request, repoName, tagName string) {
	repo, ok := h.resolve(w, r, repoName, false)
	if !ok {
		return
	}
	tag, err := h.service.Get(r.Context(), repo.ID, tagName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("tag not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, tag)
}

// Delete handles DELETE /api/projects/{project}/repositories/{repo...}/tags/{tag}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request, repoName, tagName string) {
	project := chi.URLParam(r, "project")
	fullName := project + "/" + repoName

	repo, ok := h.resolve(w, r, repoName, true)
	if !ok {
		return
	}
	claims, _ := authz.RequireSession(h.sessionMgr, r)

	tag, err := h.service.Get(r.Context(), repo.ID, tagName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("tag not found"))
		return
	}

	// Delete from registry first; if it fails the DB record stays intact
	// so the user can retry.
	if err := h.registry.DeleteManifest(r.Context(), fullName, tag.Digest); err != nil {
		httpx.JSONError(w, httpx.Internal("registry delete failed"))
		return
	}

	if err := h.service.MarkDeleted(r.Context(), repo.ID, tagName); err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}

	h.auditSvc.Log(r.Context(), "user", claims.UserID, "tag.delete", fullName+":"+tagName, map[string]any{
		"digest": tag.Digest,
	}, httpx.ClientIP(r), r.UserAgent())

	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
