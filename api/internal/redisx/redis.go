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

// rateLimitScript atomically cleans expired entries, checks the limit, and adds a new entry.
// KEYS[1] = key
// ARGV[1] = windowStart (score to remove up to)
// ARGV[2] = now (score for new member)
// ARGV[3] = member
// ARGV[4] = limit
// ARGV[5] = ttl in seconds
var rateLimitScript = redis.NewScript(`
	redis.call("ZREMRANGEBYSCORE", KEYS[1], "0", ARGV[1])
	local count = redis.call("ZCARD", KEYS[1])
	if tonumber(count) >= tonumber(ARGV[4]) then
		return 0
	end
	redis.call("ZADD", KEYS[1], ARGV[2], ARGV[3])
	redis.call("EXPIRE", KEYS[1], ARGV[5])
	return 1
`)

func (c *Client) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())
	member := fmt.Sprintf("%d:%s", now, uuid.New().String())

	result, err := rateLimitScript.Run(ctx, c.Client, []string{key},
		windowStart, now, member, limit, int(window.Seconds()),
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (c *Client) MarkWebhookEvent(ctx context.Context, eventID string, ttl time.Duration) (bool, error) {
	key := "webhook:event:" + eventID
	return c.SetNX(ctx, key, "1", ttl).Result()
}

// UnmarkWebhookEvent removes the dedup marker, e.g. when processing failed
// and the event should be retried.
func (c *Client) UnmarkWebhookEvent(ctx context.Context, eventID string) {
	key := "webhook:event:" + eventID
	_ = c.Del(ctx, key).Err()
}
