package httpexpect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	req  *http.Request
	resp http.Response
	err  error
	cb   func(req *http.Request) // callback in .Do
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	defer func() {
		if c.cb != nil {
			c.cb(req)
		}
	}()
	c.req = req
	if c.err == nil {
		c.resp.Header = c.req.Header
		c.resp.Body = c.req.Body
		return &c.resp, nil
	}
	return nil, c.err
}

// mockTransportRedirect mocks a transport that implements RoundTripper
//
// When tripCount < maxRedirect,
// mockTransportRedirect responses with redirectHTTPStatusCode
//
// When tripCount = maxRedirect,
// mockTransportRedirect responses with HTTP 200 OK
type mockTransportRedirect struct {
	t *testing.T

	// assertFn asserts the HTTP request
	assertFn func(*http.Request) bool

	// redirectHTTPStatusCode indicates the HTTP status code of redirection response
	redirectHTTPStatusCode int

	// tripCount tracks the number of trip that has been done
	tripCount int

	// maxRedirect indicates the number of trip that can be done for redirection.
	// -1 means always redirect.
	maxRedirect int
}

func newMockTransportRedirect(
	t *testing.T,
) *mockTransportRedirect {
	return &mockTransportRedirect{
		t:                      t,
		assertFn:               nil,
		redirectHTTPStatusCode: http.StatusPermanentRedirect,
		maxRedirect:            -1,
	}
}

func (mt *mockTransportRedirect) RoundTrip(origReq *http.Request) (
	*http.Response, error,
) {
	mt.tripCount++

	if mt.assertFn != nil {
		assert.True(mt.t, mt.assertFn(origReq))
	}

	res := httptest.NewRecorder()

	if mt.maxRedirect == -1 || mt.tripCount <= mt.maxRedirect {
		res.Result().StatusCode = mt.redirectHTTPStatusCode
		res.Result().Header.Set("Location", "/redirect")
	} else {
		res.Result().StatusCode = http.StatusOK
	}

	return res.Result(), nil
}

func (mt *mockTransportRedirect) WithAssertFn(
	fn func(*http.Request) bool,
) *mockTransportRedirect {
	mt.assertFn = fn

	return mt
}

func (mt *mockTransportRedirect) WithRedirectHTTPStatusCode(
	statusCode int,
) *mockTransportRedirect {
	if !(statusCode >= 300 && statusCode < 400) {
		mt.t.Fatal("invalid redirect status code")
	}

	mt.redirectHTTPStatusCode = statusCode

	return mt
}

func (mt *mockTransportRedirect) WithMaxRedirect(
	maxRedirect int,
) *mockTransportRedirect {
	if maxRedirect != -1 && maxRedirect < 0 {
		mt.t.Fatal("max redirect less than 0")
	}

	mt.maxRedirect = maxRedirect

	return mt
}

type mockBody struct {
	reader      io.Reader
	closed      bool
	readErr     error
	closeErr    error
	hasBeenRead bool
}

func newMockBody(body string) *mockBody {
	return &mockBody{
		reader: bytes.NewBufferString(body),
	}
}

func (b *mockBody) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	b.hasBeenRead = true
	return b.reader.Read(p)
}

func (b *mockBody) Close() error {
	b.closed = true
	if b.closeErr != nil {
		return b.closeErr
	}
	return nil
}

func newMockConfig(r Reporter) Config {
	return Config{Reporter: r}.withDefaults()
}

func newMockChain(t *testing.T) *chain {
	return newChainWithDefaults("test", newMockReporter(t))
}

type mockLogger struct {
	testing     *testing.T
	logged      bool
	lastMessage string
}

func newMockLogger(t *testing.T) *mockLogger {
	return &mockLogger{testing: t}
}

func (l *mockLogger) Logf(message string, args ...interface{}) {
	l.testing.Logf(message, args...)
	l.lastMessage = fmt.Sprintf(message, args...)
	l.logged = true
}

type mockReporter struct {
	testing  *testing.T
	reported bool
}

func newMockReporter(t *testing.T) *mockReporter {
	return &mockReporter{testing: t}
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

func (f *mockFormatter) FormatSuccess(ctx *AssertionContext) string {
	f.formattedSuccess++
	return ctx.TestName
}

func (f *mockFormatter) FormatFailure(
	ctx *AssertionContext, failure *AssertionFailure,
) string {
	f.formattedFailure++
	return ctx.TestName
}

type mockAssertionHandler struct {
	ctx     *AssertionContext
	failure *AssertionFailure
}

func (h *mockAssertionHandler) Success(ctx *AssertionContext) {
	h.ctx = ctx
}

func (h *mockAssertionHandler) Failure(
	ctx *AssertionContext, failure *AssertionFailure,
) {
	h.ctx = ctx
	h.failure = failure
}

func mockFailure() AssertionFailure {
	return AssertionFailure{
		Errors: []error{
			errors.New("test_error"),
		},
	}
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

type mockNetError struct {
	isTimeout   bool
	isTemporary bool
}

func (e *mockNetError) Error() string {
	return "mock net error"
}

func (e *mockNetError) Timeout() bool {
	return e.isTimeout
}

func (e *mockNetError) Temporary() bool {
	return e.isTemporary
}
