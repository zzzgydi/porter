package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zzzgydi/porter/api/internal/audit"
	"github.com/zzzgydi/porter/api/internal/bootstrap"
	"github.com/zzzgydi/porter/api/internal/config"
	"github.com/zzzgydi/porter/api/internal/db"
	"github.com/zzzgydi/porter/api/internal/log"
	"github.com/zzzgydi/porter/api/internal/manifests"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/projects"
	"github.com/zzzgydi/porter/api/internal/redisx"
	"github.com/zzzgydi/porter/api/internal/registry"
	"github.com/zzzgydi/porter/api/internal/registrytoken"
	"github.com/zzzgydi/porter/api/internal/repositories"
	"github.com/zzzgydi/porter/api/internal/robots"
	"github.com/zzzgydi/porter/api/internal/router"
	"github.com/zzzgydi/porter/api/internal/session"
	"github.com/zzzgydi/porter/api/internal/tags"
	"github.com/zzzgydi/porter/api/internal/users"
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
	auditSvc := audit.NewService(auditRepo, logger)

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

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case <-sig:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	// Flush any buffered audit events before exiting.
	auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer auditCancel()
	auditSvc.Shutdown(auditCtx)

	logger.Info("server stopped")
}
