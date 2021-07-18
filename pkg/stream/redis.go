package stream

import (
	"github.com/go-redis/redis/v8"
)

// RedisOptions is an alias to redis.Options so that package users could
// avoid importing go-redis.
type RedisOptions = redis.Options

// newRedisClient creates a new Redis client with the given options. If
// options is empty, it will use default options.
func newRedisClient(options *RedisOptions) redis.UniversalClient {
	if options == nil {
		options = &RedisOptions{}
	}
	return redis.NewClient(options)
}
