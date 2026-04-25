package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/httpx"
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
	if _, err := session.FromRequest(h.sessionMgr, r); err != nil {
		httpx.JSONError(w, httpx.Unauthorized("session required"))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("db error"))
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}
