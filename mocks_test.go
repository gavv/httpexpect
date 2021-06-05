package httpexpect

import (
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

type mockFormatter struct {
	testing          *testing.T
	formattedSuccess int
	formattedFailure int
}

func (m *mockFormatter) Success(context *Context) string {
	m.formattedSuccess++
	return context.TestName
}

func (m *mockFormatter) Failure(context *Context, failure Failure) string {
	m.formattedFailure++
	return context.TestName + "." + failure.assertionName
}

func newMockFormatter(t *testing.T) *mockFormatter {
	return &mockFormatter{testing: t}
}

func newMockAssertionHandler(t *testing.T) AssertionHandler {
	return DefaultAssertionHandler{
		Reporter:  newMockReporter(t),
		Formatter: DefaultFormatter{},
	}
}

func newMockContext(t *testing.T) *Context {
	return &Context{AssertionHandler: newMockAssertionHandler(t)}
}
