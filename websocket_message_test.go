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

	msg := NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10)

	cases := []struct {
		name   string
		fn     func() *WebsocketMessage
		testFn func() *WebsocketMessage
		want   chainResult
	}{
		{
			name:   "close message type with CloseMessage function",
			fn:     msg.CloseMessage,
			testFn: msg.CloseMessage,
			want:   success,
		},
		{
			name:   "close message type with NotCloseMessage function",
			fn:     msg.CloseMessage,
			testFn: msg.NotCloseMessage,
			want:   failure,
		},
		{
			name:   "close message type with BinaryMessage function",
			fn:     msg.CloseMessage,
			testFn: msg.BinaryMessage,
			want:   failure,
		},
		{
			name:   "close message type with NotBinaryMessage function",
			fn:     msg.CloseMessage,
			testFn: msg.NotBinaryMessage,
			want:   success,
		},
		{
			name:   "close message type with TextMessage function",
			fn:     msg.CloseMessage,
			testFn: msg.TextMessage,
			want:   failure,
		},
		{
			name:   "close message type with NotTextMessage function",
			fn:     msg.CloseMessage,
			testFn: msg.NotTextMessage,
			want:   success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFn().chain.assert(t, tc.want)
			msg.chain.clear()
		})
	}

	cases2 := []struct {
		name        string
		typ         int
		testTyp     int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "close message type with CloseMessage type",
			typ:         websocket.CloseMessage,
			testTyp:     websocket.CloseMessage,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "close message type with TextMessage type",
			typ:         websocket.CloseMessage,
			testTyp:     websocket.TextMessage,
			wantCode:    failure,
			wantNotCode: success,
		},
	}
	for _, tc := range cases2 {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, 10).Type(tc.testTyp).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, 10).NotType(tc.testTyp).
				chain.assert(t, tc.wantNotCode)
		})

	}

	cases3 := []struct {
		name        string
		typ         int
		code        int
		testCode    int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "close message type with code 1000 and test code 1000",
			typ:         websocket.CloseMessage,
			code:        1000,
			testCode:    1000,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "close message type with code 1000 and test code 1001",
			typ:         websocket.CloseMessage,
			code:        1000,
			testCode:    1001,
			wantCode:    failure,
			wantNotCode: success,
		},
	}

	for _, tc := range cases3 {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).Code(tc.testCode).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).NotCode(tc.testCode).
				chain.assert(t, tc.wantNotCode)
		})
	}

}

func TestWebsocketMessage_TextMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.TextMessage, nil, 10)
	cases := []struct {
		name   string
		fn     func() *WebsocketMessage
		testFn func() *WebsocketMessage
		want   chainResult
	}{
		{
			name:   "text message type with CloseMessage function",
			fn:     msg.TextMessage,
			testFn: msg.CloseMessage,
			want:   failure,
		},
		{
			name:   "text message type with NotClose function",
			fn:     msg.TextMessage,
			testFn: msg.NotCloseMessage,
			want:   success,
		},
		{
			name:   "text message type with BinaryMessage function",
			fn:     msg.TextMessage,
			testFn: msg.BinaryMessage,
			want:   failure,
		},
		{
			name:   "text message type with NotBinary function",
			fn:     msg.TextMessage,
			testFn: msg.NotBinaryMessage,
			want:   success,
		},
		{
			name:   "text message type with TextMessage function",
			fn:     msg.TextMessage,
			testFn: msg.TextMessage,
			want:   success,
		},
		{
			name:   "text message type is with NotTextMessage function",
			fn:     msg.TextMessage,
			testFn: msg.NotTextMessage,
			want:   failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFn().chain.assert(t, tc.want)
			msg.chain.clear()
		})
	}

	cases2 := []struct {
		name        string
		typ         int
		testTyp     int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "text message type with CloseMessage type",
			typ:         websocket.TextMessage,
			testTyp:     websocket.CloseMessage,
			wantCode:    failure,
			wantNotCode: success,
		},
		{
			name:        "text message type with TextMessage type",
			typ:         websocket.TextMessage,
			testTyp:     websocket.TextMessage,
			wantCode:    success,
			wantNotCode: failure,
		},
	}

	for _, tc := range cases2 {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, 10).Type(tc.testTyp).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, 10).NotType(tc.testTyp).
				chain.assert(t, tc.wantNotCode)
		})

	}

}

func TestWebsocketMessage_BinaryMessage(t *testing.T) {
	reporter := newMockReporter(t)

	msg := NewWebsocketMessage(reporter, websocket.BinaryMessage, nil, 10)

	cases := []struct {
		name   string
		fn     func() *WebsocketMessage
		testFn func() *WebsocketMessage
		want   chainResult
	}{
		{
			name:   "binary message type with CloseMessage function",
			fn:     msg.BinaryMessage,
			testFn: msg.CloseMessage,
			want:   failure,
		},
		{
			name:   "binary message with NotCloseMessage function",
			fn:     msg.BinaryMessage,
			testFn: msg.NotCloseMessage,
			want:   success,
		},
		{
			name:   "binary message type with BinaryMessage function",
			fn:     msg.BinaryMessage,
			testFn: msg.BinaryMessage,
			want:   success,
		},
		{
			name:   "binary message type with NotBinaryMessage function",
			fn:     msg.BinaryMessage,
			testFn: msg.NotBinaryMessage,
			want:   failure,
		},
		{
			name:   "binary message type with TextMessage function",
			fn:     msg.BinaryMessage,
			testFn: msg.TextMessage,
			want:   failure,
		},
		{
			name:   "binary message type with NotTextMessage function",
			fn:     msg.BinaryMessage,
			testFn: msg.NotTextMessage,
			want:   success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFn().chain.assert(t, tc.want)
			msg.chain.clear()
		})
	}

	cases2 := []struct {
		name        string
		typ         int
		testTyp     int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "binary message type with BinaryMessage type",
			typ:         websocket.BinaryMessage,
			testTyp:     websocket.BinaryMessage,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "binary message type with TextMessage type",
			typ:         websocket.BinaryMessage,
			testTyp:     websocket.TextMessage,
			wantCode:    failure,
			wantNotCode: success,
		},
	}

	for _, tc := range cases2 {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, 10).Type(tc.testTyp).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, 10).NotType(tc.testTyp).
				chain.assert(t, tc.wantNotCode)
		})

	}
}

func TestWebsocketMessage_MatchTypes(t *testing.T) {
	cases := []struct {
		name        string
		typ         int
		testTyp     int
		testTyp2    int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "text message type with text and binary message types",
			typ:         websocket.TextMessage,
			testTyp:     websocket.TextMessage,
			testTyp2:    websocket.BinaryMessage,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "text message type with binary and text message types",
			typ:         websocket.TextMessage,
			testTyp:     websocket.BinaryMessage,
			testTyp2:    websocket.TextMessage,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "text message type with close and binary message types",
			typ:         websocket.TextMessage,
			testTyp:     websocket.CloseMessage,
			testTyp2:    websocket.BinaryMessage,
			wantCode:    failure,
			wantNotCode: success,
		},
		{
			name:        "text message type with binary and close message types",
			typ:         websocket.TextMessage,
			testTyp:     websocket.BinaryMessage,
			testTyp2:    websocket.CloseMessage,
			wantCode:    failure,
			wantNotCode: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, 0).Type(tc.testTyp, tc.testTyp2).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, 0).NotType(tc.testTyp, tc.testTyp2).
				chain.assert(t, tc.wantNotCode)
		})
	}
}

func TestWebsocketMessage_MatchCodes(t *testing.T) {
	cases := []struct {
		name        string
		typ         int
		code        int
		tc1         int //test code 1
		tc2         int //test code 2
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "close message with close code 10 and with test codes 10 and 20",
			typ:         websocket.CloseMessage,
			code:        10,
			tc1:         10,
			tc2:         20,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "close message with close code 10 and with test codes 20 and 10",
			typ:         websocket.CloseMessage,
			code:        10,
			tc1:         20,
			tc2:         10,
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "close message with close code 10 and with test codes 30 and 20",
			typ:         websocket.CloseMessage,
			code:        10,
			tc1:         30,
			tc2:         20,
			wantCode:    failure,
			wantNotCode: success,
		},
		{
			name:        "close message with close code 10 and with test codes 20 and 30",
			typ:         websocket.CloseMessage,
			code:        10,
			tc1:         20,
			tc2:         30,
			wantCode:    failure,
			wantNotCode: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).Code(tc.tc1, tc.tc2).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).NotCode(tc.tc1, tc.tc2).
				chain.assert(t, tc.wantNotCode)
		})
	}
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
