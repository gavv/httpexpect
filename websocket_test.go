package httpexpect

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func noWsPreSteps(ws *Websocket) {}

func TestWebsocketFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	ws := makeWebsocket(Config{}, chain, nil)

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

func TestWebsocketExpect(t *testing.T) {

	failedChain := makeChain(newMockReporter(t))
	failedChain.fail("some previous fail...")

	type args struct {
		config     Config
		chain      chain
		wsConn     WebsocketConn
		wsPreSteps func(*Websocket)
	}
	tests := []struct {
		name     string
		args     args
		assertOk bool
	}{
		{
			name: "Expect success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "Expect failed to read message from conn",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn().WithReadMsgError(fmt.Errorf("failed to read message")),
			},
			assertOk: false,
		},
		{
			name: "Expect chain already failed",
			args: args{
				config:     Config{},
				chain:      failedChain,
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
			},
			assertOk: false,
		},
		{
			name: "Expect connection closed",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn: newMockWebsocketConn(),
			},
			assertOk: false,
		},
		{
			name: "Expect failed to set read deadline",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn().WithReadDlError(fmt.Errorf("failed to set read deadline")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.Expect()

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketClose(t *testing.T) {

	failedChain := makeChain(newMockReporter(t))
	failedChain.fail("some previous fail...")

	type args struct {
		config     Config
		chain      chain
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
			name: "Close success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "Close websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsPreSteps: func(ws *Websocket) {
					ws.Disconnect()
				},
				wsConn:    newMockWebsocketConn(),
				closeCode: []int{websocket.CloseNormalClosure},
			},
			assertOk: false,
		},
		{
			name: "Close too many close codes",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.Close(tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketCloseWithBytes(t *testing.T) {

	failedChain := makeChain(newMockReporter(t))
	failedChain.fail("some previous fail...")

	type args struct {
		config     Config
		chain      chain
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("connection closed..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.CloseWithBytes(tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketCloseWithText(t *testing.T) {

	failedChain := makeChain(newMockReporter(t))
	failedChain.fail("some previous fail...")

	type args struct {
		config     Config
		chain      chain
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    "connection closed...",
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.CloseWithText(tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketCloseWithJSON(t *testing.T) {

	failedChain := makeChain(newMockReporter(t))
	failedChain.fail("some previous fail...")

	type args struct {
		config     Config
		chain      chain
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
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
			name: "websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
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
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.CloseWithJSON(tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteMessage(t *testing.T) {
	type args struct {
		config     Config
		chain      chain
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
			name: "Text message success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: true,
		},
		{
			name: "Text message fail unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
			name: "Text message failed to set write deadline",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsConn:     newMockWebsocketConn().WithWriteDlError(fmt.Errorf("failed to set write deadline")),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: false,
		},
		{
			name: "Text message failed to write to conn",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsConn:     newMockWebsocketConn().WithWriteMsgError(fmt.Errorf("failed to write message to conn")),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.TextMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: false,
		},
		{
			name: "Text binary message success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				typ:        websocket.BinaryMessage,
				content:    []byte("random message..."),
				closeCode:  []int{},
			},
			assertOk: true,
		},
		{
			name: "Close message success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMessage,
				content:    []byte("closing message..."),
				closeCode:  []int{websocket.CloseNormalClosure},
			},
			assertOk: true,
		},
		{
			name: "Close message too many close codes",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				typ:        websocket.CloseMessage,
				content:    []byte("closing message..."),
				closeCode:  []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			assertOk: false,
		},
		{
			name: "Unsupported message type",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteMessage(tt.args.typ, tt.args.content, tt.args.closeCode...)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteBytesBinary(t *testing.T) {
	type args struct {
		config     Config
		chain      chain
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
			name: "Write binary success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("random message..."),
			},
			assertOk: true,
		},
		{
			name: "Write binary websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteBytesBinary(tt.args.content)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteBytesText(t *testing.T) {
	type args struct {
		config     Config
		chain      chain
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
			name: "Write bytes text success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    []byte("random message..."),
			},
			assertOk: true,
		},
		{
			name: "Write bytes text websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteBytesText(tt.args.content)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteText(t *testing.T) {
	type args struct {
		config     Config
		chain      chain
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
			name: "Write text success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content:    "random message...",
			},
			assertOk: true,
		},
		{
			name: "Write text websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteText(tt.args.content)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocketWriteJSON(t *testing.T) {
	type args struct {
		config     Config
		chain      chain
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
			name: "Write JSON success",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsPreSteps: noWsPreSteps,
				wsConn:     newMockWebsocketConn(),
				content: map[string]string{
					"msg": "random message",
				},
			},
			assertOk: true,
		},
		{
			name: "Write JSON websocket unusable",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
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
			name: "Write JSON marshal failed",
			args: args{
				config:     Config{},
				chain:      makeChain(newMockReporter(t)),
				wsConn:     newMockWebsocketConn(),
				wsPreSteps: noWsPreSteps,
				content:    make(chan int),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			tt.args.wsPreSteps(ws)

			ws.WriteJSON(tt.args.content)

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocket_Subprotocol(t *testing.T) {
	subproto := "soap"
	ws := NewWebsocket(Config{}, newMockWebsocketConn().WithSubprotocol(subproto))

	ws.Subprotocol()

	if got := ws.Subprotocol().value; got != subproto {
		t.Errorf("Websocket.Subprotocol() = %v, want %v", got, subproto)
	}
}

func TestWebsocket_setReadDeadline(t *testing.T) {
	type args struct {
		config Config
		chain  chain
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
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsConn: newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "conn.SetReadDeadline error",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsConn: newMockWebsocketConn().WithReadDlError(fmt.Errorf("Failed to set read deadline")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn).WithReadTimeout(time.Second)

			ws.setReadDeadline()

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocket_setWriteDeadline(t *testing.T) {
	type args struct {
		config Config
		chain  chain
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
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsConn: newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "conn.SetReadDeadline error",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsConn: newMockWebsocketConn().WithWriteDlError(fmt.Errorf("Failed to set read deadline")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn).WithWriteTimeout(time.Second)

			ws.setWriteDeadline()

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestWebsocket_Disconnect(t *testing.T) {
	type args struct {
		config Config
		chain  chain
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
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsConn: newMockWebsocketConn(),
			},
			assertOk: true,
		},
		{
			name: "conn close failed",
			args: args{
				config: Config{},
				chain:  makeChain(newMockReporter(t)),
				wsConn: newMockWebsocketConn().WithCloseError(fmt.Errorf("failed to close ws conn")),
			},
			assertOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := makeWebsocket(tt.args.config, tt.args.chain, tt.args.wsConn)

			ws.Disconnect()

			if tt.assertOk {
				ws.chain.assertOK(t)
			} else {
				ws.chain.assertFailed(t)
			}
		})
	}
}

func TestPrintRead(t *testing.T) {
	printer := newMockWsPrinter()
	config := Config{
		Printers: []Printer{printer},
	}
	ws := makeWebsocket(config, makeChain(newMockReporter(t)), newMockWebsocketConn())

	ws.printRead(websocket.CloseMessage, []byte("random message"), websocket.CloseNormalClosure)

	if !printer.isReadFrom {
		t.Errorf("Websocket.printRead() failed to read from printer")
	}
}

func TestPrintWrite(t *testing.T) {
	printer := newMockWsPrinter()
	config := Config{
		Printers: []Printer{printer},
	}
	ws := makeWebsocket(config, makeChain(newMockReporter(t)), newMockWebsocketConn())

	ws.printWrite(websocket.CloseMessage, []byte("random message"), websocket.CloseNormalClosure)

	if !printer.isWrittenTo {
		t.Errorf("Websocket.printWrite() failed to write to printer")
	}
}
