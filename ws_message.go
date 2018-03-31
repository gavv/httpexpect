package httpexpect

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// WsMessage provides methods to inspect message read from WebSocket connection.
type WsMessage struct {
	chain     chain
	typ       int
	content   []byte
	closeCode int
}

// Closed is a shorthand for m.Type(websocket.CloseMessage).
func (m *WsMessage) Closed() *WsMessage {
	return m.Type(websocket.CloseMessage)
}

// NotClosed is a shorthand for m.NotType(websocket.CloseMessage).
func (m *WsMessage) NotClosed() *WsMessage {
	return m.NotType(websocket.CloseMessage)
}

// Binary is a shorthand for m.Type(websocket.BinaryMessage).
func (m *WsMessage) Binary() *WsMessage {
	return m.Type(websocket.BinaryMessage)
}

// NotBinary is a shorthand for m.NotType(websocket.BinaryMessage).
func (m *WsMessage) NotBinary() *WsMessage {
	return m.NotType(websocket.BinaryMessage)
}

// Text is a shorthand for m.Type(websocket.TextMessage).
func (m *WsMessage) Text() *WsMessage {
	return m.Type(websocket.TextMessage)
}

// NotText is a shorthand for m.NotType(websocket.TextMessage).
func (m *WsMessage) NotText() *WsMessage {
	return m.NotType(websocket.TextMessage)
}

// Type succeeds if WebSocket message type is one of the given.
//
// WebSocket message types are defined in RFC 6455, section 11.8.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//  msg := conn.Expect()
//  msg.Type(websocket.TextMessage, websocket.BinaryMessage)
func (m *WsMessage) Type(typ ...int) *WsMessage {
	switch {
	case m.chain.failed():
		return m
	case len(typ) == 0:
		m.chain.fail("\nunexpected nil argument passed to Type")
		return m
	}
	yes := false
	for _, t := range typ {
		if t == m.typ {
			yes = true
			break
		}
	}
	if !yes {
		if len(typ) > 1 {
			m.chain.fail(
				"\nexpected message type equal to one of:\n %v\n\nbut got:\n %d",
				typ, m.typ)
		} else {
			m.chain.fail(
				"\nexpected message type:\n %d\n\nbut got:\n %d",
				typ[0], m.typ)
		}
	}
	return m
}

// NotType succeeds if WebSocket message type is none of the given.
//
// WebSocket message types are defined in RFC 6455, section 11.8.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//  msg := conn.Expect()
//  msg.NotType(websocket.CloseMessage, websocket.BinaryMessage)
func (m *WsMessage) NotType(typ ...int) *WsMessage {
	switch {
	case m.chain.failed():
		return m
	case len(typ) == 0:
		m.chain.fail("\nunexpected nil argument passed to NotType")
		return m
	}
	for _, t := range typ {
		if t == m.typ {
			if len(typ) > 1 {
				m.chain.fail(
					"\nexpected message type not equal:\n %v\n\nbut got:\n %d",
					typ, m.typ)
			} else {
				m.chain.fail(
					"\nexpected message type not equal:\n %d\n\nbut it did",
					typ[0], m.typ)
			}
			return m
		}
	}
	return m
}

// Code succeeds if WebSocket close code is one of the given.
//
// Code fails if WebSocket message type is not "8 - Connection Close Frame".
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//  msg := conn.Expect().Closed()
//  msg.Code(websocket.CloseNormalClosure, websocket.CloseGoingAway)
func (m *WsMessage) Code(code ...int) *WsMessage {
	switch {
	case m.chain.failed():
		return m
	case len(code) == 0:
		m.chain.fail("\nunexpected nil argument passed to Code")
		return m
	case m.checkClosed("Code"):
		return m
	}
	yes := false
	for _, c := range code {
		if c == m.closeCode {
			yes = true
			break
		}
	}
	if !yes {
		if len(code) > 1 {
			m.chain.fail(
				"\nexpected close code equal to one of:\n %v\n\nbut got:\n %d",
				code, m.closeCode)
		} else {
			m.chain.fail(
				"\nexpected close code:\n %d\n\nbut got:\n %d",
				code[0], m.closeCode)
		}
	}
	return m
}

// NotCode succeeds if WebSocket close code is none of the given.
//
// NotCode fails if WebSocket message type is not "8 - Connection Close Frame".
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//  msg := conn.Expect().Closed()
//  msg.NotCode(websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived)
func (m *WsMessage) NotCode(code ...int) *WsMessage {
	switch {
	case m.chain.failed():
		return m
	case len(code) == 0:
		m.chain.fail("\nunexpected nil argument passed to CodeNotEqual")
		return m
	case m.checkClosed("CodeNotEqual"):
		return m
	}
	for _, c := range code {
		if c == m.closeCode {
			if len(code) > 1 {
				m.chain.fail(
					"\nexpected close code not equal:\n %v\n\nbut got:\n %d",
					code, m.closeCode)
			} else {
				m.chain.fail(
					"\nexpected close code not equal:\n %d\n\nbut it did",
					code[0], m.closeCode)
			}
			return m
		}
	}
	return m
}

func (m *WsMessage) checkClosed(where string) bool {
	if m.typ != websocket.CloseMessage {
		m.chain.fail(
			"\nunexpected %s usage for not '%' WebSocket message type\n\n"+
				"got type:\n %s",
			where,
			wsMessageTypeName(websocket.CloseMessage),
			wsMessageTypeName(m.typ))
		return true
	}
	return false
}

// Raw returns underlying type, content and close code of WebSocket message.
// Theses values are originally read from WebSocket connection.
func (m *WsMessage) Raw() (typ int, content []byte, closeCode int) {
	return m.typ, m.content, m.closeCode
}

// Body returns a new String object that may be used to inspect
// WebSocket message content.
//
// Example:
//  msg := conn.Expect()
//  msg.Body().NotEmpty()
//  msg.Body().Length().Equal(100)
func (m *WsMessage) Body() *String {
	return &String{m.chain, string(m.content)}
}

// Content returns a new Value object that may be used to inspect
// WebSocket message content.
//
// Example:
//  msg := conn.Expect()
//  msg.Content().Equal([]byte{0, 1, 2})
func (m *WsMessage) Content() *Value {
	return &Value{m.chain, m.content}
}

// NoContent succeeds if WebSocket message has no content (is empty).
func (m *WsMessage) NoContent() *WsMessage {
	switch {
	case m.chain.failed():
		return m
	case len(m.content) == 0:
		return m
	}
	switch m.typ {
	case websocket.BinaryMessage:
		m.chain.fail(
			"\nexpected message body being empty, but got:\n %d bytes",
			len(m.content))
	default:
		m.chain.fail(
			"\nexpected message body being empty, but got:\n %s",
			string(m.content))
	}
	return m
}

// JSON returns a new Value object that may be used to inspect JSON contents
// of WebSocket message.
//
// JSON succeeds if JSON may be decoded from message content.
//
// Example:
//  msg := conn.Expect()
//  msg.JSON().Array().Elements("foo", "bar")
func (m *WsMessage) JSON() *Value {
	return &Value{m.chain, m.getJSON()}
}

func (m *WsMessage) getJSON() interface{} {
	if m.chain.failed() {
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(m.content, &value); err != nil {
		m.chain.fail(err.Error())
		return nil
	}

	return value
}

func wsMessageTypeName(typ int) string {
	switch typ {
	case websocket.TextMessage:
		return "text"
	case websocket.BinaryMessage:
		return "binary"
	case websocket.CloseMessage:
		return "close"
	case websocket.PingMessage:
		return "ping"
	case websocket.PongMessage:
		return "pong"
	}
	return "unknown"
}
