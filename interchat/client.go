package interchat

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type clientOptions struct {
	writeWait  time.Duration
	pongWait   time.Duration
	pingPeriod time.Duration
	readLimit  int64
}

var defaultClientOptions = &clientOptions{
	writeWait:  10 * time.Second,
	pongWait:   10 * time.Second,
	pingPeriod: 5 * time.Second,
	readLimit:  4096,
}

// client represents a single chatting user.
type client struct {
	// socket is the web socket for this client.
	conn *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the room this client is chatting in.
	room *room
	// options is client options.
	options *clientOptions
	// stop is stop channel which is called when client is removed
	// from the room.
	stop chan struct{}
}

// newClient creates new client with default client options.
func newClient(conn *websocket.Conn, r *room) *client {
	return newClientWithOptions(conn, r, defaultClientOptions)
}

func newClientWithOptions(conn *websocket.Conn, r *room, options *clientOptions) *client {
	c := &client{
		conn:    conn,
		send:    make(chan []byte, messageBufferSize),
		room:    r,
		options: options,
		stop:    make(chan struct{}),
	}
	return c
}

// read starts reading messages comming from websocket connection and
// forwarding them to chat room.
func (c *client) read() {
	defer c.conn.Close()
	c.conn.SetReadLimit(c.options.readLimit)
	c.conn.SetReadDeadline(time.Now().Add(c.options.pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.options.pongWait))
		return nil
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Error(err)
			}
			break
		}
		// Send the message to forward channel which writes the message
		// to the stream and fans out the message to all connected clients.
		c.room.forward <- msg
	}
}

// write starts writing messages comming from the chat room to client websocket
// connection. Ticker makes sure ping message is sent once in a while
// to keep connection alive.
func (c *client) write() {
	ticker := time.NewTicker(c.options.pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.options.writeWait))

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.options.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
