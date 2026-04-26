package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/redisx"
	"github.com/zzzgydi/porter/api/internal/session"
	"github.com/zzzgydi/porter/api/internal/users"
)

type Handler struct {
	sessionMgr *session.Manager
	usersSvc   *users.Service
	redis      *redisx.Client
	secure     bool
}

func NewHandler(sessionMgr *session.Manager, usersSvc *users.Service, redis *redisx.Client, secure bool) *Handler {
	return &Handler{
		sessionMgr: sessionMgr,
		usersSvc:   usersSvc,
		redis:      redis,
		secure:     secure,
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
		if errors.Is(err, users.ErrUserNotFound) {
			httpx.JSONError(w, httpx.Unauthorized("user not found"))
		} else {
			httpx.JSONError(w, httpx.Internal("db error"))
		}
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
	limitKey := "ratelimit:login:" + httpx.ClientIP(r)
	ok, err := h.redis.RateLimit(r.Context(), limitKey, 5, time.Minute)
	if err != nil || !ok {
		httpx.JSONError(w, httpx.TooManyRequests("too many login attempts"))
		return
	}

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
	// Prevent session fixation: clear any existing session before issuing a new one.
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
	ttl := 7 * 24 * time.Hour
	token, err := h.sessionMgr.Issue(u.ID, u.Email, u.Role, ttl)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("session error"))
		return
	}
	http.SetCookie(w, h.sessionMgr.Cookie(token, h.secure, ttl))
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
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
