package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/httpx"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/users"
)

type Handler struct {
	sessionMgr *session.Manager
	usersSvc   *users.Service
}

func NewHandler(sessionMgr *session.Manager, usersSvc *users.Service) *Handler {
	return &Handler{
		sessionMgr: sessionMgr,
		usersSvc:   usersSvc,
	}
}

func (h *Handler) Routes(r chi.Router) {
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	u, err := h.usersSvc.GetByID(r.Context(), claims.UserID)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("user not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
		"role":  u.Role,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSONError(w, httpx.BadRequest("invalid body"))
		return
	}
	u, err := h.usersSvc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("invalid credentials"))
		return
	}
	token, err := h.sessionMgr.Issue(u.ID, u.Email, u.Role, 7*24*time.Hour)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("session error"))
		return
	}
	http.SetCookie(w, h.sessionMgr.Cookie(token))
	httpx.JSON(w, http.StatusOK, map[string]any{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
		"role":  u.Role,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
