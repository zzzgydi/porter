package projects

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/httpx"
	"github.com/porter/api/internal/members"
)

type Handler struct {
	service     *Service
	membersSvc  *members.Service
	sessionMgr  *session.Manager
}

func NewHandler(service *Service, membersSvc *members.Service, sessionMgr *session.Manager) *Handler {
	return &Handler{service: service, membersSvc: membersSvc, sessionMgr: sessionMgr}
}

func (h *Handler) requireSession(r *http.Request) (*session.Claims, error) {
	return session.FromRequest(h.sessionMgr, r)
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{project}", h.Get)
	r.Patch("/{project}", h.Update)
	r.Delete("/{project}", h.Delete)

	r.Get("/{project}/members", h.ListMembers)
	r.Post("/{project}/members", h.AddMember)
	r.Delete("/{project}/members/{userId}", h.RemoveMember)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, err := h.requireSession(r)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	list, err := h.service.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, err := h.requireSession(r)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Visibility  string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	p, err := h.service.Create(r.Context(), req.Name, req.DisplayName, req.Visibility, claims.UserID)
	if err != nil {
		httpx.JSONError(w, httpx.BadRequest(err.Error()))
		return
	}
	httpx.JSON(w, http.StatusCreated, p)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	var req struct {
		DisplayName string `json:"display_name"`
		Visibility  string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	if err := h.service.Update(r.Context(), p.ID, req.DisplayName, req.Visibility); err != nil {
		httpx.JSONError(w, httpx.Internal("update failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	if err := h.service.Delete(r.Context(), p.ID); err != nil {
		httpx.JSONError(w, httpx.Internal("delete failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	list, err := h.membersSvc.List(r.Context(), p.ID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	m, err := h.membersSvc.Add(r.Context(), p.ID, req.Email, req.Role)
	if err != nil {
		httpx.JSONError(w, httpx.BadRequest(err.Error()))
		return
	}
	httpx.JSON(w, http.StatusCreated, m)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return
	}
	userID := chi.URLParam(r, "userId")
	if err := h.membersSvc.Remove(r.Context(), p.ID, userID); err != nil {
		httpx.JSONError(w, httpx.Internal("remove failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
