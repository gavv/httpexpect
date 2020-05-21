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
		r := newStringReporter()
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: true,
		}

		c.Expect()

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}

		want := "\nunexpected read from closed WebSocket connection"

		if got := r.msg; got != want {
			t.Errorf("Websocket.CloseWithText() = %v, want %v", got, want)
		}
	})
}

func TestWebsocketCheckUnusable(t *testing.T) {
	type fields struct {
		conn     *websocket.Conn
		isClosed bool
	}
	type args struct {
		reporter *stringReporter
		where    string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     bool
		reported bool
		wantMsg  string
	}{
		{
			name: "conn is nil",
			fields: fields{
				conn:     nil,
				isClosed: false,
			},
			args: args{
				reporter: newStringReporter(),
				where:    "Close",
			},
			want:     true,
			reported: true,
			wantMsg:  "\nunexpected Close call for failed WebSocket connection",
		},
		{
			name: "websocket is closed",
			fields: fields{
				conn:     &websocket.Conn{},
				isClosed: true,
			},
			args: args{
				reporter: newStringReporter(),
				where:    "Close",
			},
			want:     true,
			reported: true,
			wantMsg:  "\nunexpected Close call for closed WebSocket connection",
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
			if got := tt.args.reporter.msg; got != tt.wantMsg {
				t.Errorf("Websocket.checkUnusable() error message = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func TestWebsocketWriteJSONMarshalFail(t *testing.T) {
	r := newStringReporter()

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
		reporter  *stringReporter
		typ       int
		content   []byte
		closeCode []int
	}
	tests := []struct {
		name        string
		args        args
		reported    bool
		reportedMsg string
	}{
		{
			name: "Close message, multiple close code",
			args: args{
				reporter:  newStringReporter(),
				typ:       websocket.CloseMessage,
				content:   []byte("closing message..."),
				closeCode: []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			reported:    true,
			reportedMsg: "\nunexpected multiple closeCode arguments passed to WriteMessage",
		},
		{
			name: "Close message, multiple close code",
			args: args{
				reporter:  newStringReporter(),
				typ:       websocket.PingMessage,
				content:   []byte("closing message..."),
				closeCode: []int{websocket.CloseNormalClosure, websocket.CloseAbnormalClosure},
			},
			reported:    true,
			reportedMsg: "\nunexpected WebSocket message type 'ping' passed to WriteMessage",
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

			if got := tt.args.reporter.msg; got != tt.reportedMsg {
				t.Errorf("Websocket.WriteMessage() error message = %v, want %v", got, tt.reportedMsg)
			}
		})
	}
}

func TestWebsocketCloseWithText(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newStringReporter()
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

		want := "\nunexpected multiple code arguments passed to CloseWithText"

		if got := r.msg; got != want {
			t.Errorf("Websocket.CloseWithText() = %v, want %v", got, want)
		}
	})
}

func TestWebsocketCloseWithJSON(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newStringReporter()
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

		want := "\nunexpected multiple code arguments passed to CloseWithJSON"

		if got := r.msg; got != want {
			t.Errorf("Websocket.CloseWithText() = %v, want %v", got, want)
		}
	})

	t.Run("json marshall error", func(t *testing.T) {
		r := newStringReporter()
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
		r := newStringReporter()
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

		want := "\nunexpected multiple code arguments passed to CloseWithBytes"

		if got := r.msg; got != want {
			t.Errorf("Websocket.CloseWithText() = %v, want %v", got, want)
		}
	})
}

func TestWebsocketClose(t *testing.T) {
	t.Run("multiple code args", func(t *testing.T) {
		r := newStringReporter()
		c := &Websocket{
			chain:    makeChain(r),
			conn:     &websocket.Conn{},
			isClosed: false,
		}

		c.Close(websocket.CloseNormalClosure, websocket.CloseAbnormalClosure)

		if !r.reported {
			t.Errorf("Websocket.CloseWithText() error message not reported")
		}

		want := "\nunexpected multiple code arguments passed to Close"

		if got := r.msg; got != want {
			t.Errorf("Websocket.CloseWithText() = %v, want %v", got, want)
		}
	})
}
