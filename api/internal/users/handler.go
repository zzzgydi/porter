package users

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/httpx"
)

var validUserRoles = map[string]bool{"user": true, "platform_admin": true}

func validateUserInput(email, password, role string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return httpx.BadRequest("invalid email")
	}
	if len(password) < 8 {
		return httpx.BadRequest("password must be at least 8 characters")
	}
	if role != "" && !validUserRoles[role] {
		return httpx.BadRequest("invalid role")
	}
	return nil
}

type Handler struct {
	service    *Service
	sessionMgr *session.Manager
}

func NewHandler(service *Service, sessionMgr *session.Manager) *Handler {
	return &Handler{service: service, sessionMgr: sessionMgr}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

func (h *Handler) requireAdmin(r *http.Request) error {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		return err
	}
	if claims.Role != "platform_admin" {
		return httpx.Forbidden("admin required")
	}
	return nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	list, err := h.service.List(r.Context())
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	u, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpx.JSONError(w, httpx.NotFound("user not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, u)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	if err := validateUserInput(req.Email, req.Password, req.Role); err != nil {
		httpx.JSONError(w, err)
		return
	}
	u, err := h.service.Create(r.Context(), req.Email, req.Name, req.Password, req.Role)
	if err != nil {
		httpx.JSONError(w, httpx.BadRequest(err.Error()))
		return
	}
	httpx.JSON(w, http.StatusCreated, u)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	if err := h.service.Update(r.Context(), id, req.Name, req.Role); err != nil {
		httpx.JSONError(w, httpx.Internal("update failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		httpx.JSONError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		httpx.JSONError(w, httpx.Internal("delete failed"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
