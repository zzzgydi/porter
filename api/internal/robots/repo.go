package robots

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zzzgydi/porter/api/internal/db"
)

type RobotToken struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Username    string          `json:"username"`
	TokenHash   string          `json:"-"`
	ProjectID   string          `json:"project_id,omitempty"`
	Permissions json.RawMessage `json:"permissions"`
	ExpiresAt   db.TimeString   `json:"expires_at,omitempty"`
	CreatedAt   db.TimeString   `json:"created_at"`
	RevokedAt   db.TimeString   `json:"revoked_at,omitempty"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, t *RobotToken) error {
	expiresAt := any(nil)
	if t.ExpiresAt != "" {
		expiresAt = string(t.ExpiresAt)
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO robot_tokens (name, username, token_hash, project_id, permissions, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, t.Name, t.Username, t.TokenHash, t.ProjectID, t.Permissions, expiresAt).Scan(&t.ID, &t.CreatedAt)
}

func (r *Repo) GetByUsername(ctx context.Context, username string) (*RobotToken, error) {
	var t RobotToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, username, token_hash, project_id, permissions, expires_at, created_at, revoked_at
		FROM robot_tokens WHERE username = $1
	`, username).Scan(&t.ID, &t.Name, &t.Username, &t.TokenHash, &t.ProjectID, &t.Permissions, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repo) ListByProject(ctx context.Context, projectID string) ([]RobotToken, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, username, project_id, permissions, expires_at, created_at, revoked_at
		FROM robot_tokens WHERE project_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RobotToken, 0)
	for rows.Next() {
		var t RobotToken
		if err := rows.Scan(&t.ID, &t.Name, &t.Username, &t.ProjectID, &t.Permissions, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repo) ListAll(ctx context.Context) ([]RobotToken, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, username, project_id, permissions, expires_at, created_at, revoked_at
		FROM robot_tokens WHERE revoked_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RobotToken, 0)
	for rows.Next() {
		var t RobotToken
		if err := rows.Scan(&t.ID, &t.Name, &t.Username, &t.ProjectID, &t.Permissions, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repo) GetByID(ctx context.Context, id string) (*RobotToken, error) {
	var t RobotToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, username, token_hash, project_id, permissions, expires_at, created_at, revoked_at
		FROM robot_tokens WHERE id = $1
	`, id).Scan(&t.ID, &t.Name, &t.Username, &t.TokenHash, &t.ProjectID, &t.Permissions, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repo) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE robot_tokens SET revoked_at = now() WHERE id = $1
	`, id)
	return err
}
