package examples

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WsHandler is a simple http.Handler that implements WebSocket echo server.
func WsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := &websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}
