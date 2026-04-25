package robots

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zzzgydi/porter/api/internal/auth"
	"github.com/zzzgydi/porter/api/internal/projects"
)

type Service struct {
	repo        *Repo
	projectsSvc *projects.Service
}

func NewService(repo *Repo, projectsSvc *projects.Service) *Service {
	return &Service{repo: repo, projectsSvc: projectsSvc}
}

func (s *Service) Create(ctx context.Context, projectID, name string, perms map[string][]string) (*RobotToken, string, error) {
	p, err := s.projectsSvc.GetByID(ctx, projectID)
	if err != nil {
		return nil, "", fmt.Errorf("project not found: %w", err)
	}

	tokenRaw, err := auth.GenerateRandomToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	username := fmt.Sprintf("robot$%s-%s", p.Name, name)
	permsJSON, _ := json.Marshal(perms)

	t := &RobotToken{
		Name:      name,
		Username:  username,
		TokenHash: auth.HashRobotToken(tokenRaw),
		ProjectID: projectID,
		Permissions: permsJSON,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, "", fmt.Errorf("create robot token: %w", err)
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
		if err == nil && time.Now().After(exp) {
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
	hash := auth.HashRobotToken(password)
	if !auth.ConstantTimeEqual(hash, t.TokenHash) {
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
func (s *Service) ResolveProjectScope(t *RobotToken, scopeType, name string, requested []string) []string {
	if scopeType != "repository" {
		return nil
	}
	// name format: project/repo
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
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
		grantedSet[a] = struct{}{}
	}
	var result []string
	for _, a := range requested {
		if _, ok := grantedSet[a]; ok {
			result = append(result, a)
		}
	}
	return result
}
