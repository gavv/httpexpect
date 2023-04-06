package httpexpect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// mock config
func newMockConfig(r Reporter) Config {
	return Config{Reporter: r}.withDefaults()
}

// mock chain
func newMockChain(t *testing.T, flag ...chainFlags) *chain {
	return newChainWithDefaults("test", newMockReporter(t), flag...)
}

// mock logger
type mockLogger struct {
	testing     *testing.T
	logged      bool
	lastMessage string
}

func newMockLogger(t *testing.T) *mockLogger {
	return &mockLogger{testing: t}
}

func (ml *mockLogger) Logf(message string, args ...interface{}) {
	ml.testing.Logf(message, args...)
	ml.lastMessage = fmt.Sprintf(message, args...)
	ml.logged = true
}

// mock reporter
type mockReporter struct {
	testing  *testing.T
	reported bool
	reportCb func()
}

func newMockReporter(t *testing.T) *mockReporter {
	return &mockReporter{testing: t}
}

func (mr *mockReporter) Errorf(message string, args ...interface{}) {
	mr.testing.Logf("Fail: "+message, args...)
	mr.reported = true

	if mr.reportCb != nil {
		mr.reportCb()
	}
}

// mock formatter
type mockFormatter struct {
	testing          *testing.T
	formattedSuccess int
	formattedFailure int
}

func newMockFormatter(t *testing.T) *mockFormatter {
	return &mockFormatter{testing: t}
}

func (mf *mockFormatter) FormatSuccess(ctx *AssertionContext) string {
	mf.formattedSuccess++
	return ctx.TestName
}

func (mf *mockFormatter) FormatFailure(
	ctx *AssertionContext, failure *AssertionFailure,
) string {
	mf.formattedFailure++
	return ctx.TestName
}

// mock assertion handler
type mockAssertionHandler struct {
	ctx     *AssertionContext
	failure *AssertionFailure
}

func (mh *mockAssertionHandler) Success(ctx *AssertionContext) {
	mh.ctx = ctx
}

func (mh *mockAssertionHandler) Failure(
	ctx *AssertionContext, failure *AssertionFailure,
) {
	mh.ctx = ctx
	mh.failure = failure
}

// mock websocket printer
type mockWebsocketPrinter struct {
	isWrittenTo bool
	isReadFrom  bool
}

func (mp *mockWebsocketPrinter) Request(*http.Request) {
}

func (mp *mockWebsocketPrinter) Response(*http.Response, time.Duration) {
}

func (mp *mockWebsocketPrinter) WebsocketWrite(typ int, content []byte, closeCode int) {
	mp.isWrittenTo = true
}

func (mp *mockWebsocketPrinter) WebsocketRead(typ int, content []byte, closeCode int) {
	mp.isReadFrom = true
}

// mock websocket connection
type mockWebsocketConn struct {
	subprotocol  string
	closeError   error
	readMsgErr   error
	writeMsgErr  error
	readDlError  error
	writeDlError error
	msgType      int
	msg          []byte
}

func (mc *mockWebsocketConn) Subprotocol() string {
	return mc.subprotocol
}

func (mc *mockWebsocketConn) Close() error {
	return mc.closeError
}

func (mc *mockWebsocketConn) SetReadDeadline(t time.Time) error {
	return mc.readDlError
}

func (mc *mockWebsocketConn) SetWriteDeadline(t time.Time) error {
	return mc.writeDlError
}

func (mc *mockWebsocketConn) ReadMessage() (messageType int, p []byte, err error) {
	return mc.msgType, []byte{}, mc.readMsgErr
}

func (mc *mockWebsocketConn) WriteMessage(messageType int, data []byte) error {
	return mc.writeMsgErr
}

// mock http client
type mockClient struct {
	req  *http.Request
	resp http.Response
	err  error
	cb   func(req *http.Request)
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

// mock http redirecting transport
type mockRedirectTransport struct {
	// assertFn asserts the HTTP request
	assertFn func(*http.Request)

	// redirectHTTPStatusCode indicates the HTTP status code of redirection response
	redirectHTTPStatusCode int

	// tripCount tracks the number of trip that has been done
	//
	// When tripCount < maxRedirect,
	// mockTransportRedirect responses with redirectHTTPStatusCode
	//
	// When tripCount = maxRedirect,
	// mockTransportRedirect responses with HTTP 200 OK
	tripCount int

	// maxRedirects indicates the number of trip that can be done for redirection.
	// -1 means always redirect.
	maxRedirects int
}

func newMockRedirectTransport() *mockRedirectTransport {
	return &mockRedirectTransport{
		redirectHTTPStatusCode: http.StatusPermanentRedirect,
		maxRedirects:           -1,
	}
}

func (mt *mockRedirectTransport) RoundTrip(origReq *http.Request) (
	*http.Response, error,
) {
	mt.tripCount++

	if mt.assertFn != nil {
		mt.assertFn(origReq)
	}

	res := httptest.NewRecorder()

	if mt.maxRedirects == -1 || mt.tripCount <= mt.maxRedirects {
		res.Result().StatusCode = mt.redirectHTTPStatusCode
		res.Result().Header.Set("Location", "/redirect")
	} else {
		res.Result().StatusCode = http.StatusOK
	}

	return res.Result(), nil
}

// mock http request factory
type mockRequestFactory struct {
	lastreq *http.Request
	fail    bool
}

func (mf *mockRequestFactory) NewRequest(
	method, urlStr string, body io.Reader) (*http.Request, error) {
	if mf.fail {
		return nil, errors.New("testRequestFactory")
	}
	mf.lastreq = httptest.NewRequest(method, urlStr, body)
	return mf.lastreq, nil
}

// mock http request or response body
type mockBody struct {
	reader io.Reader

	readCount int
	readErr   error

	closeCount int
	closeErr   error

	errCount int
	eofCount int
}

func newMockBody(body string) *mockBody {
	return &mockBody{
		reader: bytes.NewBufferString(body),
	}
}

func (mb *mockBody) Read(p []byte) (int, error) {
	mb.readCount++

	if mb.readErr != nil {
		return 0, mb.readErr
	}

	n, err := mb.reader.Read(p)
	if err == io.EOF {
		mb.eofCount++
	} else if err != nil {
		mb.errCount++
	}

	return n, err
}

func (mb *mockBody) Close() error {
	mb.closeCount++

	if mb.closeErr != nil {
		return mb.closeErr
	}

	return nil
}

// mock query string encoder (query.Encoder.EncodeValues)
type mockQueryEncoder string

func (mq mockQueryEncoder) EncodeValues(key string, v *url.Values) error {
	if mq == "err" {
		return errors.New("encoding error")
	}
	v.Set(key, string(mq))
	return nil
}

// mock io.Writer
type mockWriter struct {
	io.Writer
	err error
}

func (mw *mockWriter) Write(p []byte) (n int, err error) {
	if mw.err != nil {
		return 0, err
	}

	return mw.Writer.Write(p)
}

// mock network error
type mockNetError struct {
	isTimeout   bool
	isTemporary bool
}

func (me *mockNetError) Error() string {
	return "mock net error"
}

func (me *mockNetError) Timeout() bool {
	return me.isTimeout
}

func (me *mockNetError) Temporary() bool {
	return me.isTemporary
}

// // mock custom error
type mockError struct{}

func (me *mockError) Error() string {
	return "mock error"
}
