package httpexpect

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func noWsPreSteps(ws *Websocket) {}

func TestWebsocketFailed(t *testing.T) {
	reporter := newMockReporter(t)
	chain := newChainWithDefaults("test", reporter)
	config := newMockConfig(reporter)

	chain.fail(mockFailure())

	ws := newWebsocket(chain, config, nil)

	ws.Conn()
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
	ws.Close()
}

func TestWebsocketNilConn(t *testing.T) {
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

		ws.chain.assertNotFailed(t)
	})

	t.Run("expect", func(t *testing.T) {
		ws := NewWebsocketC(config, nil)

		msg := ws.Expect()
		msg.chain.assertFailed(t)

		ws.chain.assertFailed(t)
	})
}

func TestWebsocketMockConn(t *testing.T) {
	reporter := newMockReporter(t)

	config := Config{
		Reporter: reporter,
	}

	t.Run("getters", func(t *testing.T) {
		ws := NewWebsocketC(config, newMockWebsocketConn())

		if ws.Conn() == nil {
			t.Fatal("Conn returned nil")
		}

		if ws.Raw() != nil {
			t.Fatal("Raw returned not nil")
		}

		if ws.Subprotocol() == nil {
			t.Fatal("Subprotocol returned nil")
		}

		ws.chain.assertNotFailed(t)
	})

	t.Run("expect", func(t *testing.T) {
		ws := NewWebsocketC(config, newMockWebsocketConn())

		msg := ws.Expect()
		msg.chain.assertNotFailed(t)

		ws.chain.assertNotFailed(t)
	})
}

func TestWebsocketExpect(t *testing.T) {
	type args struct {
		failedChain bool
		wsConn      WebsocketConn
		wsPreSteps  func(*Websocket)
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "fail to read message from conn",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn: newMockWebsocketConn().WithReadMsgError(
					fmt.Errorf("failed to read message")),
			},
			assertOk: false,
		},
		{
			name: "chain already failed",
			args: args{
				failedChain: true,
				wsPreSteps:  noWsPreSteps,
				wsConn:      newMockWebsocketConn(),
			},
			assertOk: false,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
			},
			assertOk: false,
		},
		{
			name: "connection closed",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn: newMockWebsocketConn(),
			},
			assertOk: false,
		},
		{
			name: "failed to set read deadline",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn: newMockWebsocketConn().WithReadDlError(
					fmt.Errorf("failed to set read deadline")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			if tt.args.failedChain {
				chain.fail(mockFailure())
			}

			ws := newWebsocket(chain, config, tt.args.wsConn)
			tt.args.wsPreSteps(ws)

			ws.Expect()

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketClose(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		closeCode  []int
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    newMockWebsocketConn(),
				closeCode: []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.Close(tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketCloseWithBytes(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    []byte
		closeCode  []int
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    newMockWebsocketConn(),
				content:   []byte("connection closed..."),
				closeCode: []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.CloseWithBytes(tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketCloseWithText(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    string
		closeCode  []int
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "conn is nil",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     nil,
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    newMockWebsocketConn(),
				content:   "connection closed...",
				closeCode: []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.CloseWithText(tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketCloseWithJSON(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    interface{}
		closeCode  []int
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
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
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn: newMockWebsocketConn(),
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "too many close codes",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content: map[string]string{
					"msg": "connection closing...",
				},
				closeCode: []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			assertOk: false,
		},
		{
			name: "marshall failed",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    make(chan int),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.CloseWithJSON(tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteMessage(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		typ        int
		content    []byte
		closeCode  []int
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "text message success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: true,
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
			assertOk: false,
		},
		{
			name: "text message fail unusable",
			args: args{
				wsConn: newMockWebsocketConn(),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				typ:       websocket.TextMessage,
				content:   []byte("random message..."),
				closeCode: []int{},
			},
			assertOk: false,
		},
		{
			name: "text message failed to set write deadline",
			args: args{
				wsConn: newMockWebsocketConn().WithWriteDlError(
					fmt.Errorf("failed to set write deadline")),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: false,
		},
		{
			name: "text message failed to write to conn",
			args: args{
				wsConn: newMockWebsocketConn().WithWriteMsgError(
					fmt.Errorf("failed to write message to conn")),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: false,
		},
		{
			name: "text binary message success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				typ:        websocket.BinaryMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: true,
		},
		{
			name: "close message success",
			args: args{
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMessage,
				content:    []byte("closing message..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "close message too many close codes",
			args: args{
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMessage,
				content:    []byte("closing message..."),
				closeCode: []int{
					websocket.CloseNormalClosure,
					websocket.CloseAbnormalClosure,
				},
			},
			assertOk: false,
		},
		{
			name: "unsupported message type",
			args: args{
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMandatoryExtension,
				content:    []byte("unsupported message..."),
				closeCode:  []int{},
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteMessage(tt.args.typ, tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteBytesBinary(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    []byte
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("random message..."),
			},
			assertOk: true,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content:    []byte("random message..."),
			},
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: newMockWebsocketConn(),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: []byte("random message..."),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteBytesBinary(tt.args.content)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteBytesText(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    []byte
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("random message..."),
			},
			assertOk: true,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content:    []byte("random message..."),
			},
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: newMockWebsocketConn(),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: []byte("random message..."),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteBytesText(tt.args.content)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteText(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    string
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    "random message...",
			},
			assertOk: true,
		},
		{
			name: "conn is nil",
			args: args{
				wsConn:     nil,
				wsPreSteps: noWsPreSteps,
				content:    "random message...",
			},
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: newMockWebsocketConn(),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: "random message...",
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteText(tt.args.content)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteJSON(t *testing.T) {
	type args struct {
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
		content    interface{}
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content: map[string]string{
					"msg": "random message",
				},
			},
			assertOk: true,
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
			assertOk: false,
		},
		{
			name: "websocket unusable",
			args: args{
				wsConn: newMockWebsocketConn(),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				content: map[string]string{
					"msg": "random message",
				},
			},
			assertOk: false,
		},
		{
			name: "JSON marshal failed",
			args: args{
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				content:    make(chan int),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteJSON(tt.args.content)

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketSubprotocol(t *testing.T) {
	subproto := "soap"
	ws := NewWebsocketC(Config{
		Reporter: NewAssertReporter(t),
	}, newMockWebsocketConn().WithSubprotocol(subproto))

	ws.Subprotocol()

	if got := ws.Subprotocol().value; got != subproto {
		t.Errorf("Websocket.Subprotocol() = %v, want %v", got, subproto)
	}
}

func TestWebsocketSetReadDeadline(t *testing.T) {
	type args struct {
		wsConn WebsocketConn
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsConn: newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "conn.SetReadDeadline error",
			args: args{
				wsConn: newMockWebsocketConn().WithReadDlError(
					fmt.Errorf("Failed to set read deadline")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn).
				WithReadTimeout(time.Second)

			ws.setReadDeadline()

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketSetWriteDeadline(t *testing.T) {
	type args struct {
		wsConn WebsocketConn
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsConn: newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "conn.SetReadDeadline error",
			args: args{
				wsConn: newMockWebsocketConn().WithWriteDlError(
					fmt.Errorf("Failed to set read deadline")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn).
				WithWriteTimeout(time.Second)

			ws.setWriteDeadline()

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketDisconnect(t *testing.T) {
	type args struct {
		wsConn WebsocketConn
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "success",
			args: args{
				wsConn: newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "success even if conn is nil",
			args: args{
				wsConn: nil,
			},
			assertOk: true,
		},
		{
			name: "conn close failed",
			args: args{
				wsConn: newMockWebsocketConn().WithCloseError(
					fmt.Errorf("failed to close ws conn")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			chain := newChainWithDefaults("test", reporter)
			config := newMockConfig(reporter)

			ws := newWebsocket(chain, config, tt.args.wsConn)

			ws.Disconnect()

			if tt.assertOk {
				ws.chain.assertNotFailed(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketPrintRead(t *testing.T) {
	reporter := newMockReporter(t)
	printer := newMockWsPrinter()
	config := Config{
		Reporter: reporter,
		Printers: []Printer{printer},
	}.withDefaults()
	ws := newWebsocket(newMockChain(t), config, newMockWebsocketConn())

	ws.printRead(websocket.CloseMessage,
		[]byte("random message"),
		websocket.CloseNormalClosure)

	if !printer.isReadFrom {
		t.Errorf("Websocket.printRead() failed to read from printer")
	}
}

func TestWebsocketPrintWrite(t *testing.T) {
	reporter := newMockReporter(t)
	printer := newMockWsPrinter()
	config := Config{
		Reporter: reporter,
		Printers: []Printer{printer},
	}.withDefaults()
	ws := newWebsocket(newMockChain(t), config, newMockWebsocketConn())

	ws.printWrite(websocket.CloseMessage,
		[]byte("random message"),
		websocket.CloseNormalClosure)

	if !printer.isWrittenTo {
		t.Errorf("Websocket.printWrite() failed to write to printer")
	}
}
