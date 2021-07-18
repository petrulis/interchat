package interchat

import (
	"os"
)

// Config provides Interchat configuration options.
type Config struct {
	// HttpAddress is http server address.
	HttpAddress string
	// RedisAddr is address of the redis cluster.
	RedisAddr string
}

// FromEnv creates new config from env variables.
func FromEnv() (*Interchat, error) {
	config := &Config{
		HttpAddress: os.Getenv("WS_HTTP_ADDRESS"),
		RedisAddr:   os.Getenv("WS_REDIS_ADDR"),
	}
	return New(config)
}
