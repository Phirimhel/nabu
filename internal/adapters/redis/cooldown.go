package redisadapter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cooldown struct{ client *redis.Client }

func NewCooldown(client *redis.Client) *Cooldown { return &Cooldown{client: client} }
func (c *Cooldown) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, time.Duration, error) {
	ok, err := c.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil || ok {
		return ok, 0, err
	}
	retryAfter, err := c.client.TTL(ctx, key).Result()
	return false, retryAfter, err
}
