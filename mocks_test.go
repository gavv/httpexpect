package httpexpect

import (
	"bytes"
	"io"
	"io/ioutil"
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

type mockBody struct {
	io.Reader
	closed bool
}

func newMockBody(body string) *mockBody {
	return &mockBody{
		Reader: bytes.NewBufferString(body),
		closed: false,
	}
}

func (b *mockBody) Close() error {
	b.closed = true
	return nil
}

func newMockChain(t *testing.T) *chain {
	return newDefaultChain("test", newMockReporter(t))
}

type mockLogger struct {
	testing *testing.T
	logged  bool
}

func newMockLogger(t *testing.T) *mockLogger {
	return &mockLogger{t, false}
}

func (r *mockLogger) Logf(message string, args ...interface{}) {
	r.testing.Logf(message, args...)
	r.logged = true
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

type mockFormatter struct {
	testing          *testing.T
	formattedSuccess int
	formattedFailure int
}

func newMockFormatter(t *testing.T) *mockFormatter {
	return &mockFormatter{testing: t}
}

func (m *mockFormatter) FormatSuccess(ctx *AssertionContext) string {
	m.formattedSuccess++
	return ctx.TestName
}

func (m *mockFormatter) FormatFailure(
	ctx *AssertionContext, failure *AssertionFailure,
) string {
	m.formattedFailure++
	return ctx.TestName
}

type mockPrinter struct {
	reqBody  []byte
	respBody []byte
	rtt      time.Duration
}

func (p *mockPrinter) Request(req *http.Request) {
	if req.Body != nil {
		p.reqBody, _ = ioutil.ReadAll(req.Body)
		req.Body.Close()
	}
}

func (p *mockPrinter) Response(resp *http.Response, rtt time.Duration) {
	if resp.Body != nil {
		p.respBody, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	p.rtt = rtt
}

type mockWebsocketPrinter struct {
	isWrittenTo bool
	isReadFrom  bool
}

func newMockWsPrinter() *mockWebsocketPrinter {
	return &mockWebsocketPrinter{
		isWrittenTo: false,
		isReadFrom:  false,
	}
}

func (p *mockWebsocketPrinter) Request(*http.Request) {
}

func (p *mockWebsocketPrinter) Response(*http.Response, time.Duration) {
}

func (p *mockWebsocketPrinter) WebsocketWrite(typ int, content []byte, closeCode int) {
	p.isWrittenTo = true
}

func (p *mockWebsocketPrinter) WebsocketRead(typ int, content []byte, closeCode int) {
	p.isReadFrom = true
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
