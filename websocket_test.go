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

func TestWebsocketExpect(t *testing.T) {
	t.Run("websocket is closed", func(t *testing.T) {
		r := newMockReporter(t)
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: true,
		}

		c.Expect()

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}
	})
}

func TestWebsocketCheckUnusable(t *testing.T) {
	type fields struct {
		conn     *websocket.Conn
		isClosed bool
	}
	type args struct {
		reporter *mockReporter
		where    string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     bool
		reported bool
	}{
		{
			name: "conn is nil",
			fields: fields{
				conn:     nil,
				isClosed: false,
			},
			args: args{
				reporter: newMockReporter(t),
				where:    "Close",
			},
			want:     true,
			reported: true,
		},
		{
			name: "websocket is closed",
			fields: fields{
				conn:     &websocket.Conn{},
				isClosed: true,
			},
			args: args{
				reporter: newMockReporter(t),
				where:    "Close",
			},
			want:     true,
			reported: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Websocket{
				chain:    makeChain(tt.args.reporter),
				conn:     tt.fields.conn,
				isClosed: tt.fields.isClosed,
			}
			if got := c.checkUnusable(tt.args.where); got != tt.want {
				t.Errorf("Websocket.checkUnusable() = %v, want %v", got, tt.want)
				return
			}
			if got := tt.args.reporter.reported; got != tt.reported {
				t.Errorf("Websocket.checkUnusable() is error message reported = %v, want %v",
					got, tt.want)
				return
			}
		})
	}
}

func TestWebsocketWriteJSONMarshalFail(t *testing.T) {
	r := newMockReporter(t)

	ws := &Websocket{
		chain:    makeChain(r),
		conn:     &websocket.Conn{},
		isClosed: false,
	}

	channel := make(chan int)

	ws.WriteJSON(channel)

	ws.chain.assertFailed(t)

	if !r.reported {
		t.Errorf("Error message not reported")
		return
	}
}

func TestWebsocketWriteMessage(t *testing.T) {
	type args struct {
		reporter  *mockReporter
		typ       int
		content   []byte
		closeCode []int
	}
	tests := []struct {
		name     string
		args     args
		reported bool
	}{
		{
			name: "Close message, multiple close code",
			args: args{
				reporter:  newMockReporter(t),
				typ:       websocket.CloseMessage,
				content:   []byte("closing message..."),
				closeCode: []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			reported: true,
		},
		{
			name: "Close message, multiple close code",
			args: args{
				reporter:  newMockReporter(t),
				typ:       websocket.PingMessage,
				content:   []byte("closing message..."),
				closeCode: []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			reported: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Websocket{
				chain:    makeChain(tt.args.reporter),
				conn:     &websocket.Conn{},
				isClosed: false,
			}

			c.WriteMessage(tt.args.typ, tt.args.content, tt.args.closeCode...)

			if got := tt.args.reporter.reported; got != tt.reported {
				t.Errorf("Websocket.WriteMessage() is error message reported = %v, want %v",
					got, tt.reported)
			}
		})
	}
}

func TestWebsocketCloseWithText(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newMockReporter(t)
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: false,
		}

		c.CloseWithText("Closing...", websocket.CloseNormalClosure,
			websocket.CloseAbnormalClosure)

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}
	})
}

func TestWebsocketCloseWithJSON(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newMockReporter(t)
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: false,
		}

		c.CloseWithJSON("Closing...", websocket.CloseNormalClosure,
			websocket.CloseAbnormalClosure)

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}
	})

	t.Run("json marshall error", func(t *testing.T) {
		r := newMockReporter(t)
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: false,
		}

		c.CloseWithJSON(make(chan int), websocket.CloseAbnormalClosure)

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}
	})
}

func TestWebsocketCloseWithBytes(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newMockReporter(t)
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: false,
		}

		c.CloseWithBytes([]byte("Closing..."), websocket.CloseNormalClosure,
			websocket.CloseAbnormalClosure)

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}
	})
}

func TestWebsocketClose(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newMockReporter(t)
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: false,
		}

		c.Close(websocket.CloseNormalClosure, websocket.CloseAbnormalClosure)

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}
	})
}
