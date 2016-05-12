package httpexpect

import (
	"net/http"
	"net/http/httputil"
)

// CompactPrinter implements Printer. It prints requests in compact form.
type CompactPrinter struct {
	logger Logger
}

// NewCompactPrinter returns a new CompactPrinter given a logger.
func NewCompactPrinter(logger Logger) CompactPrinter {
	return CompactPrinter{logger}
}

// Request implements Printer.Request.
func (p CompactPrinter) Request(req *http.Request) {
	if req != nil {
		p.logger.Logf("%s %s", req.Method, req.URL)
	}
}

// Response implements Logger.Response.
func (CompactPrinter) Response(*http.Response) {
}

// DebugPrinter implements Printer. Uses net/http/httputil to dump
// both requests and responses.
type DebugPrinter struct {
	logger Logger
	body   bool
}

// NewDebugPrinter returns a new DebugPrinter given a logger and body
// flag. If body is true, request and response body is also printed.
func NewDebugPrinter(logger Logger, body bool) DebugPrinter {
	return DebugPrinter{logger, body}
}

// Request implements Printer.Request.
func (p DebugPrinter) Request(req *http.Request) {
	if req != nil {
		dump, err := httputil.DumpRequestOut(req, p.body)
		if err != nil {
			panic(err)
		}
		p.logger.Logf("%s", dump)
	}
}

// Response implements Printer.Response.
func (p DebugPrinter) Response(resp *http.Response) {
	if resp != nil {
		dump, err := httputil.DumpResponse(resp, p.body)
		if err != nil {
			panic(err)
		}
		p.logger.Logf("%s", dump)
	}
}
