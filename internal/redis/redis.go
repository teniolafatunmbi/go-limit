package redis

import "github.com/redis/go-redis/v9"

func NewRedisClient() *redis.Client {
	opts := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	}
	return redis.NewClient(opts)
}
