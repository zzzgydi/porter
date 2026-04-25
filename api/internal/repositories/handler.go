package repositories

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/httpx"
	"github.com/porter/api/internal/projects"
)

type Handler struct {
	service     *Service
	projectsSvc *projects.Service
	sessionMgr  *session.Manager
}

func NewHandler(service *Service, projectsSvc *projects.Service, sessionMgr *session.Manager) *Handler {
	return &Handler{service: service, projectsSvc: projectsSvc, sessionMgr: sessionMgr}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.ListByProject)
}

func (h *Handler) ListByProject(w http.ResponseWriter, r *http.Request) {
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.projectsSvc.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
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
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	project := chi.URLParam(r, "project")
	repo := chi.URLParam(r, "repo")
	fullName := project + "/" + repo
	repository, err := h.service.GetByFullName(r.Context(), fullName)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("repository not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, repository)
}
