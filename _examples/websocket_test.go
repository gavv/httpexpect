package examples

import (
	"testing"
	"net/http"

	"github.com/gavv/httpexpect"
	"git.instrumentisto.com/streaming/api/_cache/dep/sources/https---github.com-gorilla-websocket"
)

func wsTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL: "http://example.com",
		Dialer: httpexpect.NewBinder(http.HandlerFunc(WsHandler)).Dialer(),
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestWsHandler_Text(t *testing.T) {
	conn := wsTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteText("hi").
		Expect().Text().Body().Equal("hi")
}

func TestWsHandler_JSON(t *testing.T) {
	conn := wsTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteJSON(struct {
		Message string `json:"message"`
	}{"hi"}).
		Expect().
		Text().JSON().Object().ValueEqual("message", "hi")
}

func TestWsHandler_Close(t *testing.T) {
	conn := wsTester(t).WS("/ws").Expect().
		Status(http.StatusSwitchingProtocols).
		Connection()
	defer conn.Disconnect()

	conn.WriteMessage(websocket.CloseMessage,
		[]byte("Namárië..."), websocket.CloseGoingAway)
	conn.Expect().
		Closed().
		NoContent()
}
