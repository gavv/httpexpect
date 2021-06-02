package httpexpect

import (
	"time"
)

type AssertionHandler interface {
	// Reporter is implemented for compatibility only and shouldn't be used directly.
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
	TestName         string
	AssertionHandler AssertionHandler
	Request          *Request
	Response         *Response
	RTT              *time.Duration
}

// Errorf implements Reporter for compatibility and shouldn't be used directly.
func (c *Context) Errorf(message string, args ...interface{}) {
	c.AssertionHandler.Errorf(message, args...)
}
