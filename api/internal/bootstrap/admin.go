package bootstrap

import (
	"context"
	"log/slog"

	"github.com/porter/api/internal/users"
)

func EnsureAdmin(ctx context.Context, usersSvc *users.Service, email, password string, logger *slog.Logger) {
	if email == "" || password == "" {
		logger.Warn("bootstrap admin skipped: missing email or password")
		return
	}

	list, err := usersSvc.List(ctx)
	if err != nil {
		logger.Error("bootstrap admin: failed to list users", "error", err)
		return
	}
	if len(list) > 0 {
		return
	}

	_, err = usersSvc.Create(ctx, email, "Admin", password, "platform_admin")
	if err != nil {
		logger.Error("bootstrap admin: failed to create", "error", err)
		return
	}
	logger.Info("bootstrap admin created", "email", email)
}
