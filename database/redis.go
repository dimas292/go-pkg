package database

import (
	"github.com/dimas292/go-pkg/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		PoolSize:     10,
		MinIdleConns: 5,
	})

	return rdb, nil
}
