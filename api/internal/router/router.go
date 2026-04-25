package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zzzgydi/porter/api/internal/audit"
	"github.com/zzzgydi/porter/api/internal/auth"
	"github.com/zzzgydi/porter/api/internal/session"
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

func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(httpx.RequestID)
	r.Use(httpx.Logger(deps.Logger))
	r.Use(httpx.Recoverer)
	r.Use(httpx.CORS(deps.Config.ConsoleOrigin()))

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

	// Repositories under project
	repoH := repositories.NewHandler(deps.RepoSvc, deps.ProjectsSvc, deps.MembersSvc, deps.SessionMgr)
	r.Get("/api/projects/{project}/repositories", repoH.ListByProject)
	r.Get("/api/projects/{project}/repositories/{repo}", repoH.Get)

	// Tags
	tagsH := tags.NewHandler(deps.TagSvc, deps.RepoSvc, deps.ProjectsSvc, deps.MembersSvc, deps.RegistryCl, deps.AuditSvc, deps.SessionMgr)
	r.Route("/api/projects/{project}/repositories/{repo}/tags", tagsH.Routes)

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
