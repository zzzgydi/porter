package robots

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zzzgydi/porter/api/internal/auth"
	"github.com/zzzgydi/porter/api/internal/projects"
)

var robotNameRe = regexp.MustCompile(`^[a-z0-9_-]+$`)

var validActions = map[string]bool{"pull": true, "push": true, "delete": true}

type Service struct {
	repo        *Repo
	projectsSvc *projects.Service
}

func NewService(repo *Repo, projectsSvc *projects.Service) *Service {
	return &Service{repo: repo, projectsSvc: projectsSvc}
}

// validatePermissions ensures a robot token can only hold permissions on its
// own project, and only known registry actions.
func validatePermissions(projectName string, perms map[string][]string) error {
	if len(perms) == 0 {
		return fmt.Errorf("permissions are required")
	}
	wantKey := projectName + "/*"
	for key, actions := range perms {
		if key != wantKey {
			return fmt.Errorf("invalid permission scope %q: only %q is allowed", key, wantKey)
		}
		if len(actions) == 0 {
			return fmt.Errorf("permission scope %q must grant at least one action", key)
		}
		for _, a := range actions {
			if !validActions[a] {
				return fmt.Errorf("invalid action %q: allowed actions are pull, push, delete", a)
			}
		}
	}
	return nil
}

func (s *Service) Create(ctx context.Context, projectID, name string, perms map[string][]string) (*RobotToken, string, error) {
	p, err := s.projectsSvc.GetByID(ctx, projectID)
	if err != nil {
		return nil, "", fmt.Errorf("project not found")
	}

	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || len(name) > 32 || !robotNameRe.MatchString(name) {
		return nil, "", fmt.Errorf("robot name must be 1-32 characters of lowercase letters, numbers, hyphens and underscores")
	}
	if err := validatePermissions(p.Name, perms); err != nil {
		return nil, "", err
	}

	tokenRaw, err := auth.GenerateRandomToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	username := fmt.Sprintf("robot$%s-%s", p.Name, name)
	permsJSON, _ := json.Marshal(perms)

	tokenHash := auth.HashRobotToken(tokenRaw)
	if tokenHash == "" {
		return nil, "", fmt.Errorf("hash robot token failed")
	}
	t := &RobotToken{
		Name:        name,
		Username:    username,
		TokenHash:   tokenHash,
		ProjectID:   projectID,
		Permissions: permsJSON,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, "", fmt.Errorf("create robot token failed (name may already exist)")
	}
	return t, tokenRaw, nil
}

func (s *Service) GetByUsername(ctx context.Context, username string) (*RobotToken, error) {
	t, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if t.RevokedAt != "" {
		return nil, fmt.Errorf("token revoked")
	}
	if t.ExpiresAt != "" {
		exp, err := time.Parse(time.RFC3339, string(t.ExpiresAt))
		if err != nil {
			// Fail closed: an unreadable expiry must not grant access.
			return nil, fmt.Errorf("token expiry invalid")
		}
		if time.Now().After(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}
	return t, nil
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (*RobotToken, error) {
	t, err := s.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !auth.CheckRobotToken(password, t.TokenHash) {
		return nil, fmt.Errorf("invalid credentials")
	}
	return t, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*RobotToken, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]RobotToken, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) ListAll(ctx context.Context) ([]RobotToken, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) Revoke(ctx context.Context, id string) error {
	return s.repo.Revoke(ctx, id)
}

// ResolveProjectScope returns granted actions for a robot token against a scope.
// The token can only ever access repositories inside its own project.
func (s *Service) ResolveProjectScope(ctx context.Context, t *RobotToken, scopeType, name string, requested []string) []string {
	if scopeType != "repository" {
		return nil
	}
	// name format: project/repo
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil
	}
	p, err := s.projectsSvc.GetByID(ctx, t.ProjectID)
	if err != nil || p.Name != parts[0] {
		return nil
	}
	var perms map[string][]string
	_ = json.Unmarshal(t.Permissions, &perms)
	pattern := parts[0] + "/*"
	granted, ok := perms[pattern]
	if !ok {
		return nil
	}
	grantedSet := make(map[string]struct{})
	for _, a := range granted {
		if validActions[a] {
			grantedSet[a] = struct{}{}
		}
	}
	var result []string
	for _, a := range requested {
		if _, ok := grantedSet[a]; ok {
			result = append(result, a)
		}
	}
	return result
}
