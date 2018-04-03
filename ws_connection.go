package httpexpect

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// DefaultWsConnectionTimeout is a default timeout duration
// for WebSocket connection read and writes.
//
// You may preconfigure this option globally.
var DefaultWsConnectionTimeout = 30 * time.Second

const noDuration = time.Duration(0)

var infiniteTime = time.Time{}

// WsConnection provides methods to read from, write into and close WebSocket
// connection.
type WsConnection struct {
	chain        chain
	conn         *websocket.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	isClosed     bool
}

// NewWsConnection returns a new WsConnection given a reporter used to report
// failures and websocket.Conn to be inspected and handled.
func NewWsConnection(reporter Reporter, conn *websocket.Conn) *WsConnection {
	return makeWsConnection(makeChain(reporter), conn)
}

func makeWsConnection(chain chain, conn *websocket.Conn) *WsConnection {
	return &WsConnection{
		chain:        chain,
		conn:         conn,
		readTimeout:  DefaultWsConnectionTimeout,
		writeTimeout: DefaultWsConnectionTimeout,
	}
}

// Subprotocol returns a new String object that may be used to inspect
// negotiated protocol for the connection.
func (c *WsConnection) Subprotocol() *String {
	s := &String{chain: c.chain}
	if c.conn != nil {
		s.value = c.conn.Subprotocol()
	}
	return s
}

// ReadTimeout set timeout duration for WebSocket connection reads.
//
// If not set then DefaultWsConnectionTimeout will be used.
func (c *WsConnection) ReadTimeout(timeout time.Duration) *WsConnection {
	c.readTimeout = timeout
	return c
}

// NoReadTimeout removes timeout for WebSocket connection reads.
//
// If not used then DefaultWsConnectionTimeout will be used.
func (c *WsConnection) NoReadTimeout() *WsConnection {
	c.readTimeout = noDuration
	return c
}

// WriteTimeout set timeout duration for WebSocket connection writes.
//
// If not set then DefaultWsConnectionTimeout will be used.
func (c *WsConnection) WriteTimeout(timeout time.Duration) *WsConnection {
	c.writeTimeout = timeout
	return c
}

// NoWriteTimeout removes timeout for WebSocket connection writes.
//
// If not used then DefaultWsConnectionTimeout will be used.
func (c *WsConnection) NoWriteTimeout() *WsConnection {
	c.writeTimeout = noDuration
	return c
}

// Raw returns underlying websocket.Conn object.
// This is the value originally passed to NewConnection.
func (c *WsConnection) Raw() *websocket.Conn {
	return c.conn
}

// Expect reads next message from WebSocket connection and
// returns a new WsMessage object to inspect received message.
//
// Example:
//  msg := conn.Expect()
//  msg.JSON().Object().ValueEqual("message", "hi")
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
	case c.isClosed:
		c.chain.fail("\nunexpected read closed WebSocket connection")
		return
	case !c.setReadDeadline():
		return
	}
	var err error
	m.typ, m.content, err = c.conn.ReadMessage()
	if err != nil {
		if cls, ok := err.(*websocket.CloseError); ok {
			m.typ = websocket.CloseMessage
			m.closeCode = cls.Code
			m.content = []byte(cls.Text)
		} else {
			c.chain.fail(
				"\nexpected read WebSocket connection, "+
					"but got failure: %s", err.Error())
			return
		}
	}
	return
}

func (c *WsConnection) setReadDeadline() bool {
	deadline := infiniteTime
	if c.readTimeout != noDuration {
		deadline = time.Now().Add(c.readTimeout)
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		c.chain.fail(
			"\nunexpected failure when setting "+
				"read WebSocket connection deadline: %s", err.Error())
		return false
	}
	return true
}

// Disconnect closes the underlying WebSocket connection without sending or
// waiting for a close message.
//
// It's okay to call this function multiple times.
//
// It's recommended to always call this function after connection usage is over
// to ensure that no resource leaks will happen.
//
// Example:
//  conn := resp.Connection()
//  defer conn.Disconnect()
func (c *WsConnection) Disconnect() *WsConnection {
	if c.conn == nil || c.isClosed {
		return c
	}
	c.isClosed = true
	c.conn.Close() // nolint: errcheck
	return c
}

// Close cleanly closes the underlying WebSocket connection
// by sending an empty close message and then waiting (with timeout)
// for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  conn := resp.Connection()
//  conn.Close(websocket.CloseUnsupportedData)
func (c *WsConnection) Close(code ...int) *WsConnection {
	switch {
	case c.isClosed || c.checkUnusable("Close"):
		return c
	case len(code) > 1:
		c.chain.fail("\nunexpected multiple code arguments passed to Close")
		return c
	}
	return c.CloseWithBytes(nil, code...)
}

// CloseWithBytes cleanly closes the underlying WebSocket connection
// by sending given slice of bytes as a close message and then waiting
// (with timeout) for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  conn := resp.Connection()
//  conn.CloseWithBytes([]byte("bye!"), websocket.CloseGoingAway)
func (c *WsConnection) CloseWithBytes(b []byte, code ...int) *WsConnection {
	switch {
	case c.isClosed || c.checkUnusable("CloseWithBytes"):
		return c
	case len(code) > 1:
		c.chain.fail(
			"\nunexpected multiple code arguments passed to CloseWithBytes")
		return c
	}

	defer c.Disconnect()

	c.WriteMessage(websocket.CloseMessage, b, code...)

	// Waiting (with timeout) for the server to close the connection.
	c.conn.SetReadDeadline(time.Now().Add(time.Second))
	c.conn.ReadMessage() // nolint: errcheck

	return c
}

// CloseWithJSON cleanly closes the underlying WebSocket connection
// by sending given object (marshaled using json.Marshal()) as a close message
// and then waiting (with timeout) for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  type MyJSON struct {
//    Foo int `json:"foo"`
//  }
//
//  conn := resp.Connection()
//  conn.CloseWithJSON(MyJSON{Foo: 123}, websocket.CloseUnsupportedData)
func (c *WsConnection) CloseWithJSON(
	object interface{}, code ...int,
) *WsConnection {
	switch {
	case c.isClosed || c.checkUnusable("CloseWithJSON"):
		return c
	case len(code) > 1:
		c.chain.fail(
			"\nunexpected multiple code arguments passed to CloseWithJSON")
		return c
	}

	defer c.Disconnect()

	b, err := json.Marshal(object)
	if err != nil {
		c.chain.fail(err.Error())
		return c
	}
	return c.CloseWithBytes(b, code...)
}

// CloseWithText cleanly closes the underlying WebSocket connection
// by sending given text as a close message and then waiting (with timeout)
// for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  conn := resp.Connection()
//  conn.CloseWithText("bye!")
func (c *WsConnection) CloseWithText(s string, code ...int) *WsConnection {
	switch {
	case c.isClosed || c.checkUnusable("CloseWithText"):
		return c
	case len(code) > 1:
		c.chain.fail(
			"\nunexpected multiple code arguments passed to CloseWithText")
		return c
	}
	return c.CloseWithBytes([]byte(s), code...)
}

// WriteMessage writes to the underlying WebSocket connection a message
// of given type with given content.
// Additionally, WebSocket close code may be specified for close messages.
//
// WebSocket message types are defined in RFC 6455, section 11.8.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//  conn := resp.Connection()
//  conn.WriteMessage(websocket.CloseMessage, []byte("Namárië..."))
func (c *WsConnection) WriteMessage(
	typ int, content []byte, closeCode ...int,
) *WsConnection {
	if c.checkUnusable("WriteMessage") {
		return c
	}

	switch typ {
	case websocket.TextMessage, websocket.BinaryMessage:
	case websocket.CloseMessage:
		if len(closeCode) > 1 {
			c.chain.fail("\nunexpected multiple closeCode arguments " +
				"passed to WriteMessage")
			return c
		}
		code := websocket.CloseNormalClosure
		if len(closeCode) > 0 {
			code = closeCode[0]
		}
		content = websocket.FormatCloseMessage(code, string(content))
	default:
		c.chain.fail("\nunexpected WebSocket message type '%s' "+
			"passed to WriteMessage", wsMessageTypeName(typ))
		return c
	}

	if !c.setWriteDeadline() {
		return c
	}
	if err := c.conn.WriteMessage(typ, content); err != nil {
		c.chain.fail(
			"\nexpected write into WebSocket connection, "+
				"but got failure: %s", err.Error())
	}

	return c
}

// WriteBytes is a shorthand for c.WriteMessage(websocket.TextMessage, b).
func (c *WsConnection) WriteBytes(b []byte) *WsConnection {
	if c.checkUnusable("WriteBytes") {
		return c
	}
	return c.WriteMessage(websocket.TextMessage, b)
}

// WriteBytes is a shorthand for c.WriteMessage(websocket.BinaryMessage, b).
func (c *WsConnection) WriteBinary(b []byte) *WsConnection {
	if c.checkUnusable("WriteBinary") {
		return c
	}
	return c.WriteMessage(websocket.BinaryMessage, b)
}

// WriteJSON writes to the underlying WebSocket connection given object,
// marshaled using json.Marshal().
func (c *WsConnection) WriteJSON(object interface{}) *WsConnection {
	if c.checkUnusable("WriteJSON") {
		return c
	}

	b, err := json.Marshal(object)
	if err != nil {
		c.chain.fail(err.Error())
		return c
	}

	return c.WriteMessage(websocket.TextMessage, b)
}

// WriteText is a shorthand for
// c.WriteMessage(websocket.TextMessage, []byte(s)).
func (c *WsConnection) WriteText(s string) *WsConnection {
	if c.checkUnusable("WriteText") {
		return c
	}
	return c.WriteMessage(websocket.TextMessage, []byte(s))
}

func (c *WsConnection) checkUnusable(where string) bool {
	switch {
	case c.chain.failed():
		return true
	case c.conn == nil:
		c.chain.fail("\nunexpected %s usage for failed WebSocket connection",
			where)
		return true
	case c.isClosed:
		c.chain.fail("\nunexpected %s usage for closed WebSocket connection",
			where)
		return true
	}
	return false
}

func (c *WsConnection) setWriteDeadline() bool {
	deadline := infiniteTime
	if c.writeTimeout != noDuration {
		deadline = time.Now().Add(c.writeTimeout)
	}
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		c.chain.fail(
			"\nunexpected failure when setting "+
				"write WebSocket connection deadline: %s", err.Error())
		return false
	}
	return true
}
