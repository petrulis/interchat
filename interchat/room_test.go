package interchat

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/petrulis/interchat/pkg/stream"
	"github.com/stretchr/testify/assert"
)

type mockStream struct {
	revn    func(ch chan<- []byte, n int64) error
	publish func(msg []byte) error
}

func (s *mockStream) RevN(ch chan<- []byte, n int64) error {
	return s.revn(ch, n)
}

func (s *mockStream) Publish(msg []byte) error {
	return s.publish(msg)
}

func TestRoom(t *testing.T) {
	newStream = func(name string, options *stream.StreamOptions) stream.Stream {
		return &mockStream{
			publish: func(msg []byte) error {
				return nil
			},
			revn: func(ch chan<- []byte, n int64) error {
				return nil
			},
		}
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
		for i := range clients {
			r.join <- clients[i]
		}

		wg := sync.WaitGroup{}
		wg.Add(numOfClients)

		ch := make(chan []byte, 1000000)
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

	t.Run("check message never gets sent on error", func(tt *testing.T) {
		newStream = func(name string, options *stream.StreamOptions) stream.Stream {
			return &mockStream{
				publish: func(msg []byte) error {
					return fmt.Errorf("error")
				},
				revn: func(ch chan<- []byte, n int64) error {
					return nil
				},
			}
		}
		r := newRoom("public", nil)
		go r.open()

		client := &client{
			send: make(chan []byte, 1),
		}
		r.join <- client
		r.forward <- []byte("message")

		assert.Equal(t, 0, len(client.send))
	})

	t.Run("check history is received", func(tt *testing.T) {
		numberOfMessages := 100
		done := make(chan struct{})
		newStream = func(name string, options *stream.StreamOptions) stream.Stream {
			return &mockStream{
				revn: func(ch chan<- []byte, n int64) error {
					for i := int64(0); i < n; i++ {
						ch <- []byte("message")
					}
					close(done)
					return nil
				},
			}
		}
		r := newRoom("public", nil)
		go r.open()

		client1 := &client{
			send: make(chan []byte, numberOfMessages),
		}
		r.join <- client1
		<-done
		assert.Equal(t, r.cap, int64(len(client1.send)))
	})
}
