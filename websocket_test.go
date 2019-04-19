package httpexpect

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebsocketFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	ws := &Websocket{
		chain: chain,
	}

	ws.chain.assertFailed(t)

	ws.Raw()
	ws.WithReadTimeout(0)
	ws.WithoutReadTimeout()
	ws.WithWriteTimeout(0)
	ws.WithoutWriteTimeout()

	ws.Subprotocol().chain.assertFailed(t)
	ws.Expect().chain.assertFailed(t)

	ws.WriteMessage(websocket.TextMessage, []byte("a"))
	ws.WriteBytesBinary([]byte("a"))
	ws.WriteBytesText([]byte("a"))
	ws.WriteText("a")
	ws.WriteJSON(map[string]string{"a": "b"})

	ws.Close()
	ws.CloseWithBytes([]byte("a"))
	ws.CloseWithJSON(map[string]string{"a": "b"})
	ws.CloseWithText("a")

	ws.Disconnect()
}

func TestWebsocketNil(t *testing.T) {
	config := Config{
		Reporter: newMockReporter(t),
	}

	ws := NewWebsocket(config, nil)

	msg := ws.Expect()
	msg.chain.assertFailed(t)

	ws.chain.assertFailed(t)
}
