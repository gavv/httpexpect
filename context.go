package httpexpect

import (
	"time"
)

type AssertionHandler interface {
	Reporter
	Failure(ctx *Context, failure Failure)
	Success(ctx *Context)
}

type DefaultAssertionHandler struct {
	Reporter  Reporter
	Formatter Formatter
}

func (d DefaultAssertionHandler) Errorf(message string, args ...interface{}) {
	d.Reporter.Errorf(message, args...)
}

func (d DefaultAssertionHandler) Failure(ctx *Context, failure Failure) {
	d.Errorf(d.Formatter.Failure(ctx, failure))
}

func (d DefaultAssertionHandler) Success(ctx *Context) {}

// Context contains information related to an assertion.
// It will be inherited by nested objects through the chain struct.
type Context struct {
	// Name of the test
	TestName  string
	Request   *Request
	Response  *Response
	Reporter  Reporter
	RTT       *time.Duration
	formatter Formatter
}

// Errorf implements Reporter for compatibility.
func (c *Context) Errorf(message string, args ...interface{}) {
	c.Reporter.Errorf(message, args...)
}
