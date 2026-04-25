package tags

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/audit"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/httpx"
	"github.com/porter/api/internal/members"
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

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{tag}", h.Get)
	r.Delete("/{tag}", h.Delete)
}

func (h *Handler) requireProjectMember(r *http.Request, projectID string) (*session.Claims, error) {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	_, err = h.membersSvc.Repo().GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		return nil, httpx.Forbidden("not a project member")
	}
	return claims, nil
}

func (h *Handler) requireProjectDeveloper(r *http.Request, projectID string) (*session.Claims, error) {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	role, err := h.membersSvc.Repo().GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		return nil, httpx.Forbidden("not a project member")
	}
	if role != "developer" && role != "owner" {
		return nil, httpx.Forbidden("developer or owner required")
	}
	return claims, nil
}

func (h *Handler) requireProjectOwner(r *http.Request, projectID string) (*session.Claims, error) {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	role, err := h.membersSvc.Repo().GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		return nil, httpx.Forbidden("not a project member")
	}
	if role != "owner" {
		return nil, httpx.Forbidden("owner required")
	}
	return claims, nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	repoName := chi.URLParam(r, "repo")
	fullName := project + "/" + repoName
	repo, err := h.repoSvc.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	p, err := h.projectsSvc.GetByID(r.Context(), repo.ProjectID)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	if _, err := h.requireProjectMember(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	list, err := h.service.ListByRepo(r.Context(), repo.ID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	repoName := chi.URLParam(r, "repo")
	tagName := chi.URLParam(r, "tag")
	fullName := project + "/" + repoName

	repo, err := h.repoSvc.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	p, err := h.projectsSvc.GetByID(r.Context(), repo.ProjectID)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	if _, err := h.requireProjectMember(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	tag, err := h.service.Get(r.Context(), repo.ID, tagName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("tag not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, tag)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	repoName := chi.URLParam(r, "repo")
	tagName := chi.URLParam(r, "tag")
	fullName := project + "/" + repoName

	repo, err := h.repoSvc.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	p, err := h.projectsSvc.GetByID(r.Context(), repo.ProjectID)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	claims, err := h.requireProjectOwner(r, p.ID)
	if err != nil {
		httpx.JSONError(w, err)
		return
	}

	tag, err := h.service.Get(r.Context(), repo.ID, tagName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("tag not found"))
		return
	}

	if err := h.service.MarkDeleted(r.Context(), repo.ID, tagName); err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}

	if err := h.registry.DeleteManifest(r.Context(), fullName, tag.Digest); err != nil {
		httpx.JSONError(w, httpx.Internal("registry delete failed: "+err.Error()))
		return
	}

	h.auditSvc.Log(r.Context(), "user", claims.UserID, "tag.delete", fullName+":"+tagName, map[string]any{
		"digest": tag.Digest,
	}, httpx.ClientIP(r), r.UserAgent())

	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
