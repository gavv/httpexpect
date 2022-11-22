package httpexpect

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
)

// WebsocketMessage provides methods to inspect message read from WebSocket connection.
type WebsocketMessage struct {
	chain     *chain
	typ       int
	content   []byte
	closeCode int
}

// NewWebsocketMessage returns a new WebsocketMessage instance.
//
// reporter should not be nil.
//
// Example:
//
//	m := NewWebsocketMessage(reporter, websocket.TextMessage, []byte("content"), 0)
//	m.TextMessage()
func NewWebsocketMessage(
	reporter Reporter, typ int, content []byte, closeCode ...int,
) *WebsocketMessage {
	m := newWebsocketMessage(newChainWithDefaults("WebsocketMessage()", reporter))

	m.typ = typ
	m.content = content

	if len(closeCode) != 0 {
		m.closeCode = closeCode[0]
	}

	return m
}

func newWebsocketMessage(parent *chain) *WebsocketMessage {
	return &WebsocketMessage{
		chain: parent.clone(),
	}
}

// Raw returns underlying type, content and close code of WebSocket message.
// Theses values are originally read from WebSocket connection.
func (m *WebsocketMessage) Raw() (typ int, content []byte, closeCode int) {
	return m.typ, m.content, m.closeCode
}

// CloseMessage is a shorthand for m.Type(websocket.CloseMessage).
func (m *WebsocketMessage) CloseMessage() *WebsocketMessage {
	m.chain.enter("CloseMessage()")
	defer m.chain.leave()

	m.checkType(websocket.CloseMessage)

	return m
}

// NotCloseMessage is a shorthand for m.NotType(websocket.CloseMessage).
func (m *WebsocketMessage) NotCloseMessage() *WebsocketMessage {
	m.chain.enter("NotCloseMessage()")
	defer m.chain.leave()

	m.checkNotType(websocket.CloseMessage)

	return m
}

// BinaryMessage is a shorthand for m.Type(websocket.BinaryMessage).
func (m *WebsocketMessage) BinaryMessage() *WebsocketMessage {
	m.chain.enter("BinaryMessage()")
	defer m.chain.leave()

	m.checkType(websocket.BinaryMessage)

	return m
}

// NotBinaryMessage is a shorthand for m.NotType(websocket.BinaryMessage).
func (m *WebsocketMessage) NotBinaryMessage() *WebsocketMessage {
	m.chain.enter("NotBinaryMessage()")
	defer m.chain.leave()

	m.checkNotType(websocket.BinaryMessage)

	return m
}

// TextMessage is a shorthand for m.Type(websocket.TextMessage).
func (m *WebsocketMessage) TextMessage() *WebsocketMessage {
	m.chain.enter("TextMessage()")
	defer m.chain.leave()

	m.checkType(websocket.TextMessage)

	return m
}

// NotTextMessage is a shorthand for m.NotType(websocket.TextMessage).
func (m *WebsocketMessage) NotTextMessage() *WebsocketMessage {
	m.chain.enter("NotTextMessage()")
	defer m.chain.leave()

	m.checkNotType(websocket.TextMessage)

	return m
}

// Type succeeds if WebSocket message type is one of the given.
//
// WebSocket message types are defined in RFC 6455, section 11.8.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//
//	msg := conn.Expect()
//	msg.Type(websocket.TextMessage, websocket.BinaryMessage)
func (m *WebsocketMessage) Type(typ ...int) *WebsocketMessage {
	m.chain.enter("Type()")
	defer m.chain.leave()

	m.checkType(typ...)

	return m
}

// NotType succeeds if WebSocket message type is none of the given.
//
// WebSocket message types are defined in RFC 6455, section 11.8.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//
//	msg := conn.Expect()
//	msg.NotType(websocket.CloseMessage, websocket.BinaryMessage)
func (m *WebsocketMessage) NotType(typ ...int) *WebsocketMessage {
	m.chain.enter("NotType()")
	defer m.chain.leave()

	m.checkNotType(typ...)

	return m
}

func (m *WebsocketMessage) checkType(typ ...int) {
	if m.chain.failed() {
		return
	}

	if len(typ) == 0 {
		m.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("missing type argument"),
			},
		})
		return
	}

	found := false
	for _, t := range typ {
		if t == m.typ {
			found = true
			break
		}
	}

	if !found {
		if len(typ) == 1 {
			m.chain.fail(AssertionFailure{
				Type:     AssertEqual,
				Actual:   &AssertionValue{wsMessageType(m.typ)},
				Expected: &AssertionValue{wsMessageType(typ[0])},
				Errors: []error{
					errors.New("expected: message types are equal"),
				},
			})
		} else {
			m.chain.fail(AssertionFailure{
				Type:     AssertBelongs,
				Actual:   &AssertionValue{wsMessageType(m.typ)},
				Expected: &AssertionValue{AssertionList(wsMessageTypes(typ))},
				Errors: []error{
					errors.New("expected: message type belongs to given list"),
				},
			})
		}
	}
}

func (m *WebsocketMessage) checkNotType(typ ...int) {
	if m.chain.failed() {
		return
	}

	if len(typ) == 0 {
		m.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("missing type argument"),
			},
		})
		return
	}

	found := false
	for _, t := range typ {
		if t == m.typ {
			found = true
			break
		}
	}

	if found {
		if len(typ) == 1 {
			m.chain.fail(AssertionFailure{
				Type:     AssertNotEqual,
				Actual:   &AssertionValue{wsMessageType(m.typ)},
				Expected: &AssertionValue{wsMessageType(typ[0])},
				Errors: []error{
					errors.New("expected: message types are non-equal"),
				},
			})
		} else {
			m.chain.fail(AssertionFailure{
				Type:     AssertNotBelongs,
				Actual:   &AssertionValue{wsMessageType(m.typ)},
				Expected: &AssertionValue{AssertionList(wsMessageTypes(typ))},
				Errors: []error{
					errors.New("expected: message type does not belong to given list"),
				},
			})
		}
	}
}

// Code succeeds if WebSocket close code is one of the given.
//
// Code fails if WebSocket message type is not "8 - Connection Close Frame".
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//
//	msg := conn.Expect().Closed()
//	msg.Code(websocket.CloseNormalClosure, websocket.CloseGoingAway)
func (m *WebsocketMessage) Code(code ...int) *WebsocketMessage {
	m.chain.enter("Code()")
	defer m.chain.leave()

	if m.chain.failed() {
		return m
	}

	if len(code) == 0 {
		m.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("missing code argument"),
			},
		})
		return m
	}

	if m.typ != websocket.CloseMessage {
		m.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{wsMessageType(m.typ)},
			Expected: &AssertionValue{wsMessageType(websocket.CloseMessage)},
			Errors: []error{
				errors.New("expected: close message"),
			},
		})
		return m
	}

	found := false
	for _, c := range code {
		if c == m.closeCode {
			found = true
			break
		}
	}

	if !found {
		if len(code) == 1 {
			m.chain.fail(AssertionFailure{
				Type:     AssertEqual,
				Actual:   &AssertionValue{wsCloseCode(m.closeCode)},
				Expected: &AssertionValue{wsCloseCode(code[0])},
				Errors: []error{
					errors.New("expected: close codes are equal"),
				},
			})
		} else {
			m.chain.fail(AssertionFailure{
				Type:     AssertBelongs,
				Actual:   &AssertionValue{wsCloseCode(m.closeCode)},
				Expected: &AssertionValue{AssertionList(wsCloseCodes(code))},
				Errors: []error{
					errors.New("expected: close code belongs to given list"),
				},
			})
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
//
//	msg := conn.Expect().Closed()
//	msg.NotCode(websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived)
func (m *WebsocketMessage) NotCode(code ...int) *WebsocketMessage {
	m.chain.enter("NotCode()")
	defer m.chain.leave()

	if m.chain.failed() {
		return m
	}

	if len(code) == 0 {
		m.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("missing code argument"),
			},
		})
		return m
	}

	if m.typ != websocket.CloseMessage {
		m.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{wsMessageType(m.typ)},
			Expected: &AssertionValue{wsMessageType(websocket.CloseMessage)},
			Errors: []error{
				errors.New("expected: close message"),
			},
		})
		return m
	}

	found := false
	for _, c := range code {
		if c == m.closeCode {
			found = true
			break
		}
	}

	if found {
		if len(code) == 1 {
			m.chain.fail(AssertionFailure{
				Type:     AssertNotEqual,
				Actual:   &AssertionValue{wsCloseCode(m.closeCode)},
				Expected: &AssertionValue{wsCloseCode(code[0])},
				Errors: []error{
					errors.New("expected: close codes are non-equal"),
				},
			})
		} else {
			m.chain.fail(AssertionFailure{
				Type:     AssertNotBelongs,
				Actual:   &AssertionValue{wsCloseCode(m.closeCode)},
				Expected: &AssertionValue{AssertionList(wsCloseCodes(code))},
				Errors: []error{
					errors.New("expected: close code dose not belong to given list"),
				},
			})
		}
	}

	return m
}

// Body returns a new String instance with WebSocket message content.
//
// Example:
//
//	msg := conn.Expect()
//	msg.Body().NotEmpty()
//	msg.Body().Length().Equal(100)
func (m *WebsocketMessage) Body() *String {
	m.chain.enter("Body()")
	defer m.chain.leave()

	if m.chain.failed() {
		return newString(m.chain, "")
	}

	return newString(m.chain, string(m.content))
}

// NoContent succeeds if WebSocket message has no content (is empty).
func (m *WebsocketMessage) NoContent() *WebsocketMessage {
	m.chain.enter("NoContent()")
	defer m.chain.leave()

	if m.chain.failed() {
		return m
	}

	if !(len(m.content) == 0) {
		var actual interface{}
		switch m.typ {
		case websocket.BinaryMessage:
			actual = m.content

		default:
			actual = string(m.content)
		}

		m.chain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{actual},
			Errors: []error{
				errors.New("expected: message content is empty"),
			},
		})
	}

	return m
}

// JSON returns a new Value instance with JSON contents of WebSocket message.
//
// JSON succeeds if JSON may be decoded from message content.
//
// Example:
//
//	msg := conn.Expect()
//	msg.JSON().Array().Elements("foo", "bar")
func (m *WebsocketMessage) JSON() *Value {
	m.chain.enter("JSON()")
	defer m.chain.leave()

	if m.chain.failed() {
		return newValue(m.chain, nil)
	}

	var value interface{}

	if err := json.Unmarshal(m.content, &value); err != nil {
		m.chain.fail(AssertionFailure{
			Type: AssertValid,
			Actual: &AssertionValue{
				string(m.content),
			},
			Errors: []error{
				errors.New("failed to decode json"),
				err,
			},
		})
		return newValue(m.chain, nil)
	}

	return newValue(m.chain, value)
}
