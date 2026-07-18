package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zzzgydi/porter/api/internal/db"
)

type Repository struct {
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Name        string        `json:"name"`
	FullName    string        `json:"full_name"`
	Description string        `json:"description"`
	CreatedAt   db.TimeString `json:"created_at"`
	UpdatedAt   db.TimeString `json:"updated_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Upsert(ctx context.Context, repo *Repository) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO repositories (project_id, name, full_name, description)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (full_name) DO UPDATE SET
			updated_at = now()
		RETURNING id, created_at, updated_at
	`, repo.ProjectID, repo.Name, repo.FullName, repo.Description).Scan(
		&repo.ID, &repo.CreatedAt, &repo.UpdatedAt)
}

func (r *Repo) GetByFullName(ctx context.Context, fullName string) (*Repository, error) {
	var repo Repository
	err := r.pool.QueryRow(ctx, `
		SELECT id, project_id, name, full_name, description, created_at, updated_at
		FROM repositories WHERE full_name = $1
	`, fullName).Scan(
		&repo.ID, &repo.ProjectID, &repo.Name, &repo.FullName, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *Repo) ListByProject(ctx context.Context, projectID string) ([]Repository, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, name, full_name, description, created_at, updated_at
		FROM repositories WHERE project_id = $1
		ORDER BY updated_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Repository, 0)
	for rows.Next() {
		var repo Repository
		if err := rows.Scan(
			&repo.ID, &repo.ProjectID, &repo.Name, &repo.FullName, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, repo)
	}
	return out, rows.Err()
}
