package manifests

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Manifest struct {
	Digest      string          `json:"digest"`
	MediaType   string          `json:"media_type"`
	SizeBytes   int64           `json:"size_bytes"`
	ContentJSON json.RawMessage `json:"content_json,omitempty"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Upsert(ctx context.Context, m *Manifest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO manifests (digest, media_type, size_bytes, content_json)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (digest) DO UPDATE SET
			media_type = EXCLUDED.media_type,
			size_bytes = EXCLUDED.size_bytes,
			content_json = EXCLUDED.content_json,
			pushed_at = now()
	`, m.Digest, m.MediaType, m.SizeBytes, m.ContentJSON)
	return err
}
