package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/porter/api/internal/audit"
	"github.com/porter/api/internal/session"
	"github.com/porter/api/internal/bootstrap"
	"github.com/porter/api/internal/config"
	"github.com/porter/api/internal/db"
	"github.com/porter/api/internal/log"
	"github.com/porter/api/internal/manifests"
	"github.com/porter/api/internal/members"
	"github.com/porter/api/internal/projects"
	"github.com/porter/api/internal/redisx"
	"github.com/porter/api/internal/registry"
	"github.com/porter/api/internal/registrytoken"
	"github.com/porter/api/internal/repositories"
	"github.com/porter/api/internal/robots"
	"github.com/porter/api/internal/router"
	"github.com/porter/api/internal/tags"
	"github.com/porter/api/internal/users"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load: %v\n", err)
		os.Exit(1)
	}

	logger := log.New()

	// Postgres
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		logger.Error("db migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("db migrated")

	// Redis
	redis := redisx.New(cfg.RedisAddr, cfg.RedisPassword)
	if err := redis.Ping(ctx); err != nil {
		logger.Error("redis connection failed", "error", err)
		os.Exit(1)
	}

	// Services
	usersRepo := users.NewRepo(pool)
	usersSvc := users.NewService(usersRepo)

	projectsRepo := projects.NewRepo(pool)
	membersRepo := members.NewRepo(pool)
	membersSvc := members.NewService(membersRepo, usersRepo)
	projectsSvc := projects.NewService(projectsRepo, membersRepo)

	robotsRepo := robots.NewRepo(pool)
	robotsSvc := robots.NewService(robotsRepo, projectsSvc)

	repoRepo := repositories.NewRepo(pool)
	repoSvc := repositories.NewService(repoRepo, projectsSvc)

	manifestsRepo := manifests.NewRepo(pool)
	manifestsSvc := manifests.NewService(manifestsRepo)

	tagsRepo := tags.NewRepo(pool)
	tagsSvc := tags.NewService(tagsRepo)

	auditRepo := audit.NewRepo(pool)
	auditSvc := audit.NewService(auditRepo)

	// Registry token signer
	signer, err := registrytoken.NewSigner(cfg.RegistryTokenKey)
	if err != nil {
		logger.Error("token signer init failed", "error", err)
		os.Exit(1)
	}
	signer.SetIssuer(cfg.RegistryTokenIssuer)

	// Generate JWKS for registry to trust
	jwksPath := "/certs/jwks.json"
	if _, err := os.Stat(jwksPath); os.IsNotExist(err) {
		certPath := "/certs/auth.crt"
		if _, err := os.Stat(certPath); err == nil {
			if err := registrytoken.GenerateJWKSFromCert(certPath, jwksPath); err != nil {
				logger.Warn("jwks generation failed", "error", err)
			} else {
				logger.Info("jwks generated", "path", jwksPath)
			}
		}
	}

	registryCl := registry.NewClient(cfg.RegistryInternalURL, signer, cfg.RegistryService)

	sessionMgr := session.NewManager(cfg.APIJWTSecret)

	// Bootstrap admin
	bootstrap.EnsureAdmin(ctx, usersSvc, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword, logger)

	// Router
	router := router.New(router.Deps{
		Config:      cfg,
		Logger:      logger,
		Redis:       redis,
		UsersSvc:    usersSvc,
		ProjectsSvc: projectsSvc,
		MembersSvc:  membersSvc,
		RobotsSvc:   robotsSvc,
		RepoSvc:     repoSvc,
		ManifestSvc: manifestsSvc,
		TagSvc:      tagsSvc,
		AuditSvc:    auditSvc,
		RegistryCl:  registryCl,
		Signer:      signer,
		SessionMgr:  sessionMgr,
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("server stopped")
}
