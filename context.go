package httpexpect

import (
	"time"
)

// Context contains information related to an assertion.
// It will be inherited by nested objects through the chain struct.
type Context struct {
	TestName         string
	AssertionHandler AssertionHandler
	Request          *Request
	Response         *Response
	RTT              *time.Duration
}

// contextReporterWrapper is used only to keep compatibility with constructors using the Reporter interface.
// makeChain does a type assertion to extract, if necessary, the wrapped context, or use the Reporter as it is given.
type contextReporterWrapper struct{ ctx *Context }

func (c contextReporterWrapper) Errorf(message string, args ...interface{}) {
	panic("contextReporterWrapper is not meant to be called")
}

func wrapContext(ctx *Context) contextReporterWrapper {
	return contextReporterWrapper{ctx: ctx}
}
