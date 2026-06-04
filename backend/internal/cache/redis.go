package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// Init initializes the Redis client.
func Init(redisURL string) error {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	Client = redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return Client.Ping(ctx).Err()
}

// Set stores a value in Redis with JSON marshaling.
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return nil // Graceful degradation if Redis is not configured
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from Redis with JSON unmarshaling.
func Get(ctx context.Context, key string, dest interface{}) error {
	if Client == nil {
		return redis.Nil // Treat as cache miss
	}

	data, err := Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}
