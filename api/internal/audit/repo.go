package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zzzgydi/porter/api/internal/db"
)

type Log struct {
	ID        string          `json:"id"`
	ActorType string          `json:"actor_type"`
	ActorID   string          `json:"actor_id"`
	Action    string          `json:"action"`
	Target    string          `json:"target"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	IP        string          `json:"ip"`
	UserAgent string          `json:"user_agent"`
	CreatedAt db.TimeString   `json:"created_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, l *Log) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO audit_logs (actor_type, actor_id, action, target, metadata, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, l.ActorType, l.ActorID, l.Action, l.Target, l.Metadata, l.IP, l.UserAgent).Scan(
		&l.ID, &l.CreatedAt)
}

func (r *Repo) List(ctx context.Context, limit, offset int) ([]Log, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, actor_type, actor_id, action, target, metadata, ip, user_agent, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Log, 0)
	for rows.Next() {
		var l Log
		if err := rows.Scan(
			&l.ID, &l.ActorType, &l.ActorID, &l.Action, &l.Target, &l.Metadata, &l.IP, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
