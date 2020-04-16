package httpexpect

import "time"

// Context contains information related to an assertion.
// It will be inherited by nested objects through the chain struct.
type Context struct {
	// Name of the test
	name      string
	request   *Request
	response  *Response
	reporter  *Reporter
	rtt       *time.Duration
	formatter Formatter
}
