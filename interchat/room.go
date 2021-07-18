package interchat

import (
	"net/http"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/petrulis/interchat/pkg/stream"
	"github.com/sirupsen/logrus"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var newStream = stream.NewStreamWithOptions

// room represents interchat room.
type room struct {
	// name is room name.
	name string
	// cap is a maximum number of historical messages
	// to be shown when new client connects.
	cap int64
	// stream is underlying data store, where messages are
	// published and read when needed.
	stream stream.Stream
	// forward is a channel that has incoming
	// messagages, that needs to be send (forwarded) to
	// other clients in the room.
	forward chan []byte
	// join is a channel that is tiggered when new http connection
	// is upgraded to websocket.
	join chan *client
	// leave is a channel that is triggered when a clients
	// is about to leave the room. A room can be left
	// due to closed websocket connection.
	leave chan *client
	// mu locks clients.
	mu *sync.Mutex
	// clients has all active clients in the room. When
	// new client joins or leaves the room, it is added or removed
	// to/from this structure.
	clients map[*client]bool
}

// newRoom creates new chat room.
func newRoom(name string, r redis.UniversalClient) *room {
	return &room{
		name: name,
		cap:  50,
		stream: newStream(name, &stream.StreamOptions{
			RedisClient: r,
		}),
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		mu:      new(sync.Mutex),
		clients: make(map[*client]bool),
	}
}

func (r *room) subscribe(client *client) {
	if err := r.stream.RevN(client.send, r.cap); err == nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.clients[client] = true
	}
}

// unsubscribe removes the client from the room.
func (r *room) unsubscribe(client *client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, client)
	close(client.send)
}

// broadcast publishes a message to the stream first to make sure all messages
// are written before broadcasted to the clients.
func (r *room) broadcast(msg []byte) {
	if err := r.stream.Publish(msg); err != nil {
		logrus.Error(err)
		return
	}
	for client := range r.clients {
		client.send <- msg
	}
}

// open the room for clients to be able to join, leave and write messages until
// shutdown signal is received.
func (r *room) open() {
	for {
		select {
		case client := <-r.join:
			r.subscribe(client)
		case client := <-r.leave:
			r.unsubscribe(client)
		case msg := <-r.forward:
			r.broadcast(msg)
		}
	}
}
