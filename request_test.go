package httpexpect

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRequestFailed(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	checker.Fail("fail")

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "", "")

	resp := req.Expect()

	assert.False(t, resp == nil)
}

func TestRequestEmpty(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "", "")

	resp := req.Expect()

	assert.False(t, req.config.Checker == resp.checker)
}

func TestRequestURL(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req1 := NewRequest(config, "METHOD", "http://example.com")
	req2 := NewRequest(config, "METHOD", "http://example.com/path")
	req3 := NewRequest(config, "METHOD", "/path")
	req4 := NewRequest(config, "METHOD", "path")

	req5 := NewRequest(config, "METHOD", "http://example.com/path").
		WithQuery("aa", "foo").WithQuery("bb", 123).WithQuery("cc", "*&@")

	req1.Expect()
	checker.AssertSuccess(t)
	assert.Equal(t, "http://example.com", client.req.URL.String())

	req2.Expect()
	checker.AssertSuccess(t)
	assert.Equal(t, "http://example.com/path", client.req.URL.String())

	req3.Expect()
	checker.AssertSuccess(t)
	assert.Equal(t, "/path", client.req.URL.String())

	req4.Expect()
	checker.AssertSuccess(t)
	assert.Equal(t, "path", client.req.URL.String())

	req5.Expect()
	checker.AssertSuccess(t)
	assert.Equal(t, "http://example.com/path?aa=foo&bb=123&cc=%2A%26%40",
		client.req.URL.String())
}

func TestRequestHeaders(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeader("First-Header", "foo")

	req.WithHeaders(map[string]string{
		"Second-Header": "bar",
		"Third-Header":  "baz",
	})

	expectedHeaders := map[string][]string{
		"First-Header":  {"foo"},
		"Second-Header": {"bar"},
		"Third-Header":  {"baz"},
	}

	resp := req.Expect()
	checker.AssertSuccess(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyReader(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBody(bytes.NewBufferString("body"))

	resp := req.Expect()
	checker.AssertSuccess(t)

	body, _ := ioutil.ReadAll(client.req.Body)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(body))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyBytes(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBytes([]byte("body"))

	resp := req.Expect()
	checker.AssertSuccess(t)

	body, _ := ioutil.ReadAll(client.req.Body)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(body))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyJSON(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/json; charset=utf-8"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithJSON(map[string]interface{}{"key": "value"})

	resp := req.Expect()
	checker.AssertSuccess(t)

	body, _ := ioutil.ReadAll(client.req.Body)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `{"key":"value"}`, string(body))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestErrorMarshal(t *testing.T) {
	client := &mockClient{}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithJSON(func() {})

	resp := req.Expect()
	checker.AssertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorSend(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	checker := newMockChecker(t)

	config := Config{
		Client:  client,
		Checker: checker,
	}

	req := NewRequest(config, "METHOD", "url")

	resp := req.Expect()
	checker.AssertFailed(t)

	assert.True(t, resp.Raw() == nil)
}
