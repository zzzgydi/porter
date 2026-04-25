package robots

import (
	"encoding/json"
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
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{id}", h.Revoke)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	projectName := r.URL.Query().Get("project")
	if projectName != "" {
		p, err := h.projectsSvc.GetByName(r.Context(), projectName)
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
		return
	}
	list, err := h.service.ListAll(r.Context())
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	var req struct {
		ProjectID   string              `json:"project_id"`
		Name        string              `json:"name"`
		Permissions map[string][]string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	t, tokenRaw, err := h.service.Create(r.Context(), req.ProjectID, req.Name, req.Permissions)
	if err != nil {
		httpx.JSONError(w, httpx.BadRequest(err.Error()))
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"id":       t.ID,
		"username": t.Username,
		"token":    tokenRaw,
		"name":     t.Name,
	})
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.service.Revoke(r.Context(), id); err != nil {
		httpx.JSONError(w, httpx.Internal("revoke failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
