package examples

import (
	"net/http"

	"github.com/gorilla/websocket"

	fastws "github.com/fasthttp-contrib/websocket"
	"github.com/valyala/fasthttp"
)

// WsHttpHandler is a simple http.Handler that implements WebSocket echo server.
func WsHttpHandler(w http.ResponseWriter, r *http.Request) {
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

// WsFastHandler is a simple fasthttp.RequestHandler that implements
// WebSocket echo server.
func WsFastHandler(ctx *fasthttp.RequestCtx) {
	upgrader := fastws.New(func (c *fastws.Conn) {
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
	})
	err := upgrader.Upgrade(ctx)
	if err != nil {
		panic(err)
	}
}
