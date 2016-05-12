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

	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	config := Config{
		Client: client,
	}

	req := &Request{
		config: config,
		chain:  chain,
	}

	resp := req.Expect()

	assert.False(t, resp == nil)

	req.chain.assertFailed(t)
	resp.chain.assertFailed(t)
}

func TestRequestEmpty(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "", "")

	resp := req.Expect()

	req.chain.assertOK(t)
	resp.chain.assertOK(t)
}

func TestRequestURL(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "http://example.com")
	req2 := NewRequest(config, "METHOD", "http://example.com/path")
	req3 := NewRequest(config, "METHOD", "/path")
	req4 := NewRequest(config, "METHOD", "path")

	req5 := NewRequest(config, "METHOD", "http://example.com/path").
		WithQuery("aa", "foo").WithQuery("bb", 123).WithQuery("cc", "*&@")

	req1.Expect()
	req1.chain.assertOK(t)
	assert.Equal(t, "http://example.com", client.req.URL.String())

	req2.Expect()
	req2.chain.assertOK(t)
	assert.Equal(t, "http://example.com/path", client.req.URL.String())

	req3.Expect()
	req3.chain.assertOK(t)
	assert.Equal(t, "/path", client.req.URL.String())

	req4.Expect()
	req4.chain.assertOK(t)
	assert.Equal(t, "path", client.req.URL.String())

	req5.Expect()
	req5.chain.assertOK(t)
	assert.Equal(t, "http://example.com/path?aa=foo&bb=123&cc=%2A%26%40",
		client.req.URL.String())
}

func TestExpectURLConcat(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	var reqs [5]*Request

	config1 := Config{
		BaseURL:  "",
		Client:   client,
		Reporter: reporter,
	}

	reqs[0] = NewRequest(config1, "METHOD", "http://example.com/path")

	config2 := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	reqs[1] = NewRequest(config2, "METHOD", "path")
	reqs[2] = NewRequest(config2, "METHOD", "/path")

	config3 := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs[3] = NewRequest(config3, "METHOD", "path")
	reqs[4] = NewRequest(config3, "METHOD", "/path")

	for _, req := range reqs {
		assert.Equal(t, "http://example.com/path", req.url.String())
	}

	empty1 := NewRequest(config1, "METHOD", "")
	empty2 := NewRequest(config2, "METHOD", "")
	empty3 := NewRequest(config3, "METHOD", "")

	assert.Equal(t, "", empty1.url.String())
	assert.Equal(t, "http://example.com", empty2.url.String())
	assert.Equal(t, "http://example.com/", empty3.url.String())
}

func TestExpectURLFormat(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	var reqs [3]*Request

	config1 := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs[0] = NewRequest(config1, "METHOD", "/foo/%s", "bar")
	reqs[1] = NewRequest(config1, "METHOD", "%sfoo%s", "/", "/bar")
	reqs[2] = NewRequest(config1, "%s", "/foo/bar")

	for _, req := range reqs {
		assert.Equal(t, "http://example.com/foo/bar", req.url.String())
	}

	config2 := Config{
		Reporter: newMockReporter(t),
	}

	r := NewRequest(config2, "GET", "%s", nil)

	r.chain.assertFailed(t)
}

func TestRequestHeaders(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
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
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyReader(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBody(bytes.NewBufferString("body"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	body, _ := ioutil.ReadAll(client.req.Body)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(body))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyBytes(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBytes([]byte("body"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	body, _ := ioutil.ReadAll(client.req.Body)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(body))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyJSON(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
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
	resp.chain.assertOK(t)

	body, _ := ioutil.ReadAll(client.req.Body)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `{"key":"value"}`, string(body))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestErrorMarshal(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithJSON(func() {})

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorSend(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}
