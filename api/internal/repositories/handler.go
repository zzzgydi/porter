package repositories

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/authz"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/projects"
	"github.com/zzzgydi/porter/api/internal/session"
)

type Handler struct {
	service     *Service
	projectsSvc *projects.Service
	membersSvc  *members.Service
	sessionMgr  *session.Manager
}

func NewHandler(service *Service, projectsSvc *projects.Service, membersSvc *members.Service, sessionMgr *session.Manager) *Handler {
	return &Handler{service: service, projectsSvc: projectsSvc, membersSvc: membersSvc, sessionMgr: sessionMgr}
}

func (h *Handler) requireProjectMember(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireMember(h.membersSvc.Repo(), h.sessionMgr, r, projectID)
}

func (h *Handler) ListByProject(w http.ResponseWriter, r *http.Request) {
	if _, err := authz.RequireSession(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.projectsSvc.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	if _, err := h.requireProjectMember(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	list, err := h.service.ListByProject(r.Context(), p.ID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// Get handles GET /api/projects/{project}/repositories/* where the wildcard
// is the (possibly nested) repository name.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request, repoName string) {
	if _, err := authz.RequireSession(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	project := chi.URLParam(r, "project")
	fullName := project + "/" + repoName
	repository, err := h.service.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	if _, err := h.requireProjectMember(r, repository.ProjectID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, repository)
}
