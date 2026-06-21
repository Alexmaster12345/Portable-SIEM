package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/portable-siem/siem/pkg/config"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(cfg config.RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return &RedisStore{client: client}, nil
}

func (r *RedisStore) Close() error { return r.client.Close() }

// IncrWindowCounter increments a sliding-window counter for correlation rules.
// key format: "corr:<rule_id>:<group_key>"
func (r *RedisStore) IncrWindowCounter(ctx context.Context, key string, windowSecs int) (int64, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(windowSecs)*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (r *RedisStore) GetWindowCounter(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *RedisStore) ResetWindowCounter(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Cache stores arbitrary JSON values with TTL.
func (r *RedisStore) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string, dst any) error {
	b, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func (r *RedisStore) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Publish sends an event to a Redis channel (pub/sub for real-time dashboard).
func (r *RedisStore) Publish(ctx context.Context, channel string, payload any) error {
	b, _ := json.Marshal(payload)
	return r.client.Publish(ctx, channel, b).Err()
}

// TrackLastSeen stores when a host was last seen.
func (r *RedisStore) TrackHostLastSeen(ctx context.Context, host string) error {
	return r.client.HSet(ctx, "hosts:last_seen", host, time.Now().Format(time.RFC3339)).Err()
}
