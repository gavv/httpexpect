package httpexpect

import (
	"time"
)

// Context contains information related to an assertion.
// It will be inherited by nested objects through the chain struct.
type Context struct {
	AssertionHandler AssertionHandler
	Request          *Request
	Response         *Response
	RTT              *time.Duration
}

// DEPRECATED
// Errorf implements Reporter for compatibility and shouldn't be used directly.
func (c *Context) Errorf(message string, args ...interface{}) {
	c.AssertionHandler.Errorf(message, args...)
}
