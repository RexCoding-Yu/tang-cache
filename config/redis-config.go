package config

import (
	"github.com/go-redis/redis/v8"
)

type RedisConfigMode int

type RedisConfig struct {
	Options *redis.Options
	Client  *redis.Client
}
