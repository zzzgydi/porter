package projects

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/porter/api/internal/db"
)

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Visibility  string `json:"visibility"`
	CreatedAt   db.TimeString `json:"created_at"`
	UpdatedAt   db.TimeString `json:"updated_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, p *Project) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO projects (name, display_name, visibility)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, p.Name, p.DisplayName, p.Visibility).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repo) GetByName(ctx context.Context, name string) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, display_name, visibility, created_at, updated_at
		FROM projects WHERE name = $1
	`, name).Scan(&p.ID, &p.Name, &p.DisplayName, &p.Visibility, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, display_name, visibility, created_at, updated_at
		FROM projects WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.DisplayName, &p.Visibility, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) List(ctx context.Context) ([]Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, display_name, visibility, created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Visibility, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repo) ListByUser(ctx context.Context, userID string) ([]Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, p.display_name, p.visibility, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Visibility, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repo) Update(ctx context.Context, p *Project) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE projects SET display_name = $1, visibility = $2, updated_at = now()
		WHERE id = $3
	`, p.DisplayName, p.Visibility, p.ID)
	return err
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}
