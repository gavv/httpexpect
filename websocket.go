package httpexpect

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const noDuration = time.Duration(0)

var infiniteTime = time.Time{}

// Websocket provides methods to read from, write into and close WebSocket
// connection.
type Websocket struct {
	config Config
	chain  *chain

	conn WebsocketConn

	readTimeout  time.Duration
	writeTimeout time.Duration

	isClosed bool
}

// WebsocketConn is used by Websocket to communicate with actual WebSocket connection.
type WebsocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	Subprotocol() string
}

// NewWebsocket returns a new Websocket instance.
func NewWebsocket(config Config, conn WebsocketConn) *Websocket {
	config.fillDefaults()

	return newWebsocket(
		newChainWithConfig("Websocket()", config),
		config,
		conn,
	)
}

func newWebsocket(parent *chain, config Config, conn WebsocketConn) *Websocket {
	chain := parent.clone()

	return &Websocket{
		config: config,
		chain:  chain,
		conn:   conn,
	}
}

// Conn returns underlying WebsocketConn object.
// This is the value originally passed to NewConnection.
func (c *Websocket) Conn() WebsocketConn {
	return c.conn
}

// Deprecated: use Conn instead.
func (c *Websocket) Raw() *websocket.Conn {
	if c.conn == nil {
		return nil
	}
	conn, ok := c.conn.(*websocket.Conn)
	if !ok {
		return nil
	}
	return conn
}

// WithReadTimeout sets timeout duration for WebSocket connection reads.
//
// By default no timeout is used.
func (c *Websocket) WithReadTimeout(timeout time.Duration) *Websocket {
	c.chain.enter("WithReadTimeout()")
	defer c.chain.leave()

	if c.chain.failed() {
		return c
	}

	c.readTimeout = timeout

	return c
}

// WithoutReadTimeout removes timeout for WebSocket connection reads.
func (c *Websocket) WithoutReadTimeout() *Websocket {
	c.chain.enter("WithoutReadTimeout()")
	defer c.chain.leave()

	if c.chain.failed() {
		return c
	}

	c.readTimeout = noDuration

	return c
}

// WithWriteTimeout sets timeout duration for WebSocket connection writes.
//
// By default no timeout is used.
func (c *Websocket) WithWriteTimeout(timeout time.Duration) *Websocket {
	c.chain.enter("WithWriteTimeout()")
	defer c.chain.leave()

	if c.chain.failed() {
		return c
	}

	c.writeTimeout = timeout

	return c
}

// WithoutWriteTimeout removes timeout for WebSocket connection writes.
//
// If not used then DefaultWebsocketTimeout will be used.
func (c *Websocket) WithoutWriteTimeout() *Websocket {
	c.chain.enter("WithoutWriteTimeout()")
	defer c.chain.leave()

	if c.chain.failed() {
		return c
	}

	c.writeTimeout = noDuration

	return c
}

// Subprotocol returns a new String instance with negotiated protocol
// for the connection.
func (c *Websocket) Subprotocol() *String {
	c.chain.enter("Subprotocol()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newString(c.chain, "")
	}

	if c.conn == nil {
		return newString(c.chain, "")
	}

	return newString(c.chain, c.conn.Subprotocol())
}

// Expect reads next message from WebSocket connection and
// returns a new WebsocketMessage instance.
//
// Example:
//
//	msg := conn.Expect()
//	msg.JSON().Object().ValueEqual("message", "hi")
func (c *Websocket) Expect() *WebsocketMessage {
	c.chain.enter("Expect()")
	defer c.chain.leave()

	if c.checkUnusable("Expect()") {
		return newWebsocketMessage(c.chain)
	}

	m := c.readMessage()
	if m == nil {
		return newWebsocketMessage(c.chain)
	}

	return m
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
//
//	conn := resp.Connection()
//	defer conn.Disconnect()
func (c *Websocket) Disconnect() *Websocket {
	c.chain.enter("Disconnect()")
	defer c.chain.leave()

	if c.conn == nil || c.isClosed {
		return c
	}

	c.isClosed = true

	if err := c.conn.Close(); err != nil {
		c.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("got close error when disconnecting websocket"),
				err,
			},
		})
	}

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
//
//	conn := resp.Connection()
//	conn.Close(websocket.CloseUnsupportedData)
func (c *Websocket) Close(code ...int) *Websocket {
	c.chain.enter("Close()")
	defer c.chain.leave()

	switch {
	case c.checkUnusable("Close()"):
		return c

	case len(code) > 1:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple code arguments"),
			},
		})
		return c
	}

	c.writeMessage(websocket.CloseMessage, nil, code...)

	return c
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
//
//	conn := resp.Connection()
//	conn.CloseWithBytes([]byte("bye!"), websocket.CloseGoingAway)
func (c *Websocket) CloseWithBytes(b []byte, code ...int) *Websocket {
	c.chain.enter("CloseWithBytes()")
	defer c.chain.leave()

	switch {
	case c.checkUnusable("CloseWithBytes()"):
		return c

	case len(code) > 1:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple code arguments"),
			},
		})
		return c
	}

	c.writeMessage(websocket.CloseMessage, b, code...)

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
//
//	type MyJSON struct {
//	  Foo int `json:"foo"`
//	}
//
//	conn := resp.Connection()
//	conn.CloseWithJSON(MyJSON{Foo: 123}, websocket.CloseUnsupportedData)
func (c *Websocket) CloseWithJSON(
	object interface{}, code ...int,
) *Websocket {
	c.chain.enter("CloseWithJSON()")
	defer c.chain.leave()

	switch {
	case c.checkUnusable("CloseWithJSON()"):
		return c

	case len(code) > 1:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple code arguments"),
			},
		})
		return c
	}

	b, err := json.Marshal(object)

	if err != nil {
		c.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{object},
			Errors: []error{
				errors.New("invalid json object"),
				err,
			},
		})
		return c
	}

	c.writeMessage(websocket.CloseMessage, b, code...)

	return c
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
//
//	conn := resp.Connection()
//	conn.CloseWithText("bye!")
func (c *Websocket) CloseWithText(s string, code ...int) *Websocket {
	c.chain.enter("CloseWithText()")
	defer c.chain.leave()

	switch {
	case c.checkUnusable("CloseWithText()"):
		return c

	case len(code) > 1:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple code arguments"),
			},
		})
		return c
	}

	c.writeMessage(websocket.CloseMessage, []byte(s), code...)

	return c
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
//
//	conn := resp.Connection()
//	conn.WriteMessage(websocket.CloseMessage, []byte("Namárië..."))
func (c *Websocket) WriteMessage(typ int, content []byte, closeCode ...int) *Websocket {
	c.chain.enter("WriteMessage()")
	defer c.chain.leave()

	if c.checkUnusable("WriteMessage()") {
		return c
	}

	c.writeMessage(typ, content, closeCode...)

	return c
}

// WriteBytesBinary is a shorthand for c.WriteMessage(websocket.BinaryMessage, b).
func (c *Websocket) WriteBytesBinary(b []byte) *Websocket {
	c.chain.enter("WriteBytesBinary()")
	defer c.chain.leave()

	if c.checkUnusable("WriteBytesBinary()") {
		return c
	}

	c.writeMessage(websocket.BinaryMessage, b)

	return c
}

// WriteBytesText is a shorthand for c.WriteMessage(websocket.TextMessage, b).
func (c *Websocket) WriteBytesText(b []byte) *Websocket {
	c.chain.enter("WriteBytesText()")
	defer c.chain.leave()

	if c.checkUnusable("WriteBytesText()") {
		return c
	}

	c.writeMessage(websocket.TextMessage, b)

	return c
}

// WriteText is a shorthand for
// c.WriteMessage(websocket.TextMessage, []byte(s)).
func (c *Websocket) WriteText(s string) *Websocket {
	c.chain.enter("WriteText()")
	defer c.chain.leave()

	if c.checkUnusable("WriteText()") {
		return c
	}

	return c.WriteMessage(websocket.TextMessage, []byte(s))
}

// WriteJSON writes to the underlying WebSocket connection given object,
// marshaled using json.Marshal().
func (c *Websocket) WriteJSON(object interface{}) *Websocket {
	c.chain.enter("WriteJSON()")
	defer c.chain.leave()

	if c.checkUnusable("WriteJSON()") {
		return c
	}

	b, err := json.Marshal(object)

	if err != nil {
		c.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{object},
			Errors: []error{
				errors.New("invalid json object"),
				err,
			},
		})
		return c
	}

	c.writeMessage(websocket.TextMessage, b)

	return c
}

func (c *Websocket) checkUnusable(where string) bool {
	switch {
	case c.chain.failed():
		return true

	case c.conn == nil:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf("unexpected %s call for failed websocket connection", where),
			},
		})
		return true

	case c.isClosed:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf("unexpected %s call for closed websocket connection", where),
			},
		})
		return true
	}

	return false
}

func (c *Websocket) readMessage() *WebsocketMessage {
	m := newWebsocketMessage(c.chain)

	if !c.setReadDeadline() {
		return nil
	}

	var err error
	m.typ, m.content, err = c.conn.ReadMessage()

	if err != nil {
		closeErr, ok := err.(*websocket.CloseError)
		if !ok {
			c.chain.fail(AssertionFailure{
				Type: AssertOperation,
				Errors: []error{
					errors.New("failed to read from websocket"),
					err,
				},
			})
			return nil
		}

		m.typ = websocket.CloseMessage
		m.closeCode = closeErr.Code
		m.content = []byte(closeErr.Text)
	}

	c.printRead(m.typ, m.content, m.closeCode)

	return m
}

func (c *Websocket) writeMessage(typ int, content []byte, closeCode ...int) {
	switch typ {
	case websocket.TextMessage, websocket.BinaryMessage:
		c.printWrite(typ, content, 0)

	case websocket.CloseMessage:
		if len(closeCode) > 1 {
			c.chain.fail(AssertionFailure{
				Type: AssertUsage,
				Errors: []error{
					errors.New("unexpected multiple closeCode arguments"),
				},
			})
			return
		}

		code := websocket.CloseNormalClosure
		if len(closeCode) > 0 {
			code = closeCode[0]
		}

		c.printWrite(typ, content, code)

		content = websocket.FormatCloseMessage(code, string(content))

	default:
		c.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf("unexpected websocket message type %s",
					wsMessageType(typ)),
			},
		})
		return
	}

	if !c.setWriteDeadline() {
		return
	}

	if err := c.conn.WriteMessage(typ, content); err != nil {
		c.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to write to websocket"),
				err,
			},
		})
		return
	}
}

func (c *Websocket) setReadDeadline() bool {
	deadline := infiniteTime
	if c.readTimeout != noDuration {
		deadline = time.Now().Add(c.readTimeout)
	}

	if err := c.conn.SetReadDeadline(deadline); err != nil {
		c.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to set read deadline for websocket"),
				err,
			},
		})
		return false
	}

	return true
}

func (c *Websocket) setWriteDeadline() bool {
	deadline := infiniteTime
	if c.writeTimeout != noDuration {
		deadline = time.Now().Add(c.writeTimeout)
	}

	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		c.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to set write deadline for websocket"),
				err,
			},
		})
		return false
	}

	return true
}

func (c *Websocket) printRead(typ int, content []byte, closeCode int) {
	for _, printer := range c.config.Printers {
		if p, ok := printer.(WebsocketPrinter); ok {
			p.WebsocketRead(typ, content, closeCode)
		}
	}
}

func (c *Websocket) printWrite(typ int, content []byte, closeCode int) {
	for _, printer := range c.config.Printers {
		if p, ok := printer.(WebsocketPrinter); ok {
			p.WebsocketWrite(typ, content, closeCode)
		}
	}
}
