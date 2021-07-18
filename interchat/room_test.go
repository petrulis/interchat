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

	t.Run("client receives message", func(tt *testing.T) {
		r := newRoom("public", nil)
		go r.open()

		client1 := &client{}
		client2 := &client{}

		r.join <- client1
		r.join <- client2

		r.forward <- []byte("message")
	})
}
