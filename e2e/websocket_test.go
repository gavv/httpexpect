package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fastwebsocket "github.com/fasthttp/websocket"
	"github.com/gavv/httpexpect/v2"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type wsHandlerOpts struct {
	preRead  func()
	preWrite func()
}

func createWebsocketHandler(opts wsHandlerOpts) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
	})

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		upgrader := &websocket.Upgrader{}

		hdr := make(http.Header)
		hdr["X-Test"] = []string{"test_header"}

		c, err := upgrader.Upgrade(w, r, hdr)
		if err != nil {
			panic(err)
		}
		defer c.Close()

		for {
			if opts.preRead != nil {
				opts.preRead()
			}
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			if opts.preWrite != nil {
				opts.preWrite()
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	})

	return mux
}

func websocketFastHandler(ctx *fasthttp.RequestCtx) {
	var upgrader fastwebsocket.FastHTTPUpgrader

	ctx.Response.Header.Set("X-Test", "test_header")

	err := upgrader.Upgrade(ctx, func(c *fastwebsocket.Conn) {
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	})
	if err != nil {
		panic(err)
	}
}

func testWebsocketConn(e *httpexpect.Expect) {
	ws := e.GET("/test").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	if ws.Conn() == nil {
		panic("Conn returned nil")
	}
}

func testWebsocketHeader(e *httpexpect.Expect) {
	resp := e.GET("/test").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols)

	hdr := resp.Header("X-Test")
	hdr.IsEqual("test_header")

	ws := resp.Websocket()
	ws.Disconnect()
}

func testWebsocketSession(e *httpexpect.Expect) {
	ws := e.GET("/test").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.Subprotocol().IsEmpty()

	ws.WriteBytesBinary([]byte("my binary bytes")).
		Expect().
		BinaryMessage().Body().IsEqual("my binary bytes")

	ws.WriteBytesText([]byte("my text bytes")).
		Expect().
		TextMessage().Body().IsEqual("my text bytes")

	ws.WriteText("my text").
		Expect().
		TextMessage().Body().IsEqual("my text")

	ws.WriteJSON(struct {
		Message string `json:"message"`
	}{"my json"}).
		Expect().
		TextMessage().JSON().Object().HasValue("message", "my json")

	ws.CloseWithText("my close message").
		Expect().
		CloseMessage().NoContent()
}

func testWebsocketTypes(e *httpexpect.Expect) {
	ws := e.GET("/test").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.WriteMessage(websocket.TextMessage, []byte("test")).
		Expect().
		Type(websocket.TextMessage).Body().IsEqual("test")

	ws.WriteMessage(websocket.BinaryMessage, []byte("test")).
		Expect().
		Type(websocket.BinaryMessage).Body().IsEqual("test")

	ws.WriteMessage(websocket.CloseMessage, []byte("test")).
		Expect().
		Type(websocket.CloseMessage).NoContent()
}

func testWebsocket(e *httpexpect.Expect) {
	testWebsocketConn(e)
	testWebsocketHeader(e)
	testWebsocketSession(e)
	testWebsocketTypes(e)
}

func TestE2EWebsocket_Live(t *testing.T) {
	handler := createWebsocketHandler(wsHandlerOpts{})

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	testWebsocket(e)
}

func TestE2EWebsocket_HandlerStandard(t *testing.T) {
	t.Run("dialer-config", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		e := httpexpect.WithConfig(httpexpect.Config{
			Reporter:        httpexpect.NewAssertReporter(t),
			WebsocketDialer: httpexpect.NewWebsocketDialer(handler),
			Printers: []httpexpect.Printer{
				httpexpect.NewDebugPrinter(t, true),
			},
		})

		testWebsocket(e)
	})

	t.Run("dialer-method", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		e := httpexpect.WithConfig(httpexpect.Config{
			Reporter: httpexpect.NewAssertReporter(t),
			Printers: []httpexpect.Printer{
				httpexpect.NewDebugPrinter(t, true),
			},
		})

		testWebsocket(e.Builder(func(req *httpexpect.Request) {
			req.WithWebsocketDialer(httpexpect.NewWebsocketDialer(handler))
		}))
	})
}

func TestE2EWebsocket_HandlerFast(t *testing.T) {
	t.Run("dialer-config", func(t *testing.T) {
		e := httpexpect.WithConfig(httpexpect.Config{
			Reporter:        httpexpect.NewAssertReporter(t),
			WebsocketDialer: httpexpect.NewFastWebsocketDialer(websocketFastHandler),
			Printers: []httpexpect.Printer{
				httpexpect.NewDebugPrinter(t, true),
			},
		})

		testWebsocket(e)
	})

	t.Run("dialer-method", func(t *testing.T) {
		e := httpexpect.WithConfig(httpexpect.Config{
			Reporter: httpexpect.NewAssertReporter(t),
			Printers: []httpexpect.Printer{
				httpexpect.NewDebugPrinter(t, true),
			},
		})

		testWebsocket(e.Builder(func(req *httpexpect.Request) {
			req.WithWebsocketDialer(httpexpect.NewFastWebsocketDialer(websocketFastHandler))
		}))
	})
}

func testWebsocketTimeout(
	t *testing.T,
	handler http.Handler,
	blockCh chan struct{},
	timeout bool,
	setupFn func(*httpexpect.Websocket),
) {
	server := httptest.NewServer(handler)
	defer server.Close()

	reporter := &mockReporter{}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: reporter,
	})

	ws := e.GET("/test").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	setupFn(ws)

	blockCh <- struct{}{}

	ws.WriteText("test").Expect()

	assert.False(t, reporter.failed)

	go func() {
		time.Sleep(time.Millisecond * 100)
		blockCh <- struct{}{}
	}()

	ws.WriteText("test").Expect()
	if timeout {
		assert.True(t, reporter.failed)
	} else {
		assert.False(t, reporter.failed)
	}
}

func TestE2EWebsocket_Timeouts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("with-read-timeout", func(t *testing.T) {
		blockCh := make(chan struct{}, 1)

		handler := createWebsocketHandler(wsHandlerOpts{
			preWrite: func() {
				<-blockCh
			},
		})

		testWebsocketTimeout(t, handler, blockCh, true, func(ws *httpexpect.Websocket) {
			ws.WithReadTimeout(time.Millisecond * 10)
		})
	})

	t.Run("without-read-timeout", func(t *testing.T) {
		blockCh := make(chan struct{}, 1)

		handler := createWebsocketHandler(wsHandlerOpts{
			preWrite: func() {
				<-blockCh
			},
		})

		testWebsocketTimeout(t, handler, blockCh, false, func(ws *httpexpect.Websocket) {
			ws.WithoutReadTimeout()
		})
	})

	t.Run("without-write-timeout", func(t *testing.T) {
		blockCh := make(chan struct{}, 1)

		handler := createWebsocketHandler(wsHandlerOpts{
			preRead: func() {
				<-blockCh
			},
		})

		testWebsocketTimeout(t, handler, blockCh, false, func(ws *httpexpect.Websocket) {
			ws.WithoutWriteTimeout()
		})
	})
}

func TestE2EWebsocket_Closed(t *testing.T) {
	t.Run("close-write", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		server := httptest.NewServer(handler)
		defer server.Close()

		reporter := &mockReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		ws := e.GET("/test").WithWebsocketUpgrade().
			Expect().
			Status(http.StatusSwitchingProtocols).
			Websocket()
		defer ws.Disconnect()

		ws.CloseWithText("bye")
		assert.False(t, reporter.failed)

		ws.WriteText("test")
		assert.True(t, reporter.failed)
	})

	t.Run("close-close", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		server := httptest.NewServer(handler)
		defer server.Close()

		reporter := &mockReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		ws := e.GET("/test").WithWebsocketUpgrade().
			Expect().
			Status(http.StatusSwitchingProtocols).
			Websocket()
		defer ws.Disconnect()

		ws.CloseWithText("bye")
		assert.False(t, reporter.failed)

		ws.CloseWithText("bye")
		assert.True(t, reporter.failed)
	})
}

func TestE2EWebsocket_Disconnected(t *testing.T) {
	t.Run("disconnect-write", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		server := httptest.NewServer(handler)
		defer server.Close()

		reporter := &mockReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		ws := e.GET("/test").WithWebsocketUpgrade().
			Expect().
			Status(http.StatusSwitchingProtocols).
			Websocket()

		ws.Disconnect()
		assert.False(t, reporter.failed)

		ws.WriteText("test")
		assert.True(t, reporter.failed)
	})

	t.Run("disconnect-close", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		server := httptest.NewServer(handler)
		defer server.Close()

		reporter := &mockReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		ws := e.GET("/test").WithWebsocketUpgrade().
			Expect().
			Status(http.StatusSwitchingProtocols).
			Websocket()

		ws.Disconnect()
		assert.False(t, reporter.failed)

		ws.CloseWithText("test")
		assert.True(t, reporter.failed)
	})

	t.Run("disconnect-disconnect", func(t *testing.T) {
		handler := createWebsocketHandler(wsHandlerOpts{})

		server := httptest.NewServer(handler)
		defer server.Close()

		reporter := &mockReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		ws := e.GET("/test").WithWebsocketUpgrade().
			Expect().
			Status(http.StatusSwitchingProtocols).
			Websocket()

		ws.Disconnect()
		assert.False(t, reporter.failed)

		ws.Disconnect()
		assert.False(t, reporter.failed)
	})
}

func TestE2EWebsocket_Invalid(t *testing.T) {
	handler := createWebsocketHandler(wsHandlerOpts{})

	server := httptest.NewServer(handler)
	defer server.Close()

	reporter := &mockReporter{}

	t.Run("no_upgrade_on_client", func(t *testing.T) {
		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		// missing WithWebsocketUpgrade()
		resp := e.GET("/empty").
			Expect().
			Status(http.StatusOK)

		ws := resp.Websocket()
		defer ws.Disconnect()

		assert.True(t, reporter.failed)
	})

	t.Run("no_upgrade_on_server", func(t *testing.T) {
		reporter := &mockReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: reporter,
		})

		e.GET("/empty").WithWebsocketUpgrade().
			Expect()

		assert.True(t, reporter.failed)
	})
}
