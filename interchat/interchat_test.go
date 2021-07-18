package interchat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestExample(t *testing.T) {
	c := Interchat{
		room: newRoom("public", nil),
	}
	s := httptest.NewServer(http.HandlerFunc(c.ServeHTTP))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	errch := make(chan error)
	go func() {
		_, p, err := ws.ReadMessage()
		if err != nil {
			errch <- err
		}
		if string(p) != "hello" {
			errch <- err
		}
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello 2")); err != nil {
			errch <- err
		}
	}()
	go func() {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			errch <- err
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			errch <- err
		}
		if string(p) != "hello 2" {
			errch <- err
		}
	}()
	err = <-errch
	if err != nil {
		t.Fatal(err)
	}
}
