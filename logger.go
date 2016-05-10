package httpexpect

import (
	"net/http"
	"net/http/httputil"
)

// CompactLogger implements logger. It logs requests in compact form.
type CompactLogger struct {
	backend LoggerBackend
}

// NewCompactLogger returns a new CompactLogger given a backend.
func NewCompactLogger(backend LoggerBackend) CompactLogger {
	return CompactLogger{backend}
}

// Request implements Logger.Request.
func (logger CompactLogger) Request(req *http.Request) {
	if req != nil {
		logger.backend.Logf("%s %s", req.Method, req.URL)
	}
}

// Response implements Logger.Response.
func (CompactLogger) Response(*http.Response) {
}

// DebugLogger implements logger. Uses net/http/httputil to dump requests and responses.
type DebugLogger struct {
	backend LoggerBackend
	body    bool
}

// NewDebugLogger returns a new DebugLogger given a backend and body.
// If body is true, request and response body is also logged.
func NewDebugLogger(backend LoggerBackend, body bool) DebugLogger {
	return DebugLogger{backend, body}
}

// Request implements Logger.Request.
func (logger DebugLogger) Request(req *http.Request) {
	if req != nil {
		dump, err := httputil.DumpRequestOut(req, logger.body)
		if err != nil {
			panic(err)
		}
		logger.backend.Logf("%s", dump)
	}
}

// Response implements Logger.Response.
func (logger DebugLogger) Response(resp *http.Response) {
	if resp != nil {
		dump, err := httputil.DumpResponse(resp, logger.body)
		if err != nil {
			panic(err)
		}
		logger.backend.Logf("%s", dump)
	}
}
