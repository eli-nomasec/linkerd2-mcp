// internal/redis/redis.go

package redisutil

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Addr   string
	Client *redis.Client
	// Add fields for connection pool, logger, etc.
}

func NewRedisClient(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{
		Addr:   addr,
		Client: rdb,
	}
}

// SetMeshSnapshot sets the mesh snapshot with expiration (ttl)
func (r *RedisClient) SetMeshSnapshot(ctx context.Context, data []byte, ttl time.Duration) error {
	return r.Client.Set(ctx, "mesh:snapshot", data, ttl).Err()
}

// PublishMeshDelta publishes a mesh delta to the mesh:delta channel
func (r *RedisClient) PublishMeshDelta(ctx context.Context, delta []byte) error {
	return r.Client.Publish(ctx, "mesh:delta", delta).Err()
}

// TryAcquireLeader attempts to acquire leadership using SETNX with expiration
func (r *RedisClient) TryAcquireLeader(ctx context.Context, podUID string, ttl time.Duration) (bool, error) {
	ok, err := r.Client.SetNX(ctx, "mesh:leader", podUID, ttl).Result()
	return ok, err
}

// GetMeshSnapshot retrieves the mesh snapshot from Redis
func (r *RedisClient) GetMeshSnapshot(ctx context.Context) ([]byte, error) {
	val, err := r.Client.Get(ctx, "mesh:snapshot").Bytes()
	if err == redis.Nil {
		return nil, nil // Not found
	}
	return val, err
}

// SubscribeMeshDelta subscribes to the mesh:delta channel and calls handler on each message
func (r *RedisClient) SubscribeMeshDelta(ctx context.Context, handler func([]byte)) error {
	sub := r.Client.Subscribe(ctx, "mesh:delta")
	ch := sub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			handler([]byte(msg.Payload))
		case <-ctx.Done():
			_ = sub.Close()
			return ctx.Err()
		}
	}
}
