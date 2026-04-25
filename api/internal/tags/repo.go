package tags

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/porter/api/internal/db"
)

type Tag struct {
	ID           string         `json:"id"`
	RepositoryID string         `json:"repository_id"`
	Name         string         `json:"name"`
	Digest       string         `json:"digest"`
	MediaType    string         `json:"media_type"`
	SizeBytes    int64          `json:"size_bytes"`
	PushedBy     sql.NullString `json:"-"`
	PushedAt     db.TimeString         `json:"pushed_at"`
	UpdatedAt    db.TimeString         `json:"updated_at"`
	DeletedAt    *string        `json:"deleted_at,omitempty"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Upsert(ctx context.Context, t *Tag) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO tags (repository_id, name, digest, media_type, size_bytes, pushed_by, pushed_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (repository_id, name) DO UPDATE SET
			digest = EXCLUDED.digest,
			media_type = EXCLUDED.media_type,
			size_bytes = EXCLUDED.size_bytes,
			pushed_by = EXCLUDED.pushed_by,
			pushed_at = EXCLUDED.pushed_at,
			updated_at = now(),
			deleted_at = NULL
		RETURNING id, pushed_at, updated_at
	`, t.RepositoryID, t.Name, t.Digest, t.MediaType, t.SizeBytes, t.PushedBy).Scan(
		&t.ID, &t.PushedAt, &t.UpdatedAt)
}

func (r *Repo) Get(ctx context.Context, repoID, name string) (*Tag, error) {
	var t Tag
	err := r.pool.QueryRow(ctx, `
		SELECT id, repository_id, name, digest, media_type, size_bytes, pushed_by, pushed_at, updated_at, deleted_at
		FROM tags WHERE repository_id = $1 AND name = $2 AND deleted_at IS NULL
	`, repoID, name).Scan(
		&t.ID, &t.RepositoryID, &t.Name, &t.Digest, &t.MediaType, &t.SizeBytes, &t.PushedBy, &t.PushedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repo) ListByRepo(ctx context.Context, repoID string) ([]Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, repository_id, name, digest, media_type, size_bytes, pushed_by, pushed_at, updated_at, deleted_at
		FROM tags
		WHERE repository_id = $1 AND deleted_at IS NULL
		ORDER BY pushed_at DESC
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Tag, 0)
	for rows.Next() {
		var t Tag
		if err := rows.Scan(
			&t.ID, &t.RepositoryID, &t.Name, &t.Digest, &t.MediaType, &t.SizeBytes, &t.PushedBy, &t.PushedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repo) MarkDeleted(ctx context.Context, repoID, name string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tags SET deleted_at = now() WHERE repository_id = $1 AND name = $2
	`, repoID, name)
	return err
}

func (r *Repo) MarkDeletedByDigest(ctx context.Context, repoID, digest string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tags SET deleted_at = now() WHERE repository_id = $1 AND digest = $2
	`, repoID, digest)
	return err
}
