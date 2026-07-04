package registrytoken

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/zzzgydi/porter/api/internal/auth"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/projects"
	"github.com/zzzgydi/porter/api/internal/redisx"
	"github.com/zzzgydi/porter/api/internal/robots"
	"github.com/zzzgydi/porter/api/internal/users"
)

type Handler struct {
	signer      *Signer
	usersSvc    *users.Service
	robotsSvc   *robots.Service
	projectsSvc *projects.Service
	membersRepo *members.Repo
	redis       *redisx.Client
	issuer      string
	service     string
	ttl         time.Duration
}

func NewHandler(
	signer *Signer,
	usersSvc *users.Service,
	robotsSvc *robots.Service,
	projectsSvc *projects.Service,
	membersRepo *members.Repo,
	redis *redisx.Client,
	issuer, service string,
	ttl time.Duration,
) *Handler {
	return &Handler{
		signer:      signer,
		usersSvc:    usersSvc,
		robotsSvc:   robotsSvc,
		projectsSvc: projectsSvc,
		membersRepo: membersRepo,
		redis:       redis,
		issuer:      issuer,
		service:     service,
		ttl:         ttl,
	}
}

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	limitKey := "ratelimit:token:" + httpx.ClientIP(r)
	ok, err := h.redis.RateLimit(r.Context(), limitKey, 10, time.Minute)
	if err != nil || !ok {
		httpx.JSONError(w, httpx.TooManyRequests("too many token requests"))
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
		httpx.JSONError(w, httpx.Unauthorized("basic auth required"))
		return
	}

	var subject string
	var userID string
	var userRole string
	var robot *robots.RobotToken
	var isRobot bool

	if auth.IsRobotUsername(username) {
		var err error
		robot, err = h.robotsSvc.Authenticate(r.Context(), username, password)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
			httpx.JSONError(w, httpx.Unauthorized("invalid credentials"))
			return
		}
		subject = robot.Username
		isRobot = true
	} else {
		u, err := h.usersSvc.Authenticate(r.Context(), username, password)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
			httpx.JSONError(w, httpx.Unauthorized("invalid credentials"))
			return
		}
		subject = u.Email
		userID = u.ID
		userRole = u.Role
	}

	service := r.URL.Query().Get("service")
	if service == "" {
		service = h.service
	}
	scopes := r.URL.Query()["scope"]

	var access []AccessEntry

	for _, scopeStr := range scopes {
		parsed, err := ParseScope(scopeStr)
		if err != nil {
			continue
		}
		var allowed []string
		if isRobot {
			allowed = h.robotsSvc.ResolveProjectScope(robot, parsed.Type, parsed.Name, parsed.Actions)
		} else {
			allowed = h.resolveUserScope(r.Context(), userID, userRole, parsed.Type, parsed.Name, parsed.Actions)
		}
		access = append(access, AccessEntry{
			Type:    parsed.Type,
			Name:    parsed.Name,
			Actions: allowed,
		})
	}

	token, err := h.signer.Sign(subject, service, access, h.ttl)
	if err != nil {
		httpx.JSONError(w, httpx.Internal("token signing failed"))
		return
	}

	httpx.RawJSON(w, http.StatusOK, map[string]any{
		"token":        token,
		"access_token": token,
		"expires_in":   int(h.ttl.Seconds()),
		"issued_at":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) resolveUserScope(ctx context.Context, userID, userRole, scopeType, name string, requested []string) []string {
	if scopeType != "repository" {
		return nil
	}
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil
	}
	projectName := parts[0]
	p, err := h.projectsSvc.GetByName(ctx, projectName)
	if err != nil {
		return nil
	}

	if userRole == "platform_admin" {
		return requested
	}

	role, err := h.membersRepo.GetRole(ctx, p.ID, userID)
	if err != nil {
		return nil
	}

	grantedSet := make(map[string]struct{})
	switch role {
	case "owner":
		grantedSet["pull"] = struct{}{}
		grantedSet["push"] = struct{}{}
		grantedSet["delete"] = struct{}{}
	case "developer":
		grantedSet["pull"] = struct{}{}
		grantedSet["push"] = struct{}{}
	case "guest":
		grantedSet["pull"] = struct{}{}
	}

	var result []string
	for _, a := range requested {
		if _, ok := grantedSet[a]; ok {
			result = append(result, a)
		}
	}
	return result
}
