package httpexpect

import (
	"github.com/gorilla/websocket"
	"time"
)

type WsConnection struct {
	chain  chain
	conn   *websocket.Conn
	closed bool
}

func NewWsConnection(reporter Reporter, conn *websocket.Conn) *WsConnection {
	return makeWsConnection(makeChain(reporter), conn)
}

func makeWsConnection(chain chain, conn *websocket.Conn) *WsConnection {
	return &WsConnection{
		chain: chain,
		conn:  conn,
	}
}

func (c *WsConnection) Close() {
	if c.conn == nil || c.closed {
		return
	}

	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return
	}
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	c.conn.ReadMessage() // nolint: errcheck
}

// Raw returns underlying websocket.Conn object.
// This is the value originally passed to NewConnection.
func (c *WsConnection) Raw() *websocket.Conn {
	return c.conn
}

func (c *WsConnection) Expect() (m *WsMessage) {
	m = &WsMessage{
		chain: c.chain,
	}
	switch {
	case c.chain.failed():
		return
	case c.conn == nil:
		c.chain.fail("\nunexpected read failed WebSocket connection")
		return
	case c.closed:
		c.chain.fail("\nunexpected read closed WebSocket connection")
		return
	}
	var err error
	// TODO: deadline???
	// TODO: correct Read ping message?
	m.typ, m.body, err = c.conn.ReadMessage()
	if err != nil {
		if cls, ok := err.(*websocket.CloseError); ok {
			m.closeCode = cls.Code
			m.body = []byte(cls.Text)
		} else {
			c.chain.fail(
				"\nexpected read WebSocket connection, "+
					"but got failure: %s", err.Error())
			return
		}
	}
	return
}
