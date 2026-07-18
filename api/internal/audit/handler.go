package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/session"
)

type Handler struct {
	service    *Service
	sessionMgr *session.Manager
}

func NewHandler(service *Service, sessionMgr *session.Manager) *Handler {
	return &Handler{service: service, sessionMgr: sessionMgr}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, err := session.FromRequest(h.sessionMgr, r)
	if err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	if claims.Role != "platform_admin" {
		httpx.JSONError(w, httpx.Forbidden("admin required"))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	list, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}
