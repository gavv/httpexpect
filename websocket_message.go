package httpexpect

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

// WebsocketMessage provides methods to inspect message read from WebSocket connection.
type WebsocketMessage struct {
	chain     chain
	typ       int
	content   []byte
	closeCode int
}

// NewWebsocketMessage returns a new WebsocketMessage object given a reporter used to
// report failures and the message parameters to be inspected.
//
// reporter should not be nil.
//
// Example:
//   m := NewWebsocketMessage(reporter, websocket.TextMessage, []byte("content"), 0)
//   m.TextMessage()
func NewWebsocketMessage(
	reporter Reporter, typ int, content []byte, closeCode ...int,
) *WebsocketMessage {
	m := &WebsocketMessage{
		chain:   makeChain(reporter),
		typ:     typ,
		content: content,
	}
	if len(closeCode) != 0 {
		m.closeCode = closeCode[0]
	}
	return m
}

func makeWebsocketMessage(chain chain) *WebsocketMessage {
	return &WebsocketMessage{
		chain: chain,
	}
}

// Raw returns underlying type, content and close code of WebSocket message.
// Theses values are originally read from WebSocket connection.
func (m *WebsocketMessage) Raw() (typ int, content []byte, closeCode int) {
	return m.typ, m.content, m.closeCode
}

// CloseMessage is a shorthand for m.Type(websocket.CloseMessage).
func (m *WebsocketMessage) CloseMessage() *WebsocketMessage {
	return m.Type(websocket.CloseMessage)
}

// NotCloseMessage is a shorthand for m.NotType(websocket.CloseMessage).
func (m *WebsocketMessage) NotCloseMessage() *WebsocketMessage {
	return m.NotType(websocket.CloseMessage)
}

// BinaryMessage is a shorthand for m.Type(websocket.BinaryMessage).
func (m *WebsocketMessage) BinaryMessage() *WebsocketMessage {
	return m.Type(websocket.BinaryMessage)
}

// NotBinaryMessage is a shorthand for m.NotType(websocket.BinaryMessage).
func (m *WebsocketMessage) NotBinaryMessage() *WebsocketMessage {
	return m.NotType(websocket.BinaryMessage)
}

// TextMessage is a shorthand for m.Type(websocket.TextMessage).
func (m *WebsocketMessage) TextMessage() *WebsocketMessage {
	return m.Type(websocket.TextMessage)
}

// NotTextMessage is a shorthand for m.NotType(websocket.TextMessage).
func (m *WebsocketMessage) NotTextMessage() *WebsocketMessage {
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
func (m *WebsocketMessage) Type(typ ...int) *WebsocketMessage {
	switch {
	case m.chain.failed():
		return m
	case len(typ) == 0:
		m.chain.fail(newErrorFailure(fmt.Errorf("unexpected nil argument passed to Type")))
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
		failure := Failure{
			AssertionName: "websocket.message",
			AssertType:    FailureAssertEqual,
			Actual:        m.typ,
			Expected:      typ,
		}
		if len(typ) == 1 {
			failure.Expected = typ[0]
		}
		m.chain.fail(failure)
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
func (m *WebsocketMessage) NotType(typ ...int) *WebsocketMessage {
	switch {
	case m.chain.failed():
		return m
	case len(typ) == 0:
		m.chain.fail(newErrorFailure(fmt.Errorf("unexpected nil argument passed to NotType")))
		return m
	}
	for _, t := range typ {
		if t == m.typ {
			failure := Failure{
				AssertionName: "websocket.message",
				AssertType:    FailureAssertNotEqual,
				Expected:      typ,
				Actual:        m.typ,
			}
			if len(typ) == 1 {
				failure.Expected = typ[0]
			}

			m.chain.fail(failure)

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
func (m *WebsocketMessage) Code(code ...int) *WebsocketMessage {
	switch {
	case m.chain.failed():
		return m
	case len(code) == 0:
		m.chain.fail(newErrorFailure(fmt.Errorf("unexpected nil argument passed to Code")))
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
		failure := Failure{
			AssertionName: "websocket close code",
			AssertType:    FailureAssertNotEqual,
			Expected:      code,
			Actual:        m.closeCode,
		}
		if len(code) == 1 {
			failure.Expected = code[0]
		}
		m.chain.fail(failure)
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
func (m *WebsocketMessage) NotCode(code ...int) *WebsocketMessage {
	switch {
	case m.chain.failed():
		return m
	case len(code) == 0:
		m.chain.fail(newErrorFailure(fmt.Errorf("unexpected nil argument passed to CodeNotEqual")))
		return m
	case m.checkClosed("NotCode"):
		return m
	}
	for _, c := range code {
		if c == m.closeCode {
			failure := Failure{
				AssertionName: "websocket close code",
				AssertType:    FailureAssertNotEqual,
				Expected:      code,
				Actual:        m.closeCode,
			}
			if len(code) == 1 {
				failure.Expected = code[0]
			}

			m.chain.fail(failure)

			return m
		}
	}
	return m
}

func (m *WebsocketMessage) checkClosed(where string) bool {
	if m.typ != websocket.CloseMessage {
		failure := Failure{
			AssertionName: "websocket check closed",
			OriginalError: fmt.Errorf("where: %s", where),
			AssertType:    FailureAssertEqual,
			Expected:      websocket.CloseMessage,
			Actual:        m.typ,
		}
		m.chain.fail(failure)
		return true
	}
	return false
}

// Body returns a new String object that may be used to inspect
// WebSocket message content.
//
// Example:
//  msg := conn.Expect()
//  msg.Body().NotEmpty()
//  msg.Body().Length().Equal(100)
func (m *WebsocketMessage) Body() *String {
	return &String{m.chain, string(m.content)}
}

// NoContent succeeds if WebSocket message has no content (is empty).
func (m *WebsocketMessage) NoContent() *WebsocketMessage {
	switch {
	case m.chain.failed():
		return m
	case len(m.content) == 0:
		return m
	}

	failure := Failure{
		AssertionName: "websocket message body being empty",
		AssertType:    FailureAssertNotEmpty,
	}

	switch m.typ {
	case websocket.BinaryMessage:
		failure.Actual = len(m.content)
	default:
		failure.Actual = string(m.content)
	}

	m.chain.fail(failure)

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
func (m *WebsocketMessage) JSON() *Value {
	return &Value{m.chain, m.getJSON()}
}

func (m *WebsocketMessage) getJSON() interface{} {
	if m.chain.failed() {
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(m.content, &value); err != nil {
		m.chain.fail(newErrorFailure(err))
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
