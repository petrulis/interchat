package interchat

import (
	"sync"
	"testing"
	"time"

	"github.com/petrulis/interchat/pkg/stream"
	"github.com/stretchr/testify/assert"
)

type mockStream struct {
}

func (s *mockStream) RevN(ch chan<- []byte, n int64) error {
	return nil
}
func (s *mockStream) Publish(msg []byte) error {
	return nil
}

func TestRoom(t *testing.T) {
	newStream = func(name string, options *stream.StreamOptions) stream.Stream {
		return &mockStream{}
	}
	t.Run("joins multiple clients", func(tt *testing.T) {
		numOfClients := 10000
		r := newRoom("public", nil)
		go r.open()

		wg := sync.WaitGroup{}

		clients := make([]*client, numOfClients)
		wg.Add(numOfClients)
		for i := 0; i < numOfClients; i++ {
			c := newClient(nil, r)
			clients[i] = c
			go func(c *client) {
				defer wg.Done()
				r.join <- c
			}(c)
		}
		wg.Wait()
		time.Sleep(10 * time.Millisecond)

		assert.Equal(tt, numOfClients, len(r.clients))

		wg = sync.WaitGroup{}
		wg.Add(numOfClients)
		for i := 0; i < numOfClients; i++ {
			go func(i int) {
				defer wg.Done()
				r.leave <- clients[i]
			}(i)
		}
		wg.Wait()
		time.Sleep(10 * time.Millisecond)

		assert.Equal(tt, 0, len(r.clients))
	})

	t.Run("check all clients receives a message", func(tt *testing.T) {
		r := newRoom("public", nil)
		go r.open()

		numOfClients := 10000
		clients := make([]*client, numOfClients)
		for i := 0; i < numOfClients; i++ {
			clients[i] = &client{
				send: make(chan []byte),
			}
		}
		for _, client := range clients {
			r.join <- client
		}

		wg := sync.WaitGroup{}
		wg.Add(numOfClients)

		ch := make(chan []byte, numOfClients)
		after := time.After(1 * time.Second)

		read := func(c *client) {
			defer wg.Done()
			select {
			case <-after:
				return
			case ch <- <-c.send:
			}
		}
		for _, client := range clients {
			go read(client)
		}
		r.forward <- []byte("message")
		wg.Wait()
		assert.Equal(t, numOfClients, len(ch))
	})
}
