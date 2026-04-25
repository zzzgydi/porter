package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func New(addr, password string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &Client{rdb}
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

func (c *Client) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	pipe := c.Pipeline()
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d:%s", now, uuid.New().String())})
	pipe.Expire(ctx, key, window)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := cmds[1].(*redis.IntCmd).Val()
	return count < int64(limit), nil
}

func (c *Client) MarkWebhookEvent(ctx context.Context, eventID string, ttl time.Duration) (bool, error) {
	key := "webhook:event:" + eventID
	return c.SetNX(ctx, key, "1", ttl).Result()
}
