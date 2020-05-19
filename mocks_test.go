package httpexpect

import (
	"fmt"
	"net/http"
	"testing"
)

type mockClient struct {
	req  *http.Request
	resp http.Response
	err  error
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req
	if c.err == nil {
		c.resp.Header = c.req.Header
		c.resp.Body = c.req.Body
		return &c.resp, nil
	}
	return nil, c.err
}

type mockReporter struct {
	testing  *testing.T
	reported bool
}

func newMockReporter(t *testing.T) *mockReporter {
	return &mockReporter{t, false}
}

func (r *mockReporter) Errorf(message string, args ...interface{}) {
	r.testing.Logf("Fail: "+message, args...)
	r.reported = true
}

type stringReporter struct {
	msg      string
	reported bool
}

func newStringReporter() *stringReporter {
	return &stringReporter{"", false}
}

func (r *stringReporter) Errorf(message string, args ...interface{}) {
	r.msg = fmt.Sprintf(message, args...)
	r.reported = true
}
