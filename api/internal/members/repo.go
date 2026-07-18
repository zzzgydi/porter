package members

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zzzgydi/porter/api/internal/db"
)

type Member struct {
	ID        string        `json:"id"`
	ProjectID string        `json:"project_id"`
	UserID    string        `json:"user_id"`
	Role      string        `json:"role"`
	CreatedAt db.TimeString `json:"created_at"`
	Email     string        `json:"email,omitempty"`
	Name      string        `json:"name,omitempty"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Add(ctx context.Context, projectID, userID, role string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3
	`, projectID, userID, role)
	return err
}

func (r *Repo) Remove(ctx context.Context, projectID, userID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	return err
}

func (r *Repo) ListByProject(ctx context.Context, projectID string) ([]Member, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.created_at, u.email, u.name
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY pm.created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Member, 0)
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.CreatedAt, &m.Email, &m.Name); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ErrNotMember is returned by GetRole when the user has no membership row.
var ErrNotMember = errors.New("not a member")

// IsNotMember reports whether err means the user is not a project member.
func IsNotMember(err error) bool {
	return errors.Is(err, ErrNotMember)
}

func (r *Repo) GetRole(ctx context.Context, projectID, userID string) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx, `
		SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotMember
		}
		return "", fmt.Errorf("db error: %w", err)
	}
	return role, nil
}

func (r *Repo) CountOwners(ctx context.Context, projectID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM project_members WHERE project_id = $1 AND role = 'owner'
	`, projectID).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *Repo) IsOwner(ctx context.Context, projectID, userID string) (bool, error) {
	role, err := r.GetRole(ctx, projectID, userID)
	if err != nil {
		return false, err
	}
	return role == "owner", nil
}
