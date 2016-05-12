package httpexpect

import (
	"bytes"
	"net/http"
	"testing"
)

type closingBuffer struct {
	*bytes.Buffer
}

func (b closingBuffer) Close() error {
	return nil
}

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
	testing *testing.T
}

func (r mockReporter) Errorf(message string, args ...interface{}) {
	r.testing.Logf("Fail: "+message, args...)
}
