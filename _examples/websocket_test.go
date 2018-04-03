package examples

import (
	"testing"
	"net/http"

	"github.com/gavv/httpexpect"
	"github.com/gorilla/websocket"
)

func wsHttpHandlerTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL: "http://example.com",
		Dialer: httpexpect.NewBinder(http.HandlerFunc(WsHttpHandler)).Dialer(),
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestWsHttpHandler_Text(t *testing.T) {
	conn := wsHttpHandlerTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteText("hi").
		Expect().Text().Body().Equal("hi")
}

func TestWsHttpHandler_JSON(t *testing.T) {
	conn := wsHttpHandlerTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteJSON(struct {
		Message string `json:"message"`
	}{"hi"}).
		Expect().
		Text().JSON().Object().ValueEqual("message", "hi")
}

func TestWsHttpHandler_Close(t *testing.T) {
	conn := wsHttpHandlerTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteMessage(websocket.CloseMessage,
		[]byte("Namárië..."), websocket.CloseGoingAway)
	conn.Expect().
		Closed().
		NoContent()
}

func wsFastHandlerTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL: "http://example.com",
		Dialer: httpexpect.NewFastBinder(WsFastHandler).Dialer(),
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestWsFastHandler_Text(t *testing.T) {
	conn := wsFastHandlerTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteText("hi").
		Expect().Text().Body().Equal("hi")
}

func TestWsFastHandler_JSON(t *testing.T) {
	conn := wsFastHandlerTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteJSON(struct {
		Message string `json:"message"`
	}{"hi"}).
		Expect().
		Text().JSON().Object().ValueEqual("message", "hi")
}

func TestWsFastHandler_Close(t *testing.T) {
	conn := wsFastHandlerTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteMessage(websocket.CloseMessage,
		[]byte("Namárië..."), websocket.CloseGoingAway)
	conn.Expect().
		Closed().
		NoContent()
}
