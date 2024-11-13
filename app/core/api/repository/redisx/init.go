package redisx

import "github.com/redis/go-redis/v9"

func NewRedis(host, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})
	return rdb
}
