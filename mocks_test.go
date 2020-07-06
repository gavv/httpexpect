package httpexpect

import (
	"net/http"
	"testing"
	"time"
)

type mockClient struct {
	req  *http.Request
	resp http.Response
	err  error
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req
	if c.err == nil {
		c.resp.Header = c.req.Header
		c.resp.Body = c.req.Body
		return &c.resp, nil
	}
	return nil, c.err
}

type mockReporter struct {
	testing  *testing.T
	reported bool
}

func newMockReporter(t *testing.T) *mockReporter {
	return &mockReporter{t, false}
}

func (r *mockReporter) Errorf(message string, args ...interface{}) {
	r.testing.Logf("Fail: "+message, args...)
	r.reported = true
}

type mockWebsocketConn struct {
	msgType      int
	readMsgErr   error
	writeMsgErr  error
	closeError   error
	readDlError  error
	writeDlError error
	msg          []byte
	subprotocol  string
}

func newMockWebsocketConn() *mockWebsocketConn {
	return &mockWebsocketConn{}
}

func (wc *mockWebsocketConn) WithWriteMsgError(retError error) *mockWebsocketConn {
	wc.writeMsgErr = retError
	return wc
}

func (wc *mockWebsocketConn) WithReadMsgError(retError error) *mockWebsocketConn {
	wc.readMsgErr = retError
	return wc
}

func (wc *mockWebsocketConn) WithWriteDlError(retError error) *mockWebsocketConn {
	wc.writeDlError = retError
	return wc
}

func (wc *mockWebsocketConn) WithReadDlError(retError error) *mockWebsocketConn {
	wc.readDlError = retError
	return wc
}

func (wc *mockWebsocketConn) WithCloseError(retError error) *mockWebsocketConn {
	wc.closeError = retError
	return wc
}

func (wc *mockWebsocketConn) WithMsgType(msgType int) *mockWebsocketConn {
	wc.msgType = msgType
	return wc
}

func (wc *mockWebsocketConn) WithSubprotocol(subprotocol string) *mockWebsocketConn {
	wc.subprotocol = subprotocol
	return wc
}

func (wc *mockWebsocketConn) WithMessage(msg []byte) *mockWebsocketConn {
	wc.msg = msg
	return wc
}

func (wc *mockWebsocketConn) ReadMessage() (messageType int, p []byte, err error) {
	return wc.msgType, []byte{}, wc.readMsgErr
}
func (wc *mockWebsocketConn) WriteMessage(messageType int, data []byte) error {
	return wc.writeMsgErr

}
func (wc *mockWebsocketConn) Close() error {
	return wc.closeError
}
func (wc *mockWebsocketConn) SetReadDeadline(t time.Time) error {
	return wc.readDlError
}
func (wc *mockWebsocketConn) SetWriteDeadline(t time.Time) error {
	return wc.writeDlError
}
func (wc *mockWebsocketConn) Subprotocol() string {
	return wc.subprotocol
}

type MockWsPrinter struct {
	isWrittenTo bool
	isReadFrom  bool
}

func newMockWsPrinter() *MockWsPrinter {
	return &MockWsPrinter{
		isWrittenTo: false,
		isReadFrom:  false,
	}
}

func (pr *MockWsPrinter) Request(*http.Request) {}

// Response is called after response is received.
func (pr *MockWsPrinter) Response(*http.Response, time.Duration) {}

func (pr *MockWsPrinter) WebsocketWrite(typ int, content []byte, closeCode int) {
	pr.isWrittenTo = true
}

func (pr *MockWsPrinter) WebsocketRead(typ int, content []byte, closeCode int) {
	pr.isReadFrom = true
}
