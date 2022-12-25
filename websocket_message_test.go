package httpexpect

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestWebsocketMessageFailed(t *testing.T) {
	chain := newMockChain(t)
	chain.fail(mockFailure())

	msg := newEmptyWebsocketMessage(chain)

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

func TestWebsocketMessageConstructors(t *testing.T) {
	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil)
		msg.CloseMessage()
		msg.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		msg := NewWebsocketMessageC(Config{
			Reporter: reporter,
		}, websocket.CloseMessage, nil)
		msg.CloseMessage()
		msg.chain.assertNotFailed(t)
	})
}

func TestWebsocketMessageBadUsage(t *testing.T) {
	chain := newMockChain(t)

	msg := newEmptyWebsocketMessage(chain)

	msg.chain.assertNotFailed(t)

	msg.Type()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Code()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCode()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()
}

func TestWebsocketMessageCloseMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 1000)

	msg.CloseMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotCloseMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.BinaryMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotBinaryMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.TextMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotTextMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.CloseMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.CloseMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.TextMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Code(1000)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotCode(1000)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Code(1001)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCode(1001)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()
}

func TestWebsocketMessageTextMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCloseMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.BinaryMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotBinaryMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.TextMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotTextMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.CloseMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.CloseMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.TextMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()
}

func TestWebsocketMessageBinaryMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.BinaryMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCloseMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.BinaryMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotBinaryMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.TextMessage()
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotTextMessage()
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.BinaryMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.BinaryMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.TextMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()
}

func TestWebsocketMessageMatchTypes(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 0)

	msg.Type(websocket.TextMessage, websocket.BinaryMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.BinaryMessage, websocket.TextMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.CloseMessage, websocket.BinaryMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Type(websocket.BinaryMessage, websocket.CloseMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.TextMessage, websocket.BinaryMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.BinaryMessage, websocket.TextMessage)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.CloseMessage, websocket.BinaryMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotType(websocket.BinaryMessage, websocket.CloseMessage)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()
}

func TestWebsocketMessageMatchCodes(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10)

	msg.Code(10, 20)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Code(20, 10)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.Code(30, 20)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.Code(20, 30)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCode(10, 20)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCode(20, 10)
	msg.chain.assertFailed(t)
	msg.chain.clearFailed()

	msg.NotCode(30, 20)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()

	msg.NotCode(20, 30)
	msg.chain.assertNotFailed(t)
	msg.chain.clearFailed()
}

func TestWebsocketMessageCodeAndType(t *testing.T) {
	reporter := newMockReporter(t)

	m1 := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 10)

	m1.Code(10)
	m1.chain.assertFailed(t)
	m1.chain.clearFailed()

	m1.NotCode(10)
	m1.chain.assertFailed(t)
	m1.chain.clearFailed()

	m2 := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10)

	m2.Code(10)
	m2.chain.assertNotFailed(t)
	m2.chain.clearFailed()

	m2.NotCode(10)
	m2.chain.assertFailed(t)
	m2.chain.clearFailed()
}

func TestWebsocketMessageNoContent(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("text", func(t *testing.T) {
		m1 := NewWebsocketMessage(reporter, websocket.TextMessage, nil)
		m1.NoContent()
		m1.chain.assertNotFailed(t)

		m2 := NewWebsocketMessage(reporter, websocket.TextMessage, []byte(""))
		m2.NoContent()
		m2.chain.assertNotFailed(t)

		m3 := NewWebsocketMessage(reporter, websocket.TextMessage, []byte("test"))
		m3.NoContent()
		m3.chain.assertFailed(t)
	})

	t.Run("binary", func(t *testing.T) {
		m1 := NewWebsocketMessage(reporter, websocket.BinaryMessage, nil)
		m1.NoContent()
		m1.chain.assertNotFailed(t)

		m2 := NewWebsocketMessage(reporter, websocket.BinaryMessage, []byte(""))
		m2.NoContent()
		m2.chain.assertNotFailed(t)

		m3 := NewWebsocketMessage(reporter, websocket.BinaryMessage, []byte("test"))
		m3.NoContent()
		m3.chain.assertFailed(t)
	})
}

func TestWebsocketMessageBody(t *testing.T) {
	reporter := newMockReporter(t)

	body := []byte("test")

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

	s := msg.Body()
	s.chain.assertNotFailed(t)

	require.Equal(t, "test", s.Raw())
}

func TestWebsocketMessageJSON(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("good", func(t *testing.T) {
		body := []byte(`{"foo":"bar"}`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assertNotFailed(t)

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
