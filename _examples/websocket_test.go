package examples

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/gorilla/websocket"
)

func wsHandlerTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:         "ws://example.com",
		WebsocketDialer: httpexpect.NewWebsocketDialer(http.HandlerFunc(WsHTTPHandler)),
		Reporter:        httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestWsHandlerText(t *testing.T) {
	e := wsHandlerTester(t)

	ws := e.GET("/path").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.WriteText("hi").
		Expect().
		TextMessage().Body().IsEqual("hi")
}

func TestWsHandlerJSON(t *testing.T) {
	e := wsHandlerTester(t)

	ws := e.GET("/path").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.WriteJSON(struct {
		Message string `json:"message"`
	}{"hi"}).
		Expect().
		TextMessage().JSON().Object().ValueEqual("message", "hi")
}

func TestWsHandlerClose(t *testing.T) {
	e := wsHandlerTester(t)

	ws := e.GET("/path").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.CloseWithText("Namárië...", websocket.CloseGoingAway).
		Expect().
		CloseMessage().NoContent()
}

func wsFastHandlerTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:         "http://example.com",
		WebsocketDialer: httpexpect.NewFastWebsocketDialer(WsFastHTTPHandler),
		Reporter:        httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestWsFastHandlerText(t *testing.T) {
	e := wsFastHandlerTester(t)

	ws := e.GET("/path").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.WriteText("hi").
		Expect().
		TextMessage().Body().IsEqual("hi")
}

func TestWsFastHandlerJSON(t *testing.T) {
	e := wsFastHandlerTester(t)

	ws := e.GET("/path").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.WriteJSON(struct {
		Message string `json:"message"`
	}{"hi"}).
		Expect().
		TextMessage().JSON().Object().ValueEqual("message", "hi")
}

func TestWsFastHandlerClose(t *testing.T) {
	e := wsFastHandlerTester(t)

	ws := e.GET("/path").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.CloseWithText("Namárië...", websocket.CloseGoingAway).
		Expect().
		CloseMessage().NoContent()
}
