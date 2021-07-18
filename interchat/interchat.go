package interchat

import (
	"net/http"

	"github.com/go-redis/redis/v8"
)

var newRedis = func(c *Config) redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		PoolSize: 10,
		Addr:     c.RedisAddr,
	})
}

// Interchat represents http server that upgrades client connections
// and adds the clients to the room.
type Interchat struct {
	config *Config
	// room is a single room the chat currently supports.
	// When interchat is started, room is opened for client
	// to join, leave and write messages.
	room *room
}

// New creates new ws server.
func New(config *Config) (*Interchat, error) {
	c := &Interchat{
		config: config,
		room:   newRoom("public", newRedis(config)),
	}
	return c, nil
}

// Run starts and runs the server until error happens.
func (c *Interchat) Run() error {
	http.Handle("/room", c)
	// Since only one room is supported, simply open it when server
	// is started.
	go c.room.open()
	return http.ListenAndServe(c.config.HttpAddress, nil)
}

// ServeHTTP implements http.Handler and works as entrypoint to the room.
func (c *Interchat) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	// Create new client and subscribe it to the room by sending it
	// over join channel.
	client := newClient(conn, c.room)
	c.room.join <- client

	defer func() {
		c.room.leave <- client
	}()

	go client.write()
	// This blocks until the client leaves the room.
	client.read()
}
