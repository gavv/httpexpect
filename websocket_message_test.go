package httpexpect

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsocketMessage_Failed(t *testing.T) {
	chain := newMockChain(t, flagFailed)

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
	msg.Alias("foo")

	msg.Body().chain.assert(t, failure)
	msg.JSON().chain.assert(t, failure)
}

func TestWebsocketMessage_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil)
		msg.CloseMessage()
		msg.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		msg := NewWebsocketMessageC(Config{
			Reporter: reporter,
		}, websocket.CloseMessage, nil)
		msg.CloseMessage()
		msg.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newWebsocketMessage(chain, 0, nil)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestWebsocketMessage_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewWebsocketMessage(reporter, websocket.CloseMessage, nil)
	assert.Equal(t, []string{"WebsocketMessage()"}, value.chain.context.Path)
	assert.Equal(t, []string{"WebsocketMessage()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"WebsocketMessage()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Body()
	assert.Equal(t, []string{"WebsocketMessage()", "Body()"},
		childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Body()"}, childValue.chain.context.AliasedPath)
}

func TestWebsocketMessage_CloseMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 1000)

	msg.CloseMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotCloseMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.BinaryMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotBinaryMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.TextMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotTextMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Type(websocket.CloseMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotType(websocket.CloseMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Type(websocket.TextMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType(websocket.TextMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Code(1000)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotCode(1000)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Code(1001)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCode(1001)
	msg.chain.assert(t, success)
	msg.chain.clear()
}

func TestWebsocketMessage_TextMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCloseMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.BinaryMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotBinaryMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.TextMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotTextMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Type(websocket.CloseMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType(websocket.CloseMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Type(websocket.TextMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotType(websocket.TextMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()
}

func TestWebsocketMessage_BinaryMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.BinaryMessage, nil, 0)

	msg.CloseMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCloseMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.BinaryMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotBinaryMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.TextMessage()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotTextMessage()
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Type(websocket.BinaryMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotType(websocket.BinaryMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Type(websocket.TextMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType(websocket.TextMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()
}

func TestWebsocketMessage_MatchTypes(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 0)

	msg.Type(websocket.TextMessage, websocket.BinaryMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Type(websocket.BinaryMessage, websocket.TextMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Type(websocket.CloseMessage, websocket.BinaryMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Type(websocket.BinaryMessage, websocket.CloseMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType(websocket.TextMessage, websocket.BinaryMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType(websocket.BinaryMessage, websocket.TextMessage)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType(websocket.CloseMessage, websocket.BinaryMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotType(websocket.BinaryMessage, websocket.CloseMessage)
	msg.chain.assert(t, success)
	msg.chain.clear()
}

func TestWebsocketMessage_MatchCodes(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10)

	msg.Code(10, 20)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Code(20, 10)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.Code(30, 20)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Code(20, 30)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCode(10, 20)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCode(20, 10)
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCode(30, 20)
	msg.chain.assert(t, success)
	msg.chain.clear()

	msg.NotCode(20, 30)
	msg.chain.assert(t, success)
	msg.chain.clear()
}

func TestWebsocketMessage_CodeAndType(t *testing.T) {
	cases := []struct {
		name        string
		typ         int
		code        int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "text message with close code 10",
			typ:         websocket.TextMessage,
			code:        10,
			wantCode:    failure,
			wantNotCode: failure,
		},
		{
			name:        "close message with close code 10",
			typ:         websocket.CloseMessage,
			code:        10,
			wantCode:    success,
			wantNotCode: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).Code(tc.code).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).NotCode(tc.code).
				chain.assert(t, tc.wantNotCode)
		})
	}
}

func TestWebsocketMessage_NoContent(t *testing.T) {
	cases := []struct {
		name    string
		typ     int
		content []byte
		result  chainResult
	}{
		{
			name:    "nil text message",
			typ:     websocket.TextMessage,
			content: nil,
			result:  success,
		},
		{
			name:    "empty text message",
			typ:     websocket.TextMessage,
			content: []byte(""),
			result:  success,
		},
		{
			name:    "text message with content",
			typ:     websocket.TextMessage,
			content: []byte("test"),
			result:  failure,
		},
		{
			name:    "nil binary message",
			typ:     websocket.BinaryMessage,
			content: nil,
			result:  success,
		},
		{
			name:    "empty binary message",
			typ:     websocket.BinaryMessage,
			content: []byte(""),
			result:  success,
		},
		{
			name:    "binary message with content",
			typ:     websocket.BinaryMessage,
			content: []byte("test"),
			result:  failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, tc.content).NoContent().
				chain.assert(t, tc.result)
		})
	}
}

func TestWebsocketMessage_Body(t *testing.T) {
	reporter := newMockReporter(t)

	body := []byte("test")

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

	s := msg.Body()
	s.chain.assert(t, success)

	require.Equal(t, "test", s.Raw())
}

func TestWebsocketMessage_JSON(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("good", func(t *testing.T) {
		body := []byte(`{"foo":"bar"}`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assert(t, success)

		require.Equal(t, "bar", j.Object().Value("foo").Raw())
	})

	t.Run("bad", func(t *testing.T) {
		body := []byte(`{`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assert(t, failure)

		msg.chain.assert(t, failure)
	})
}

func TestWebsocketMessage_Usage(t *testing.T) {
	chain := newMockChain(t)

	msg := newEmptyWebsocketMessage(chain)

	msg.chain.assert(t, success)

	msg.Type()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotType()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.Code()
	msg.chain.assert(t, failure)
	msg.chain.clear()

	msg.NotCode()
	msg.chain.assert(t, failure)
	msg.chain.clear()
}

func TestWebsocketMessage_Codes(t *testing.T) {
	t.Run("message type", func(t *testing.T) {
		for n := 0; n < 100; n++ {
			assert.NotEmpty(t, wsMessageType(n).String())
		}
	})

	t.Run("close code", func(t *testing.T) {
		for n := 0; n < 2000; n++ {
			assert.NotEmpty(t, wsCloseCode(n).String())
		}
	})
}
