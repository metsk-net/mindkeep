// pkg/database/redis.go
package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &RedisClient{client: client}
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// セッション管理用メソッド（後で使用）
func (r *RedisClient) SetSession(ctx context.Context, sessionID string, userID string, expiration time.Duration) error {
	return r.client.Set(ctx, "session:"+sessionID, userID, expiration).Err()
}

func (r *RedisClient) GetSession(ctx context.Context, sessionID string) (string, error) {
	return r.client.Get(ctx, "session:"+sessionID).Result()
}
