package router

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/audit"
	"github.com/zzzgydi/porter/api/internal/auth"
	"github.com/zzzgydi/porter/api/internal/config"
	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/manifests"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/projects"
	"github.com/zzzgydi/porter/api/internal/redisx"
	"github.com/zzzgydi/porter/api/internal/registry"
	"github.com/zzzgydi/porter/api/internal/registrytoken"
	"github.com/zzzgydi/porter/api/internal/repositories"
	"github.com/zzzgydi/porter/api/internal/robots"
	"github.com/zzzgydi/porter/api/internal/session"
	"github.com/zzzgydi/porter/api/internal/tags"
	"github.com/zzzgydi/porter/api/internal/users"
)

type Deps struct {
	Config      *config.Config
	Logger      *slog.Logger
	Redis       *redisx.Client
	UsersSvc    *users.Service
	ProjectsSvc *projects.Service
	MembersSvc  *members.Service
	RobotsSvc   *robots.Service
	RepoSvc     *repositories.Service
	ManifestSvc *manifests.Service
	TagSvc      *tags.Service
	AuditSvc    *audit.Service
	RegistryCl  *registry.Client
	Signer      *registrytoken.Signer
	SessionMgr  *session.Manager
}

// sessionRefresh re-validates the session cookie against the database on every
// request. It injects claims with the user's *current* role, so that deleting
// or demoting a user takes effect immediately instead of at token expiry.
// Requests without a session cookie pass through untouched.
func sessionRefresh(usersSvc *users.Service, sessionMgr *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := sessionMgr.Parse(cookie.Value)
			if err == nil {
				u, err := usersSvc.GetByID(r.Context(), claims.UserID)
				if err != nil {
					// User deleted (or DB unavailable): mark the session invalid.
					r = r.WithContext(session.ContextWithClaims(r.Context(), nil))
				} else {
					fresh := &session.Claims{UserID: u.ID, Email: u.Email, Role: u.Role}
					r = r.WithContext(session.ContextWithClaims(r.Context(), fresh))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	if extractor, err := httpx.NewClientIPExtractor(deps.Config.TrustedProxies); err == nil {
		httpx.SetClientIPExtractor(extractor)
	}

	r.Use(httpx.RequestID)
	r.Use(httpx.Logger(deps.Logger))
	r.Use(httpx.Recoverer)
	r.Use(httpx.SecurityHeaders)
	r.Use(httpx.MaxBodySize(1 << 20)) // 1 MB
	r.Use(httpx.CORS(deps.Config.GetConsoleOrigin()))
	r.Use(sessionRefresh(deps.UsersSvc, deps.SessionMgr))

	// Public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	// Registry token endpoint (public, authenticated via Basic Auth)
	registryTokenH := registrytoken.NewHandler(
		deps.Signer, deps.UsersSvc, deps.RobotsSvc, deps.ProjectsSvc, deps.MembersSvc.Repo(), deps.Redis,
		deps.Config.RegistryTokenIssuer, deps.Config.RegistryService, deps.Config.TokenTTL(),
	)
	r.Get("/api/registry/token", registryTokenH.Token)

	// Auth
	authH := auth.NewHandler(deps.SessionMgr, deps.UsersSvc, deps.Redis, !deps.Config.DevMode)
	r.Route("/api/auth", authH.Routes)
	r.Get("/api/me", authH.Me)

	// Projects + Members
	projectsH := projects.NewHandler(deps.ProjectsSvc, deps.MembersSvc, deps.SessionMgr)
	r.Route("/api/projects", projectsH.Routes)

	// Repositories + Tags under a project. Repository names may contain
	// slashes (e.g. "team/app"), so everything after /repositories/ is a
	// wildcard dispatched by suffix:
	//   {repo}              -> repository detail (GET)
	//   {repo}/tags         -> tag list (GET)
	//   {repo}/tags/{tag}   -> tag detail (GET) / tag delete (DELETE)
	repoH := repositories.NewHandler(deps.RepoSvc, deps.ProjectsSvc, deps.MembersSvc, deps.SessionMgr)
	tagsH := tags.NewHandler(deps.TagSvc, deps.RepoSvc, deps.ProjectsSvc, deps.MembersSvc, deps.RegistryCl, deps.AuditSvc, deps.SessionMgr)
	r.Get("/api/projects/{project}/repositories", repoH.ListByProject)
	repoWild := func(w http.ResponseWriter, r *http.Request) {
		rest := strings.Trim(chi.URLParam(r, "*"), "/")
		if rest == "" {
			httpx.JSONError(w, httpx.NotFound("not found"))
			return
		}
		if strings.HasSuffix(rest, "/tags") {
			repoName := strings.TrimSuffix(rest, "/tags")
			if r.Method == http.MethodGet {
				tagsH.List(w, r, repoName)
				return
			}
		} else if i := strings.LastIndex(rest, "/tags/"); i >= 0 {
			repoName, tagName := rest[:i], rest[i+len("/tags/"):]
			if !strings.Contains(tagName, "/") {
				switch r.Method {
				case http.MethodGet:
					tagsH.Get(w, r, repoName, tagName)
					return
				case http.MethodDelete:
					tagsH.Delete(w, r, repoName, tagName)
					return
				}
			}
		}
		if r.Method == http.MethodGet {
			repoH.Get(w, r, rest)
			return
		}
		httpx.JSONError(w, httpx.NotFound("not found"))
	}
	r.Get("/api/projects/{project}/repositories/*", repoWild)
	r.Delete("/api/projects/{project}/repositories/*", repoWild)

	// Robot Tokens
	robotsH := robots.NewHandler(deps.RobotsSvc, deps.ProjectsSvc, deps.MembersSvc, deps.SessionMgr)
	r.Route("/api/robot-tokens", robotsH.Routes)

	// Users
	usersH := users.NewHandler(deps.UsersSvc, deps.SessionMgr)
	r.Route("/api/users", usersH.Routes)

	// Audit Logs
	auditH := audit.NewHandler(deps.AuditSvc, deps.SessionMgr)
	r.Route("/api/audit-logs", auditH.Routes)

	// Internal: Registry webhook
	eventsH := registry.NewEventHandler(deps.Redis, deps.RepoSvc, deps.ManifestSvc, deps.TagSvc, deps.AuditSvc, deps.Config.WebhookSecret, deps.Logger)
	r.Post("/internal/registry/events", eventsH.ServeHTTP)

	return r
}
