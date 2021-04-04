package httpexpect

import "time"

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
