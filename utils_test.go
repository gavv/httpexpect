package httpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type closingBuffer struct {
	*bytes.Buffer
}

func (_ closingBuffer) Close() error {
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
	} else {
		return nil, c.err
	}
}

type mockChecker struct {
	testing *testing.T
	failed bool
}

func newMockChecker(t *testing.T) *mockChecker {
	return &mockChecker{testing: t}
}

func (c *mockChecker) AssertSuccess(t *testing.T) {
	assert.False(t, c.failed)
}

func (c *mockChecker) AssertFailed(t *testing.T) {
	assert.True(t, c.failed)
}

func (c *mockChecker) Reset() {
	c.failed = false
}

func (c *mockChecker) Clone() Checker {
	copy := *c
	return &copy
}

func (_ *mockChecker) Compare(a, b interface{}) bool {
	return assert.ObjectsAreEqual(a, b)
}

func (c *mockChecker) Failed() bool {
	return c.failed
}

func (c *mockChecker) Fail(message string, args... interface{}) {
	c.testing.Logf("Fail: " + message, args...)
	c.failed = true
}

func (c *mockChecker) Equal(expected, actual interface{}) {
	if !c.Compare(expected, actual) {
		c.testing.Logf("Equal: `%v` (expected) != `%v` (actual)", expected, actual)
		c.failed = true
	}
}

func (c *mockChecker) NotEqual(expected, actual interface{}) {
	if c.Compare(expected, actual) {
		c.testing.Logf("NotEqual: `%v` (expected) == `%v` (actual)", expected, actual)
		c.failed = true
	}
}
