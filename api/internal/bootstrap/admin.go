package bootstrap

import (
	"context"
	"log/slog"

	"github.com/zzzgydi/porter/api/internal/users"
)

func EnsureAdmin(ctx context.Context, usersSvc *users.Service, email, password string, logger *slog.Logger) {
	if email == "" || password == "" {
		logger.Warn("bootstrap admin skipped: missing email or password")
		return
	}

	count, err := usersSvc.Count(ctx)
	if err != nil {
		logger.Error("bootstrap admin: failed to count users", "error", err)
		return
	}
	if count > 0 {
		return
	}

	_, err = usersSvc.Create(ctx, email, "Admin", password, "platform_admin")
	if err != nil {
		logger.Error("bootstrap admin: failed to create", "error", err)
		return
	}
	logger.Info("bootstrap admin created", "email", email)
}
