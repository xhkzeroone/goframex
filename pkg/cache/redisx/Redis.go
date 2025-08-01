package redisx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*redis.Client
}

func New(cfg *Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to Redis")
	return &Redis{client}, nil
}

func (r *Redis) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, data, expiration).Err()
}

func (r *Redis) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *Redis) HealthCheck(ctx context.Context) error {
	if err := r.Ping(ctx).Err(); err != nil {
		return err
	}
	testKey := "health_check_test"
	testValue := "ok"

	if err := r.Set(ctx, testKey, testValue, time.Second).Err(); err != nil {
		return fmt.Errorf("redis write test failed: %w", err)
	}

	val, err := r.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("redis read test failed: %w", err)
	}

	if val != testValue {
		return fmt.Errorf("redis read/write consistency test failed")
	}
	return nil
}
