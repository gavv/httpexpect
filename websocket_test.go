package httpexpect

import (
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func noWsPreSteps(ws *Websocket) {}

func TestWebsocket_FailedChain(t *testing.T) {
	reporter := newMockReporter(t)
	config := newMockConfig(reporter)
	chain := newChainWithDefaults("test", reporter, flagFailed)

	ws := newWebsocket(chain, config, nil)

	ws.Conn()
	ws.Raw()
	ws.Alias("foo")
	ws.WithReadTimeout(0)
	ws.WithoutReadTimeout()
	ws.WithWriteTimeout(0)
	ws.WithoutWriteTimeout()

	ws.Subprotocol().chain.assert(t, failure)
	ws.Expect().chain.assert(t, failure)

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
	ws.Close()
}

func TestWebsocket_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewWebsocketC(Config{Reporter: reporter}, &mockWebsocketConn{})
	assert.Equal(t, []string{"Websocket()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Websocket()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Websocket()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestWebsocket_NilConn(t *testing.T) {
	config := Config{
		Reporter: newMockReporter(t),
	}

	t.Run("getters", func(t *testing.T) {
		ws := NewWebsocketC(config, nil)

		if ws.Conn() != nil {
			t.Fatal("Conn returned not nil")
		}

		if ws.Raw() != nil {
			t.Fatal("Raw returned not nil")
		}

		if ws.Subprotocol() == nil {
			t.Fatal("Subprotocol returned nil")
		}

		ws.chain.assert(t, success)
	})

	t.Run("expect", func(t *testing.T) {
		ws := NewWebsocketC(config, nil)

		msg := ws.Expect()
		msg.chain.assert(t, failure)

		ws.chain.assert(t, failure)
	})
}

func TestWebsocket_MockConn(t *testing.T) {
	reporter := newMockReporter(t)

	config := Config{
		Reporter: reporter,
	}

	t.Run("getters", func(t *testing.T) {
		ws := NewWebsocketC(config, &mockWebsocketConn{})

		if ws.Conn() == nil {
			t.Fatal("Conn returned nil")
		}

		if ws.Raw() != nil {
			t.Fatal("Raw returned not nil")
		}

		if ws.Subprotocol() == nil {
			t.Fatal("Subprotocol returned nil")
		}

		ws.chain.assert(t, success)
	})

	t.Run("expect", func(t *testing.T) {
		ws := NewWebsocketC(config, &mockWebsocketConn{})

		msg := ws.Expect()
		msg.chain.assert(t, success)

		ws.chain.assert(t, success)
	})
}

func TestWebsocket_Expect(t *testing.T) {
	type args struct {
		chainFlags chainFlags
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
			},
			result: success,
		},
		{
			name: "fail to read message from conn",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn: &mockWebsocketConn{
					readMsgErr: errors.New("failed to read message"),
				},
			},
			result: failure,
		},
		{
			name: "chain already failed",
			args: args{
				chainFlags: flagFailed,
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
			},
			result: failure,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
			},
			result: failure,
		},
		{
			name: "connection closed",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn: &mockWebsocketConn{},
			},
			result: failure,
		},
		{
			name: "failed to set read deadline",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn: &mockWebsocketConn{
					readDlError: errors.New("failed to set read deadline"),
				},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			config := newMockConfig(reporter)
			chain := newChainWithDefaults("test", reporter, tc.args.chainFlags)

			ws := newWebsocket(chain, config, tc.args.wsConn)
			tc.args.wsPreSteps(ws)

			ws.Expect()

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_Close(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		closeCode  []int
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    &mockWebsocketConn{},
				closeCode: []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.Close(tc.args.closeCode...)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_CloseWithBytes(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    []byte
		closeCode  []int
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    &mockWebsocketConn{},
				content:   []byte("connection closed..."),
				closeCode: []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.CloseWithBytes(tc.args.content, tc.args.closeCode...)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_CloseWithText(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    string
		closeCode  []int
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    &mockWebsocketConn{},
				content:   "connection closed...",
				closeCode: []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.CloseWithText(tc.args.content, tc.args.closeCode...)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_CloseWithJSON(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    interface{}
		closeCode  []int
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure},
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn: &mockWebsocketConn{},
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			result: failure,
		},
		{
			name: "marshall failed",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    make(chan int),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.CloseWithJSON(tc.args.content, tc.args.closeCode...)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_WriteMessage(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		typ        int
		content    []byte
		closeCode  []int
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "text message success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			result: success,
		},
		{
			name: "text message fail nil conn",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			result: failure,
		},
		{
			name: "text message fail unusable",
			args: args{
				wsConn: &mockWebsocketConn{},
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				typ:       websocket.TextMessage,
				content:   []byte("random message..."),
				closeCode: []int{},
			},
			result: failure,
		},
		{
			name: "text message failed to set write deadline",
			args: args{
				wsConn: &mockWebsocketConn{
					writeDlError: errors.New("failed to set write deadline"),
				},
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			result: failure,
		},
		{
			name: "text message failed to write to conn",
			args: args{
				wsConn: &mockWebsocketConn{
					writeMsgErr: errors.New("failed to write message to conn"),
				},
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			result: failure,
		},
		{
			name: "text binary message success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				typ:        websocket.BinaryMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			result: success,
		},
		{
			name: "close message success",
			args: args{
				wsConn:     &mockWebsocketConn{},
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMessage,
				content:    []byte("closing message..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			result: success,
		},
		{
			name: "close message too many close codes",
			args: args{
				wsConn:     &mockWebsocketConn{},
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMessage,
				content:    []byte("closing message..."),
				closeCode: []int{
					websocket.CloseNormalClosure,
					websocket.CloseAbnormalClosure,
				},
			},
			result: failure,
		},
		{
			name: "unsupported message type",
			args: args{
				wsConn:     &mockWebsocketConn{},
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMandatoryExtension,
				content:    []byte("unsupported message..."),
				closeCode:  []int{},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.WriteMessage(tc.args.typ, tc.args.content, tc.args.closeCode...)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_WriteBytesBinary(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    []byte
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    []byte("random message..."),
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content:    []byte("random message..."),
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: &mockWebsocketConn{},
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: []byte("random message..."),
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.WriteBytesBinary(tc.args.content)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_WriteBytesText(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    []byte
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    []byte("random message..."),
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content:    []byte("random message..."),
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: &mockWebsocketConn{},
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: []byte("random message..."),
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.WriteBytesText(tc.args.content)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_WriteText(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    string
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content:    "random message...",
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content:    "random message...",
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: &mockWebsocketConn{},
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: "random message...",
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.WriteText(tc.args.content)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_WriteJSON(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    interface{}
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     &mockWebsocketConn{},
				content: map[string]string{
					"msg": "random message",
				},
			},
			result: success,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content: map[string]string{
					"msg": "random message",
				},
			},
			result: failure,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: &mockWebsocketConn{},
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: map[string]string{
					"msg": "random message",
				},
			},
			result: failure,
		},
		{
			name: "JSON marshal failed",
			args: args{
				wsConn:     &mockWebsocketConn{},
				wsPreSteps: noWsPreSteps,
				content:    make(chan int),
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			tc.args.wsPreSteps(ws)

			ws.WriteJSON(tc.args.content)

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_Subprotocol(t *testing.T) {
	subproto := "soap"
	ws := NewWebsocketC(
		Config{
			Reporter: NewAssertReporter(t),
		},
		&mockWebsocketConn{
			subprotocol: subproto,
		})

	ws.Subprotocol()

	if got := ws.Subprotocol().value; got != subproto {
		t.Errorf("Websocket.Subprotocol() = %v, want %v", got, subproto)
	}
}

func TestWebsocket_SetReadDeadline(t *testing.T) {
	type args struct {
		wsConn WebsocketConn
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsConn: &mockWebsocketConn{},
			},
			result: success,
		},
		{
			name: "conn.SetReadDeadline error",
			args: args{
				wsConn: &mockWebsocketConn{
					readDlError: errors.New("Failed to set read deadline"),
				},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn).
				WithReadTimeout(time.Second)

			opChain := ws.chain.enter("test")
			ws.setReadDeadline(opChain)
			opChain.leave()

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_SetWriteDeadline(t *testing.T) {
	type args struct {
		wsConn WebsocketConn
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsConn: &mockWebsocketConn{},
			},
			result: success,
		},
		{
			name: "conn.SetReadDeadline error",
			args: args{
				wsConn: &mockWebsocketConn{
					writeDlError: errors.New("Failed to set read deadline"),
				},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn).
				WithWriteTimeout(time.Second)

			opChain := ws.chain.enter("test")
			ws.setWriteDeadline(opChain)
			opChain.leave()

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_Disconnect(t *testing.T) {
	type args struct {
		wsConn WebsocketConn
	}
	cases := []struct {
		name   string
		args   args
		result chainResult
	}{
		{
			name: "success",
			args: args{
				wsConn: &mockWebsocketConn{},
			},
			result: success,
		},
		{
			name: "success even if conn is nil",
			args: args{
				wsConn: nil,
			},
			result: success,
		},
		{
			name: "conn close failed",
			args: args{
				wsConn: &mockWebsocketConn{
					closeError: errors.New("failed to close ws conn"),
				},
			},
			result: failure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tc.args.wsConn)

			ws.Disconnect()

			ws.chain.assert(t, tc.result)
		})
	}
}

func TestWebsocket_Printer(t *testing.T) {
	t.Run("print read", func(t *testing.T) {
		reporter := newMockReporter(t)
		printer := &mockWebsocketPrinter{}
		config := Config{
			Reporter: reporter,
			Printers: []Printer{printer},
		}.withDefaults()
		ws := newWebsocket(newMockChain(t), config, &mockWebsocketConn{})

		ws.printRead(websocket.CloseMessage,
			[]byte("random message"),
			websocket.CloseNormalClosure)

		if !printer.isReadFrom {
			t.Errorf("Websocket.printRead() failed to read from printer")
		}
	})

	t.Run("print write", func(t *testing.T) {
		reporter := newMockReporter(t)
		printer := &mockWebsocketPrinter{}
		config := Config{
			Reporter: reporter,
			Printers: []Printer{printer},
		}.withDefaults()
		ws := newWebsocket(newMockChain(t), config, &mockWebsocketConn{})

		ws.printWrite(websocket.CloseMessage,
			[]byte("random message"),
			websocket.CloseNormalClosure)

		if !printer.isWrittenTo {
			t.Errorf("Websocket.printWrite() failed to write to printer")
		}
	})
}
