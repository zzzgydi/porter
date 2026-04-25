package robots

import (
	"encoding/json"
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
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{id}", h.Revoke)
}

func (h *Handler) requireProjectMember(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireMember(h.membersSvc.Repo(), h.sessionMgr, r, projectID)
}

func (h *Handler) requireProjectDeveloper(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireRole(h.membersSvc.Repo(), h.sessionMgr, r, projectID, "developer", "owner")
}

func (h *Handler) requireProjectOwner(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireRole(h.membersSvc.Repo(), h.sessionMgr, r, projectID, "owner")
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
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
		return
	}
	// list all requires platform_admin
	if claims.Role != "platform_admin" {
		httpx.JSONError(w, httpx.Forbidden("admin required"))
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
	var req struct {
		ProjectID   string              `json:"project_id"`
		Name        string              `json:"name"`
		Permissions map[string][]string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	if _, err := h.requireProjectOwner(r, req.ProjectID); err != nil {
		httpx.JSONError(w, err)
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
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	id := chi.URLParam(r, "id")
	// fetch token to check project membership
	t, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("token not found"))
		return
	}
	if claims.Role != "platform_admin" {
		role, err := h.membersSvc.Repo().GetRole(r.Context(), t.ProjectID, claims.UserID)
		if err != nil {
			httpx.JSONError(w, httpx.Forbidden("not a project member"))
			return
		}
		if role != "owner" {
			httpx.JSONError(w, httpx.Forbidden("owner required"))
			return
		}
	}
	if err := h.service.Revoke(r.Context(), id); err != nil {
		httpx.JSONError(w, httpx.Internal("revoke failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
