package httpexpect

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestWebsocketMessageFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	msg := &WebsocketMessage{
		chain: chain,
	}

	msg.chain.assertFailed(t)

	msg.Raw()
	msg.CloseMessage()
	msg.NotCloseMessage()
	msg.BinaryMessage()
	msg.NotBinaryMessage()
	msg.TextMessage()
	msg.NotTextMessage()
	msg.Type(0)
	msg.NotType(0)
	msg.Code(0)
	msg.NotCode(0)
	msg.NoContent()

	msg.Body().chain.assertFailed(t)
	msg.JSON().chain.assertFailed(t)
}

func TestWebsocketMessageBadUsage(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	msg := &WebsocketMessage{
		chain: chain,
	}

	msg.Type()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.Code()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotCode()
	msg.chain.assertFailed(t)
	msg.chain.reset()
}

func TestWebsocketMessageCloseMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotCloseMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.BinaryMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotBinaryMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.TextMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotTextMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Type(websocket.CloseMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotType(websocket.CloseMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.Type(websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType(websocket.TextMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()
}

func TestWebsocketMessageTextMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotCloseMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.BinaryMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotBinaryMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.TextMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotTextMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.Type(websocket.CloseMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType(websocket.CloseMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Type(websocket.TextMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotType(websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()
}

func TestWebsocketMessageBinaryMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.BinaryMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotCloseMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.BinaryMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotBinaryMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.TextMessage()
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotTextMessage()
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Type(websocket.BinaryMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotType(websocket.BinaryMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.Type(websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType(websocket.TextMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()
}

func TestWebsocketMessageMatchTypes(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 0)

	msg.Type(websocket.TextMessage, websocket.BinaryMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Type(websocket.BinaryMessage, websocket.TextMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Type(websocket.CloseMessage, websocket.BinaryMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.Type(websocket.BinaryMessage, websocket.CloseMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType(websocket.TextMessage, websocket.BinaryMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType(websocket.BinaryMessage, websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotType(websocket.CloseMessage, websocket.BinaryMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotType(websocket.BinaryMessage, websocket.CloseMessage)
	msg.chain.assertOK(t)
	msg.chain.reset()
}

func TestWebsocketMessageMatchCodes(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10)

	msg.Code(10, 20)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Code(20, 10)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.Code(30, 20)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.Code(20, 30)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotCode(10, 20)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotCode(20, 10)
	msg.chain.assertFailed(t)
	msg.chain.reset()

	msg.NotCode(30, 20)
	msg.chain.assertOK(t)
	msg.chain.reset()

	msg.NotCode(20, 30)
	msg.chain.assertOK(t)
	msg.chain.reset()
}

func TestWebsocketMessageCodeAndType(t *testing.T) {
	reporter := newMockReporter(t)

	m1 := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 10)

	m1.Code(10)
	m1.chain.assertFailed(t)
	m1.chain.reset()

	m1.NotCode(10)
	m1.chain.assertFailed(t)
	m1.chain.reset()

	m2 := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10)

	m2.Code(10)
	m2.chain.assertOK(t)
	m2.chain.reset()

	m2.NotCode(10)
	m2.chain.assertFailed(t)
	m2.chain.reset()
}

func TestWebsocketMessageNoContent(t *testing.T) {
	reporter := newMockReporter(t)

	m1 := NewWebsocketMessage(reporter, websocket.TextMessage, nil)
	m1.NoContent()
	m1.chain.assertOK(t)

	m2 := NewWebsocketMessage(reporter, websocket.TextMessage, []byte(""))
	m2.NoContent()
	m2.chain.assertOK(t)

	m3 := NewWebsocketMessage(reporter, websocket.TextMessage, []byte("test"))
	m3.NoContent()
	m3.chain.assertFailed(t)
}

func TestWebsocketMessageBody(t *testing.T) {
	reporter := newMockReporter(t)

	body := []byte("test")

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

	s := msg.Body()
	s.chain.assertOK(t)

	require.Equal(t, "test", s.Raw())
}

func TestWebsocketMessageJSON(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("good", func(t *testing.T) {
		body := []byte(`{"foo":"bar"}`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assertOK(t)

		require.Equal(t, "bar", j.Object().Value("foo").Raw())
	})

	t.Run("bad", func(t *testing.T) {
		body := []byte(`{`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assertFailed(t)

		msg.chain.assertFailed(t)
	})
}
