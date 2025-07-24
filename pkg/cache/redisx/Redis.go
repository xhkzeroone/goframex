package redisx

import (
	"context"
	"encoding/json"
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
