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

func TestWebsocketMessage_CloseCode(t *testing.T) {
	t.Run("close code is not passed", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewWebsocketMessage(reporter, websocket.CloseMessage, nil).Code(0).
			chain.assert(t, success)
		NewWebsocketMessage(reporter, websocket.CloseMessage, nil).NotCode(0).
			chain.assert(t, failure)
	})

	t.Run("single close code is passed", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10).Code(10).
			chain.assert(t, success)
		NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10).NotCode(10).
			chain.assert(t, failure)
	})

	t.Run("multiple close code is passed", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewWebsocketMessage(reporter, websocket.CloseMessage, nil, 10, 20).
			chain.assert(t, failure)
	})
}

func TestWebsocketMessage_CloseMessage(t *testing.T) {
	t.Run("CloseMessage type with functions", func(t *testing.T) {
		cases := []struct {
			name string
			typ  int
			fn   func(s *WebsocketMessage) *WebsocketMessage
			want chainResult
		}{
			{
				name: "CloseMessage type with CloseMessage function",
				typ:  websocket.CloseMessage,
				fn:   (*WebsocketMessage).CloseMessage,
				want: success,
			},
			{
				name: "CloseMessage type with NotCloseMessage function",
				typ:  websocket.CloseMessage,
				fn:   (*WebsocketMessage).NotCloseMessage,
				want: failure,
			},
			{
				name: "CloseMessage type with BinaryMessage function",
				typ:  websocket.CloseMessage,
				fn:   (*WebsocketMessage).BinaryMessage,
				want: failure,
			},
			{
				name: "CloseMessage type with NotBinaryMessage function",
				typ:  websocket.CloseMessage,
				fn:   (*WebsocketMessage).NotBinaryMessage,
				want: success,
			},
			{
				name: "CloseMessage type with TextMessage function",
				typ:  websocket.CloseMessage,
				fn:   (*WebsocketMessage).TextMessage,
				want: failure,
			},
			{
				name: "CloseMessage type with NotTextMessage function",
				typ:  websocket.CloseMessage,
				fn:   (*WebsocketMessage).NotTextMessage,
				want: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				tc.fn(NewWebsocketMessage(reporter, tc.typ, nil, 1000)).
					chain.assert(t, tc.want)
			})
		}
	})

	t.Run("CloseMessage type with codes", func(t *testing.T) {
		cases := []struct {
			name        string
			typ         int
			code        int
			testCode    int
			wantCode    chainResult
			wantNotCode chainResult
		}{
			{
				name:        "CloseMessage type with code 1000 and test code 1000",
				typ:         websocket.CloseMessage,
				code:        1000,
				testCode:    1000,
				wantCode:    success,
				wantNotCode: failure,
			},
			{
				name:        "CloseMessage type with code 1000 and test code 1001",
				typ:         websocket.CloseMessage,
				code:        1000,
				testCode:    1001,
				wantCode:    failure,
				wantNotCode: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewWebsocketMessage(reporter, tc.typ, nil, tc.code).Code(tc.testCode).
					chain.assert(t, tc.wantCode)

				NewWebsocketMessage(reporter, tc.typ, nil, tc.code).NotCode(tc.testCode).
					chain.assert(t, tc.wantNotCode)
			})
		}
	})

	t.Run("CloseMessage type with types", func(t *testing.T) {
		cases := []struct {
			name        string
			typ         int
			code        int
			testCode    int
			wantCode    chainResult
			wantNotCode chainResult
		}{
			{
				name:        "CloseMessage type with code 1000 and test code 1000",
				typ:         websocket.CloseMessage,
				code:        1000,
				testCode:    1000,
				wantCode:    success,
				wantNotCode: failure,
			},
			{
				name:        "CloseMessage type with code 1000 and test code 1001",
				typ:         websocket.CloseMessage,
				code:        1000,
				testCode:    1001,
				wantCode:    failure,
				wantNotCode: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewWebsocketMessage(reporter, tc.typ, nil, tc.code).Code(tc.testCode).
					chain.assert(t, tc.wantCode)

				NewWebsocketMessage(reporter, tc.typ, nil, tc.code).NotCode(tc.testCode).
					chain.assert(t, tc.wantNotCode)
			})
		}
	})
}

func TestWebsocketMessage_TextMessage(t *testing.T) {
	t.Run("TextMessage type with functions", func(t *testing.T) {
		cases := []struct {
			name string
			typ  int
			fn   func(s *WebsocketMessage) *WebsocketMessage
			want chainResult
		}{
			{
				name: "TextMessage type with CloseMessage function",
				typ:  websocket.TextMessage,
				fn:   (*WebsocketMessage).CloseMessage,
				want: failure,
			},
			{
				name: "TextMessage type with NotCloseMessage function",
				typ:  websocket.TextMessage,
				fn:   (*WebsocketMessage).NotCloseMessage,
				want: success,
			},
			{
				name: "TextMessage type with BinaryMessage function",
				typ:  websocket.TextMessage,
				fn:   (*WebsocketMessage).BinaryMessage,
				want: failure,
			},
			{
				name: "TextMessage type with NotBinaryMessage function",
				typ:  websocket.TextMessage,
				fn:   (*WebsocketMessage).NotBinaryMessage,
				want: success,
			},
			{
				name: "TextMessage type with TextMessage function",
				typ:  websocket.TextMessage,
				fn:   (*WebsocketMessage).TextMessage,
				want: success,
			},
			{
				name: "TextMessage type with NotTextMessage function",
				typ:  websocket.TextMessage,
				fn:   (*WebsocketMessage).NotTextMessage,
				want: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				tc.fn(NewWebsocketMessage(reporter, tc.typ, nil, 1000)).
					chain.assert(t, tc.want)
			})
		}
	})

	t.Run("TextMessage type with types", func(t *testing.T) {
		cases := []struct {
			name        string
			typ         int
			testTyp     int
			wantCode    chainResult
			wantNotCode chainResult
		}{
			{
				name:        "TextMessage type with CloseMessage type",
				typ:         websocket.TextMessage,
				testTyp:     websocket.CloseMessage,
				wantCode:    failure,
				wantNotCode: success,
			},
			{
				name:        "TextMessage type with TextMessage type",
				typ:         websocket.TextMessage,
				testTyp:     websocket.TextMessage,
				wantCode:    success,
				wantNotCode: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewWebsocketMessage(reporter, tc.typ, nil, 10).Type(tc.testTyp).
					chain.assert(t, tc.wantCode)

				NewWebsocketMessage(reporter, tc.typ, nil, 10).NotType(tc.testTyp).
					chain.assert(t, tc.wantNotCode)
			})
		}
	})
}

func TestWebsocketMessage_BinaryMessage(t *testing.T) {
	t.Run("BinaryMessage type with functions", func(t *testing.T) {
		cases := []struct {
			name string
			typ  int
			fn   func(s *WebsocketMessage) *WebsocketMessage
			want chainResult
		}{
			{
				name: "BinaryMessage type with CloseMessage function",
				typ:  websocket.BinaryMessage,
				fn:   (*WebsocketMessage).CloseMessage,
				want: failure,
			},
			{
				name: "BinaryMessage type with NotCloseMessage function",
				typ:  websocket.BinaryMessage,
				fn:   (*WebsocketMessage).NotCloseMessage,
				want: success,
			},
			{
				name: "BinaryMessage type with BinaryMessage function",
				typ:  websocket.BinaryMessage,
				fn:   (*WebsocketMessage).BinaryMessage,
				want: success,
			},
			{
				name: "BinaryMessage type with NotBinaryMessage function",
				typ:  websocket.BinaryMessage,
				fn:   (*WebsocketMessage).NotBinaryMessage,
				want: failure,
			},
			{
				name: "BinaryMessage type with TextMessage function",
				typ:  websocket.BinaryMessage,
				fn:   (*WebsocketMessage).TextMessage,
				want: failure,
			},
			{
				name: "BinaryMessage type with NotTextMessage function",
				typ:  websocket.BinaryMessage,
				fn:   (*WebsocketMessage).NotTextMessage,
				want: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				tc.fn(NewWebsocketMessage(reporter, tc.typ, nil, 1000)).
					chain.assert(t, tc.want)
			})
		}
	})

	t.Run("BinaryMessage with types ", func(t *testing.T) {
		cases := []struct {
			name        string
			typ         int
			testTyp     int
			wantCode    chainResult
			wantNotCode chainResult
		}{
			{
				name:        "BinaryMessage type with BinaryMessage type",
				typ:         websocket.BinaryMessage,
				testTyp:     websocket.BinaryMessage,
				wantCode:    success,
				wantNotCode: failure,
			},
			{
				name:        "BinaryMessage type with TextMessage type",
				typ:         websocket.BinaryMessage,
				testTyp:     websocket.TextMessage,
				wantCode:    failure,
				wantNotCode: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewWebsocketMessage(reporter, tc.typ, nil, 10).Type(tc.testTyp).
					chain.assert(t, tc.wantCode)

				NewWebsocketMessage(reporter, tc.typ, nil, 10).NotType(tc.testTyp).
					chain.assert(t, tc.wantNotCode)
			})

		}
	})
}

func TestWebsocketMessage_MatchTypes(t *testing.T) {
	cases := []struct {
		name        string
		typ         int
		testTypes   []int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "TextMessage type with TextMessage and BinaryMessage types",
			typ:         websocket.TextMessage,
			testTypes:   []int{websocket.TextMessage, websocket.BinaryMessage},
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "TextMessage type with BinaryMessage andTexMessage types",
			typ:         websocket.TextMessage,
			testTypes:   []int{websocket.BinaryMessage, websocket.TextMessage},
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "TextMessage type with CloseMessage and BinaryMessage types",
			typ:         websocket.TextMessage,
			testTypes:   []int{websocket.CloseMessage, websocket.BinaryMessage},
			wantCode:    failure,
			wantNotCode: success,
		},
		{
			name:        "TextMessage type with BinaryMessage and CloseMessage types",
			typ:         websocket.TextMessage,
			testTypes:   []int{websocket.BinaryMessage, websocket.CloseMessage},
			wantCode:    failure,
			wantNotCode: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, 0).Type(tc.testTypes...).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, 0).NotType(tc.testTypes...).
				chain.assert(t, tc.wantNotCode)
		})
	}
}

func TestWebsocketMessage_MatchCodes(t *testing.T) {
	cases := []struct {
		name        string
		typ         int
		code        int
		testCodes   []int
		wantCode    chainResult
		wantNotCode chainResult
	}{
		{
			name:        "close message with close code 10 and with test codes 10 and 20",
			typ:         websocket.CloseMessage,
			code:        10,
			testCodes:   []int{10, 20},
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "close message with close code 10 and with test codes 20 and 10",
			typ:         websocket.CloseMessage,
			code:        10,
			testCodes:   []int{20, 10},
			wantCode:    success,
			wantNotCode: failure,
		},
		{
			name:        "close message with close code 10 and with test codes 30 and 20",
			typ:         websocket.CloseMessage,
			code:        10,
			testCodes:   []int{30, 20},
			wantCode:    failure,
			wantNotCode: success,
		},
		{
			name:        "close message with close code 10 and with test codes 20 and 30",
			typ:         websocket.CloseMessage,
			code:        10,
			testCodes:   []int{20, 30},
			wantCode:    failure,
			wantNotCode: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).Code(tc.testCodes...).
				chain.assert(t, tc.wantCode)

			NewWebsocketMessage(reporter, tc.typ, nil, tc.code).NotCode(tc.testCodes...).
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
	t.Run("good", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := []byte(`{"foo":"bar"}`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assert(t, success)

		require.Equal(t, "bar", j.Object().Value("foo").Raw())
	})

	t.Run("bad", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := []byte(`{`)

		msg := NewWebsocketMessage(reporter, websocket.TextMessage, body)

		j := msg.JSON()
		j.chain.assert(t, failure)

		msg.chain.assert(t, failure)
	})
}

func TestWebsocketMessage_Usage(t *testing.T) {
	t.Run("type", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewWebsocketMessage(reporter, websocket.TextMessage, nil).Type().
			chain.assert(t, failure)

		NewWebsocketMessage(reporter, websocket.TextMessage, nil).NotType().
			chain.assert(t, failure)
	})

	t.Run("code", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewWebsocketMessage(reporter, websocket.TextMessage, nil).Code().
			chain.assert(t, failure)

		NewWebsocketMessage(reporter, websocket.TextMessage, nil).NotCode().
			chain.assert(t, failure)
	})
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
