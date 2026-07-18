package projects

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/authz"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/session"
)

var projectNameRe = regexp.MustCompile(`^[a-z0-9_-]+$`)

func validateProjectName(name string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || len(name) > 64 {
		return httpx.BadRequest("project name must be 1-64 characters")
	}
	if !projectNameRe.MatchString(name) {
		return httpx.BadRequest("project name may only contain lowercase letters, numbers, hyphens and underscores")
	}
	return nil
}

func validateProjectVisibility(v string) error {
	if v != "" && v != "private" && v != "public" {
		return httpx.BadRequest("visibility must be private or public")
	}
	return nil
}

type Handler struct {
	service    *Service
	membersSvc *members.Service
	sessionMgr *session.Manager
}

func NewHandler(service *Service, membersSvc *members.Service, sessionMgr *session.Manager) *Handler {
	return &Handler{service: service, membersSvc: membersSvc, sessionMgr: sessionMgr}
}

func (h *Handler) requireSession(r *http.Request) (*session.Claims, error) {
	return authz.RequireSession(h.sessionMgr, r)
}

func (h *Handler) requireProjectMember(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireMember(h.membersSvc.Repo(), h.sessionMgr, r, projectID)
}

func (h *Handler) requireProjectOwner(r *http.Request, projectID string) (*session.Claims, error) {
	return authz.RequireRole(h.membersSvc.Repo(), h.sessionMgr, r, projectID, "owner")
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
		httpx.JSONError(w, err)
		return
	}
	// Admins see every project, regular users only their memberships.
	if claims.Role == "platform_admin" {
		list, err := h.service.List(r.Context())
		if err != nil {
			httpx.JSONError(w, httpx.Internal("db error"))
			return
		}
		httpx.JSON(w, http.StatusOK, list)
		return
	}
	list, err := h.service.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// getProject looks up a project by name after verifying the session, so an
// unauthenticated caller cannot probe project existence via 404 vs 401.
func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) (*Project, bool) {
	if _, err := h.requireSession(r); err != nil {
		httpx.JSONError(w, err)
		return nil, false
	}
	name := chi.URLParam(r, "project")
	p, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("project not found"))
		return nil, false
	}
	return p, true
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, ok := h.getProject(w, r)
	if !ok {
		return
	}
	if _, err := h.requireProjectMember(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, err := h.requireSession(r)
	if err != nil {
		httpx.JSONError(w, err)
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
	if err := validateProjectName(req.Name); err != nil {
		httpx.JSONError(w, err)
		return
	}
	if err := validateProjectVisibility(req.Visibility); err != nil {
		httpx.JSONError(w, err)
		return
	}
	p, err := h.service.Create(r.Context(), req.Name, req.DisplayName, req.Visibility, claims.UserID)
	if err != nil {
		httpx.JSONError(w, httpx.Conflict("project already exists"))
		return
	}
	httpx.JSON(w, http.StatusCreated, p)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, ok := h.getProject(w, r)
	if !ok {
		return
	}
	if _, err := h.requireProjectOwner(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	var req struct {
		DisplayName *string `json:"display_name"`
		Visibility  string  `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	if err := validateProjectVisibility(req.Visibility); err != nil {
		httpx.JSONError(w, err)
		return
	}
	if err := h.service.Update(r.Context(), p.ID, req.DisplayName, req.Visibility); err != nil {
		httpx.JSONError(w, httpx.Internal("update failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p, ok := h.getProject(w, r)
	if !ok {
		return
	}
	if _, err := h.requireProjectOwner(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	if err := h.service.Delete(r.Context(), p.ID); err != nil {
		httpx.JSONError(w, httpx.Internal("delete failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	p, ok := h.getProject(w, r)
	if !ok {
		return
	}
	if _, err := h.requireProjectMember(r, p.ID); err != nil {
		httpx.JSONError(w, err)
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
	p, ok := h.getProject(w, r)
	if !ok {
		return
	}
	if _, err := h.requireProjectOwner(r, p.ID); err != nil {
		httpx.JSONError(w, err)
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
		switch {
		case errors.Is(err, members.ErrInvalidRole):
			httpx.JSONError(w, httpx.BadRequest("invalid role"))
		case errors.Is(err, members.ErrUserNotFound):
			httpx.JSONError(w, httpx.NotFound("user not found"))
		case errors.Is(err, members.ErrAlreadyMember):
			httpx.JSONError(w, httpx.Conflict("user is already a member with that role"))
		case errors.Is(err, members.ErrLastOwner):
			httpx.JSONError(w, httpx.BadRequest("cannot downgrade the last owner"))
		default:
			httpx.JSONError(w, httpx.Internal("add member failed"))
		}
		return
	}
	httpx.JSON(w, http.StatusCreated, m)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	p, ok := h.getProject(w, r)
	if !ok {
		return
	}
	if _, err := h.requireProjectOwner(r, p.ID); err != nil {
		httpx.JSONError(w, err)
		return
	}
	userID := chi.URLParam(r, "userId")
	if err := h.membersSvc.Remove(r.Context(), p.ID, userID); err != nil {
		switch {
		case errors.Is(err, members.ErrUserNotFound):
			httpx.JSONError(w, httpx.NotFound("member not found"))
		case errors.Is(err, members.ErrLastOwner):
			httpx.JSONError(w, httpx.BadRequest("cannot remove the last owner"))
		default:
			httpx.JSONError(w, httpx.Internal("remove failed"))
		}
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
