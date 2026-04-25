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

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.ListByProject)
}

func (h *Handler) requireProjectMember(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireMember(h.membersSvc.Repo(), h.sessionMgr, r, projectID)
}

func (h *Handler) ListByProject(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	repo := chi.URLParam(r, "repo")
	fullName := project + "/" + repo
	repository, err := h.service.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	p, err := h.projectsSvc.GetByID(r.Context(), repository.ProjectID)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	if _, err := h.requireProjectMember(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, repository)
}
