package httpexpect

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

type (
	WsMessage struct {
		chain     chain
		typ       int
		body      []byte
		closeCode int
	}

	WsTextMessage struct {
		msg *WsMessage
	}

	WsBinaryMessage struct {
		msg *WsMessage
	}

	WsCloseMessage struct {
		msg *WsMessage
	}
)

func (m *WsMessage) Message() (msg *WsTextMessage) {
	msg = &WsTextMessage{msg: m}
	switch {
	case m.chain.failed():
		return
	case m.typ != websocket.TextMessage:
		m.chain.fail(
			"\nexpected WebSocket message type:\n %s\n\nbut got:\n %s",
			wsMessageTypeName(websocket.TextMessage),
			wsMessageTypeName(m.typ))
		return
	}
	return
}

func (m *WsMessage) Binary() (msg *WsBinaryMessage) {
	msg = &WsBinaryMessage{msg: m}
	switch {
	case m.chain.failed():
		return
	case m.typ != websocket.BinaryMessage:
		m.chain.fail(
			"\nexpected WebSocket message type:\n %s\n\nbut got:\n %s",
			wsMessageTypeName(websocket.BinaryMessage),
			wsMessageTypeName(m.typ))
		return
	}
	return
}

func (m *WsMessage) Closed() (msg *WsCloseMessage) {
	msg = &WsCloseMessage{msg: m}
	switch {
	case m.chain.failed():
		return
	case m.typ != websocket.CloseMessage:
		m.chain.fail(
			"\nexpected WebSocket message type:\n %s\n\nbut got:\n %s",
			wsMessageTypeName(websocket.CloseMessage),
			wsMessageTypeName(m.typ))
		return
	}
	return
}

func (m *WsMessage) Body() *String {
	return &String{m.chain, string(m.body)}
}

func (m *WsMessage) JSON() *Value {
	return &Value{m.chain, m.getJSON()}
}

func (m *WsMessage) NoContent() *WsMessage {
	switch {
	case m.chain.failed():
		return m
	case len(m.body) == 0:
		return m
	}
	switch m.typ {
	case websocket.BinaryMessage:
		m.chain.fail(
			"\nexpected message body being empty, but got:\n%d bytes",
			len(m.body))
	default:
		m.chain.fail(
			"\nexpected message body being empty, but got:\n%s",
			string(m.body))
	}
	return m
}

func (m *WsMessage) Raw() (typ int, body []byte, closeCode int) {
	return m.typ, m.body, m.closeCode
}

func (m *WsTextMessage) Body() *String {
	return m.msg.Body()
}

func (m *WsTextMessage) JSON() *Value {
	return m.msg.JSON()
}

func (m *WsTextMessage) NoContent() *WsTextMessage {
	m.msg.NoContent()
	return m
}

func (m *WsTextMessage) Raw() string {
	return string(m.msg.body)
}

func (m *WsBinaryMessage) Body() *Value {
	return &Value{m.msg.chain, m.msg.body}
}

func (m *WsBinaryMessage) NoContent() *WsBinaryMessage {
	m.msg.NoContent()
	return m
}

func (m *WsBinaryMessage) Raw() []byte {
	return m.msg.body
}

func (m *WsCloseMessage) Code(code ...int) *WsCloseMessage {
	switch {
	case m.msg.chain.failed():
		return m
	case len(code) == 0:
		m.msg.chain.fail("\nunexpected nil argument passed to Code")
		return m
	}
	yes := false
	for _, c := range code {
		if c == m.msg.closeCode {
			yes = true
			break
		}
	}
	if !yes {
		m.msg.chain.fail(
			"\nexpected close code equal to one of:\n%v\n\nbut got:\n%d",
			code, m.msg.closeCode)
	}
	return m
}

func (m *WsCloseMessage) CodeNotEqual(code ...int) *WsCloseMessage {
	switch {
	case m.msg.chain.failed():
		return m
	case len(code) == 0:
		m.msg.chain.fail("\nunexpected nil argument passed to CodeNotEqual")
		return m
	}
	for _, c := range code {
		if c == m.msg.closeCode {
			m.msg.chain.fail(
				"\nexpected close code not equal:\n%v\n\nbut got:\n%d",
				code, m.msg.closeCode)
			return m
		}
	}
	return m
}

func (m *WsCloseMessage) Body() *String {
	return m.msg.Body()
}

func (m *WsCloseMessage) JSON() *Value {
	return m.msg.JSON()
}

func (m *WsCloseMessage) NoContent() *WsCloseMessage {
	m.msg.NoContent()
	return m
}

func (m *WsCloseMessage) Raw() *websocket.CloseError {
	return &websocket.CloseError{
		Code: m.msg.closeCode,
		Text: string(m.msg.body),
	}
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

func (m *WsMessage) getJSON() interface{} {
	if m.chain.failed() {
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(m.body, &value); err != nil {
		m.chain.fail(err.Error())
		return nil
	}

	return value
}
