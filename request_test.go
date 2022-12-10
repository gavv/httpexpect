package httpexpect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestFailed(t *testing.T) {
	reporter := newMockReporter(t)

	chain := newChainWithDefaults("test", reporter)
	chain.fail(mockFailure())

	config := Config{
		Reporter: reporter,
	}

	config.fillDefaults()

	req := newRequest(chain, config, "GET", "")

	req.WithMatcher(func(resp *Response) {
	})
	req.WithTransformer(func(r *http.Request) {
	})
	req.WithClient(&http.Client{})
	req.WithHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	req.WithContext(context.TODO())
	req.WithTimeout(0)
	req.WithRedirectPolicy(FollowAllRedirects)
	req.WithMaxRedirects(1)
	req.WithRetryPolicy(RetryAllErrors)
	req.WithMaxRetries(1)
	req.WithRetryDelay(time.Millisecond, time.Millisecond)
	req.WithWebsocketUpgrade()
	req.WithWebsocketDialer(
		NewWebsocketDialer(
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	req.WithPath("foo", "bar")
	req.WithPathObject(map[string]interface{}{"foo": "bar"})
	req.WithQuery("foo", "bar")
	req.WithQueryObject(map[string]interface{}{"foo": "bar"})
	req.WithQueryString("foo=bar")
	req.WithURL("http://example.com")
	req.WithHeaders(map[string]string{"foo": "bar"})
	req.WithHeader("foo", "bar")
	req.WithCookies(map[string]string{"foo": "bar"})
	req.WithCookie("foo", "bar")
	req.WithBasicAuth("foo", "bar")
	req.WithHost("127.0.0.1")
	req.WithProto("HTTP/1.1")
	req.WithChunked(strings.NewReader("foo"))
	req.WithBytes([]byte("foo"))
	req.WithText("foo")
	req.WithJSON(map[string]string{"foo": "bar"})
	req.WithForm(map[string]string{"foo": "bar"})
	req.WithFormField("foo", "bar")
	req.WithFile("foo", "bar", strings.NewReader("baz"))
	req.WithFileBytes("foo", "bar", []byte("baz"))
	req.WithMultipart()

	resp := req.Expect()
	if resp == nil {
		panic("Expect returned nil")
	}

	req.chain.assertFailed(t)
	resp.chain.assertFailed(t)
}

func TestRequestEmpty(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "", "")

	resp := req.Expect()

	req.chain.assertOK(t)
	resp.chain.assertOK(t)
}

func TestRequestTime(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	for n := 0; n < 10; n++ {
		req := NewRequest(config, "", "")
		resp := req.Expect()
		require.NotNil(t, resp.rtt)
		assert.True(t, *resp.rtt >= 0)
	}
}

func TestRequestMatchers(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Reporter:       reporter,
		Client:         client,
	}

	req := NewRequest(config, "METHOD", "/")

	var resps []*Response

	req.WithMatcher(func(r *Response) {
		resps = append(resps, r)
	})

	assert.Equal(t, 0, len(resps))

	resp := req.Expect()

	assert.Equal(t, 1, len(resps))
	assert.Same(t, resp, resps[0])
}

func TestRequestTransformers(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	t.Run("save-ptr", func(t *testing.T) {
		var savedReq *http.Request
		transform := func(r *http.Request) {
			savedReq = r
		}

		req := NewRequest(config, "METHOD", "/")
		req.WithTransformer(transform)
		req.Expect().chain.assertOK(t)

		assert.NotNil(t, savedReq)
	})

	t.Run("append-header", func(t *testing.T) {
		req := NewRequest(config, "METHOD", "/")

		req.WithTransformer(func(r *http.Request) {
			r.Header.Add("foo", "11")
		})

		req.WithTransformer(func(r *http.Request) {
			r.Header.Add("bar", "22")
		})

		req.Expect().chain.assertOK(t)

		assert.Equal(t, []string{"11"}, client.req.Header["Foo"])
		assert.Equal(t, []string{"22"}, client.req.Header["Bar"])
	})

	t.Run("append-url", func(t *testing.T) {
		req := NewRequest(config, "METHOD", "/{arg1}/{arg2}")

		req.WithPath("arg1", "11")
		req.WithPath("arg2", "22")

		req.WithTransformer(func(r *http.Request) {
			r.URL.Path += "/33"
		})

		req.WithTransformer(func(r *http.Request) {
			r.URL.Path += "/44"
		})

		req.Expect().chain.assertOK(t)

		assert.Equal(t, "/11/22/33/44", client.req.URL.Path)
	})

	t.Run("nil-func", func(t *testing.T) {
		req := NewRequest(config, "METHOD", "/")
		req.WithTransformer(nil)
		req.chain.assertFailed(t)
	})
}

func TestRequestClient(t *testing.T) {
	factory := DefaultRequestFactory{}

	client1 := &mockClient{}
	client2 := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Reporter:       reporter,
		Client:         client1,
	}

	req1 := NewRequest(config, "METHOD", "/")
	req1.Expect().chain.assertOK(t)
	assert.NotNil(t, client1.req)

	req2 := NewRequest(config, "METHOD", "/")
	req2.WithClient(client2)
	req2.Expect().chain.assertOK(t)
	assert.NotNil(t, client2.req)

	req3 := NewRequest(config, "METHOD", "/")
	req3.WithClient(nil)
	req3.chain.assertFailed(t)
}

func TestRequestHandler(t *testing.T) {
	factory := DefaultRequestFactory{}

	var hr1 *http.Request
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hr1 = r
	})

	var hr2 *http.Request
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hr2 = r
	})

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Reporter:       reporter,
		Client: &http.Client{
			Transport: NewBinder(handler1),
		},
	}

	req1 := NewRequest(config, "METHOD", "/")
	req1.Expect().chain.assertOK(t)
	assert.NotNil(t, hr1)

	req2 := NewRequest(config, "METHOD", "/")
	req2.WithHandler(handler2)
	req2.Expect().chain.assertOK(t)
	assert.NotNil(t, hr2)

	req3 := NewRequest(config, "METHOD", "/")
	req3.WithHandler(nil)
	req3.chain.assertFailed(t)
}

func TestRequestHandlerResetClient(t *testing.T) {
	factory := DefaultRequestFactory{}

	var hr *http.Request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hr = r
	})

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Reporter:       reporter,
		Client:         client,
	}

	req := NewRequest(config, "METHOD", "/")
	req.WithHandler(handler)
	req.Expect().chain.assertOK(t)
	assert.NotNil(t, hr)
	assert.Nil(t, client.req)
}

func TestRequestHandlerResueClient(t *testing.T) {
	factory := DefaultRequestFactory{}

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	client := &http.Client{
		Transport: NewBinder(handler1),
		Jar:       NewJar(),
	}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Reporter:       reporter,
		Client:         client,
	}

	req := NewRequest(config, "METHOD", "/")
	req.WithHandler(handler2)

	assert.True(t, req.config.Client.(*http.Client).Jar == client.Jar)
}

func TestRequestProto(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "/")

	assert.Equal(t, 1, req.httpReq.ProtoMajor)
	assert.Equal(t, 1, req.httpReq.ProtoMinor)

	req.WithProto("HTTP/2.0")

	assert.Equal(t, 2, req.httpReq.ProtoMajor)
	assert.Equal(t, 0, req.httpReq.ProtoMinor)

	req.WithProto("HTTP/1.0")

	assert.Equal(t, 1, req.httpReq.ProtoMajor)
	assert.Equal(t, 0, req.httpReq.ProtoMinor)

	req.WithProto("bad")
	req.chain.assertFailed(t)

	assert.Equal(t, 1, req.httpReq.ProtoMajor)
	assert.Equal(t, 0, req.httpReq.ProtoMinor)
}

func TestRequestURLConcatenate(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config1 := Config{
		RequestFactory: factory,
		BaseURL:        "",
		Client:         client,
		Reporter:       reporter,
	}

	config2 := Config{
		RequestFactory: factory,
		BaseURL:        "http://example.com",
		Client:         client,
		Reporter:       reporter,
	}

	config3 := Config{
		RequestFactory: factory,
		BaseURL:        "http://example.com/",
		Client:         client,
		Reporter:       reporter,
	}

	reqs := []*Request{
		NewRequest(config2, "METHOD", "path"),
		NewRequest(config2, "METHOD", "/path"),
		NewRequest(config3, "METHOD", "path"),
		NewRequest(config3, "METHOD", "/path"),
		NewRequest(config3, "METHOD", "{arg}", "/path"),
		NewRequest(config3, "METHOD", "{arg}").WithPath("arg", "/path"),
	}

	for _, req := range reqs {
		req.Expect().chain.assertOK(t)
		assert.Equal(t, "http://example.com/path", client.req.URL.String())
	}

	empty1 := NewRequest(config1, "METHOD", "")
	empty2 := NewRequest(config2, "METHOD", "")
	empty3 := NewRequest(config3, "METHOD", "")

	empty1.Expect().chain.assertOK(t)
	empty2.Expect().chain.assertOK(t)
	empty3.Expect().chain.assertOK(t)

	assert.Equal(t, "", empty1.httpReq.URL.String())
	assert.Equal(t, "http://example.com", empty2.httpReq.URL.String())
	assert.Equal(t, "http://example.com/", empty3.httpReq.URL.String())
}

func TestRequestURLOverwrite(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config1 := Config{
		RequestFactory: factory,
		BaseURL:        "",
		Client:         client,
		Reporter:       reporter,
	}

	config2 := Config{
		RequestFactory: factory,
		BaseURL:        "http://foobar.com",
		Client:         client,
		Reporter:       reporter,
	}

	reqs := []*Request{
		NewRequest(config1, "METHOD", "/path").WithURL("http://example.com"),
		NewRequest(config1, "METHOD", "path").WithURL("http://example.com"),
		NewRequest(config1, "METHOD", "/path").WithURL("http://example.com/"),
		NewRequest(config1, "METHOD", "path").WithURL("http://example.com/"),
		NewRequest(config2, "METHOD", "/path").WithURL("http://example.com"),
		NewRequest(config2, "METHOD", "path").WithURL("http://example.com"),
		NewRequest(config2, "METHOD", "/path").WithURL("http://example.com/"),
		NewRequest(config2, "METHOD", "path").WithURL("http://example.com/"),
	}

	for _, req := range reqs {
		req.Expect().chain.assertOK(t)
		assert.Equal(t, "http://example.com/path", client.req.URL.String())
	}
}

func TestRequestURLInterpolate(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	var reqs [3]*Request

	config := Config{
		RequestFactory: factory,
		BaseURL:        "http://example.com/",
		Client:         client,
		Reporter:       reporter,
	}

	reqs[0] = NewRequest(config, "METHOD", "/foo/{arg}", "bar")
	reqs[1] = NewRequest(config, "METHOD", "{arg}foo{arg}", "/", "/bar")
	reqs[2] = NewRequest(config, "METHOD", "{arg}", "/foo/bar")

	for _, req := range reqs {
		req.Expect().chain.assertOK(t)
		assert.Equal(t, "http://example.com/foo/bar", client.req.URL.String())
	}

	r1 := NewRequest(config, "METHOD", "/{arg1}/{arg2}", "foo")
	r1.Expect().chain.assertOK(t)
	assert.Equal(t, "http://example.com/foo/%7Barg2%7D",
		client.req.URL.String())

	r2 := NewRequest(config, "METHOD", "/{arg1}/{arg2}/{arg3}")
	r2.WithPath("ARG3", "foo")
	r2.WithPath("arg2", "bar")
	r2.Expect().chain.assertOK(t)
	assert.Equal(t, "http://example.com/%7Barg1%7D/bar/foo",
		client.req.URL.String())

	r3 := NewRequest(config, "METHOD", "/{arg1}.{arg2}.{arg3}")
	r3.WithPath("arg2", "bar")
	r3.WithPathObject(map[string]string{"ARG1": "foo", "arg3": "baz"})
	r3.WithPathObject(nil)
	r3.Expect().chain.assertOK(t)
	assert.Equal(t, "http://example.com/foo.bar.baz",
		client.req.URL.String())

	type S struct {
		Arg1 string
		A2   int `path:"arg2"`
		A3   int `path:"-"`
	}

	r4 := NewRequest(config, "METHOD", "/{arg1}{arg2}")
	r4.WithPathObject(S{"foo", 1, 2})
	r4.Expect().chain.assertOK(t)
	assert.Equal(t, "http://example.com/foo1", client.req.URL.String())

	r5 := NewRequest(config, "METHOD", "/{arg1}{arg2}")
	r5.WithPathObject(&S{"foo", 1, 2})
	r5.Expect().chain.assertOK(t)
	assert.Equal(t, "http://example.com/foo1", client.req.URL.String())

	r6 := NewRequest(config, "GET", "{arg}", nil)
	r6.chain.assertFailed(t)

	r7 := NewRequest(config, "GET", "{arg}")
	r7.chain.assertOK(t)
	r7.WithPath("arg", nil)
	r7.chain.assertFailed(t)

	r8 := NewRequest(config, "GET", "{arg}")
	r8.chain.assertOK(t)
	r8.WithPath("bad", "value")
	r8.chain.assertFailed(t)

	r9 := NewRequest(config, "GET", "{arg")
	r9.chain.assertFailed(t)
	r9.WithPath("arg", "foo")
	r9.chain.assertFailed(t)

	r10 := NewRequest(config, "GET", "{arg}")
	r10.chain.assertOK(t)
	r10.WithPathObject(func() {})
	r10.chain.assertFailed(t)
}

func TestRequestURLQuery(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
		BaseURL:        "http://example.com",
	}

	req1 := NewRequest(config, "METHOD", "/path").
		WithQuery("aa", "foo").WithQuery("bb", 123).WithQuery("cc", "*&@")

	q := map[string]interface{}{
		"bb": 123,
		"cc": "*&@",
	}

	req2 := NewRequest(config, "METHOD", "/path").
		WithQuery("aa", "foo").
		WithQueryObject(q)

	type S struct {
		Bb int    `url:"bb"`
		Cc string `url:"cc"`
		Dd string `url:"-"`
	}

	req3 := NewRequest(config, "METHOD", "/path").
		WithQueryObject(S{123, "*&@", "dummy"}).WithQuery("aa", "foo")

	req4 := NewRequest(config, "METHOD", "/path").
		WithQueryObject(&S{123, "*&@", "dummy"}).WithQuery("aa", "foo")

	req5 := NewRequest(config, "METHOD", "/path").
		WithQuery("bb", 123).
		WithQueryString("aa=foo&cc=%2A%26%40")

	req6 := NewRequest(config, "METHOD", "/path").
		WithQueryString("aa=foo&cc=%2A%26%40").
		WithQuery("bb", 123)

	for _, req := range []*Request{req1, req2, req3, req4, req5, req6} {
		client.req = nil
		req.Expect()
		req.chain.assertOK(t)
		assert.Equal(t, "http://example.com/path?aa=foo&bb=123&cc=%2A%26%40",
			client.req.URL.String())
	}

	req7 := NewRequest(config, "METHOD", "/path").
		WithQuery("foo", "bar").
		WithQueryObject(nil)

	req7.Expect()
	req7.chain.assertOK(t)
	assert.Equal(t, "http://example.com/path?foo=bar", client.req.URL.String())

	NewRequest(config, "METHOD", "/path").
		WithQueryObject(func() {}).chain.assertFailed(t)

	NewRequest(config, "METHOD", "/path").
		WithQueryString("%").chain.assertFailed(t)
}

func TestRequestHeaders(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeader("first-header", "foo")

	req.WithHeaders(map[string]string{
		"Second-Header": "bar",
		"content-Type":  "baz",
		"HOST":          "example.com",
	})

	expectedHeaders := map[string][]string{
		"First-Header":  {"foo"},
		"Second-Header": {"bar"},
		"Content-Type":  {"baz"},
	}

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "example.com", client.req.Host)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestCookies(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithCookie("foo", "1")
	req.WithCookie("bar", "2 ")

	req.WithCookies(map[string]string{
		"baz": " 3",
	})

	expectedHeaders := map[string][]string{
		"Cookie": {`foo=1; bar="2 "; baz=" 3"`},
	}

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBasicAuth(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBasicAuth("Aladdin", "open sesame")
	req.chain.assertOK(t)

	assert.Equal(t, "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==",
		req.httpReq.Header.Get("Authorization"))
}

func TestRequestWithHost(t *testing.T) {
	factory1 := DefaultRequestFactory{}
	client1 := &mockClient{}
	reporter1 := newMockReporter(t)

	config1 := Config{
		RequestFactory: factory1,
		Client:         client1,
		Reporter:       reporter1,
	}

	req1 := NewRequest(config1, "METHOD", "url")

	req1.WithHost("example.com")

	resp := req1.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client1.req.Method)
	assert.Equal(t, "example.com", client1.req.Host)
	assert.Equal(t, "url", client1.req.URL.String())

	assert.Same(t, &client1.resp, resp.Raw())

	factory2 := DefaultRequestFactory{}
	client2 := &mockClient{}
	reporter2 := newMockReporter(t)

	config2 := Config{
		RequestFactory: factory2,
		Client:         client2,
		Reporter:       reporter2,
	}

	req2 := NewRequest(config2, "METHOD", "url")

	req2.WithHeader("HOST", "example1.com")
	req2.WithHost("example2.com")

	req2.Expect().chain.assertOK(t)

	assert.Equal(t, "example2.com", client2.req.Host)

	factory3 := DefaultRequestFactory{}
	client3 := &mockClient{}
	reporter3 := newMockReporter(t)

	config3 := Config{
		RequestFactory: factory3,
		Client:         client3,
		Reporter:       reporter3,
	}

	req3 := NewRequest(config3, "METHOD", "url")

	req3.WithHost("example2.com")
	req3.WithHeader("HOST", "example1.com")

	req3.Expect().chain.assertOK(t)

	assert.Equal(t, "example1.com", client3.req.Host)
}

func TestRequestBodyChunked(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithChunked(bytes.NewBufferString("body"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.False(t, client.req.Body == nil)
	assert.Equal(t, int64(-1), client.req.ContentLength)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyChunkedNil(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithChunked(nil)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.True(t, client.req.Body == http.NoBody)
	assert.Equal(t, int64(0), client.req.ContentLength)
}

func TestRequestBodyChunkedProto(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")

	req1.WithProto("HTTP/1.0")
	assert.Equal(t, 1, req1.httpReq.ProtoMajor)
	assert.Equal(t, 0, req1.httpReq.ProtoMinor)

	req1.WithChunked(bytes.NewBufferString("body"))
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")

	req2.WithProto("HTTP/2.0")
	assert.Equal(t, 2, req2.httpReq.ProtoMajor)
	assert.Equal(t, 0, req2.httpReq.ProtoMinor)

	req2.WithChunked(bytes.NewBufferString("body"))
	assert.Equal(t, 2, req2.httpReq.ProtoMajor)
	assert.Equal(t, 0, req2.httpReq.ProtoMinor)
}

func TestRequestBodyBytes(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "/path")

	req.WithBytes([]byte("body"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.False(t, client.req.Body == nil)
	assert.Equal(t, int64(len("body")), client.req.ContentLength)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "/path", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyBytesNil(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBytes(nil)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.True(t, client.req.Body == http.NoBody)
	assert.Equal(t, int64(0), client.req.ContentLength)
}

func TestRequestBodyText(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"text/plain; charset=utf-8"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithText("some text")

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, "some text", string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyForm(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithForm(map[string]interface{}{
		"a": 1,
		"b": "2",
	})

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2`, string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyField(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithFormField("a", 1)
	req.WithFormField("b", "2")

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2`, string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyFormStruct(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	req := NewRequest(config, "METHOD", "url")

	type S struct {
		A string `form:"a"`
		B int    `form:"b"`
		C int    `form:"-"`
	}

	req.WithForm(S{"1", 2, 3})

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2`, string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyFormCombined(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	req := NewRequest(config, "METHOD", "url")

	type S struct {
		A int `form:"a"`
	}

	req.WithForm(S{A: 1})
	req.WithForm(map[string]string{"b": "2"})
	req.WithFormField("c", 3)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2&c=3`, string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestBodyMultipart(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "POST", "url")

	req.WithMultipart()
	req.WithForm(map[string]string{"b": "1", "c": "2"})
	req.WithFormField("a", 3)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "POST", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())

	mediatype, params, err := mime.ParseMediaType(client.req.Header.Get("Content-Type"))

	assert.True(t, err == nil)
	assert.Equal(t, "multipart/form-data", mediatype)
	assert.True(t, params["boundary"] != "")

	reader := multipart.NewReader(bytes.NewReader(resp.content), params["boundary"])

	part1, _ := reader.NextPart()
	assert.Equal(t, "b", part1.FormName())
	assert.Equal(t, "", part1.FileName())
	b1, _ := ioutil.ReadAll(part1)
	assert.Equal(t, "1", string(b1))

	part2, _ := reader.NextPart()
	assert.Equal(t, "c", part2.FormName())
	assert.Equal(t, "", part2.FileName())
	b2, _ := ioutil.ReadAll(part2)
	assert.Equal(t, "2", string(b2))

	part3, _ := reader.NextPart()
	assert.Equal(t, "a", part3.FormName())
	assert.Equal(t, "", part3.FileName())
	b3, _ := ioutil.ReadAll(part3)
	assert.Equal(t, "3", string(b3))

	eof, _ := reader.NextPart()
	assert.True(t, eof == nil)
}

func TestRequestBodyMultipartFile(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "POST", "url")

	fh, _ := ioutil.TempFile("", "httpexpect")
	filename2 := fh.Name()
	_, _ = fh.WriteString("2")
	fh.Close()
	defer os.Remove(filename2)

	req.WithMultipart()
	req.WithForm(map[string]string{"a": "1"})
	req.WithFile("b", filename2)
	req.WithFile("c", "filename3", strings.NewReader("3"))
	req.WithFileBytes("d", "filename4", []byte("4"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "POST", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())

	mediatype, params, err := mime.ParseMediaType(client.req.Header.Get("Content-Type"))

	assert.True(t, err == nil)
	assert.Equal(t, "multipart/form-data", mediatype)
	assert.True(t, params["boundary"] != "")

	reader := multipart.NewReader(bytes.NewReader(resp.content), params["boundary"])

	part1, _ := reader.NextPart()
	assert.Equal(t, "a", part1.FormName())
	assert.Equal(t, "", part1.FileName())
	b1, _ := ioutil.ReadAll(part1)
	assert.Equal(t, "1", string(b1))

	part2, _ := reader.NextPart()
	assert.Equal(t, "b", part2.FormName())
	assert.Equal(t, filepath.Base(filename2), filepath.Base(part2.FileName()))
	b2, _ := ioutil.ReadAll(part2)
	assert.Equal(t, "2", string(b2))

	part3, _ := reader.NextPart()
	assert.Equal(t, "c", part3.FormName())
	assert.Equal(t, "filename3", filepath.Base(part3.FileName()))
	b3, _ := ioutil.ReadAll(part3)
	assert.Equal(t, "3", string(b3))

	part4, _ := reader.NextPart()
	assert.Equal(t, "d", part4.FormName())
	assert.Equal(t, "filename4", filepath.Base(part4.FileName()))
	b4, _ := ioutil.ReadAll(part4)
	assert.Equal(t, "4", string(b4))

	eof, _ := reader.NextPart()
	assert.True(t, eof == nil)
}

func TestRequestBodyJSON(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
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

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `{"key":"value"}`, string(resp.content))

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequestContentLength(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithChunked(bytes.NewReader([]byte("12345")))
	req1.Expect().chain.assertOK(t)
	assert.Equal(t, int64(-1), client.req.ContentLength)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithBytes([]byte("12345"))
	req2.Expect().chain.assertOK(t)
	assert.Equal(t, int64(5), client.req.ContentLength)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithText("12345")
	req3.Expect().chain.assertOK(t)
	assert.Equal(t, int64(5), client.req.ContentLength)

	j, _ := json.Marshal(map[string]string{"a": "b"})
	req4 := NewRequest(config, "METHOD", "url")
	req4.WithJSON(map[string]string{"a": "b"})
	req4.Expect().chain.assertOK(t)
	assert.Equal(t, int64(len(j)), client.req.ContentLength)

	f := `a=b`
	req5 := NewRequest(config, "METHOD", "url")
	req5.WithForm(map[string]string{"a": "b"})
	req5.Expect().chain.assertOK(t)
	assert.Equal(t, int64(len(f)), client.req.ContentLength)

	req6 := NewRequest(config, "METHOD", "url")
	req6.WithFormField("a", "b")
	req6.Expect().chain.assertOK(t)
	assert.Equal(t, int64(len(f)), client.req.ContentLength)

	req7 := NewRequest(config, "METHOD", "url")
	req7.WithMultipart()
	req7.WithFileBytes("a", "b", []byte("12345"))
	req7.Expect().chain.assertOK(t)
	assert.True(t, client.req.ContentLength > 0)
}

func TestRequestContentTypeOverwrite(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithText("hello")
	req1.WithHeader("Content-Type", "foo")
	req1.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithHeader("Content-Type", "foo")
	req2.WithText("hello")
	req2.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithJSON(map[string]interface{}{"a": "b"})
	req3.WithHeader("Content-Type", "foo")
	req3.WithHeader("Content-Type", "bar")
	req3.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)

	req4 := NewRequest(config, "METHOD", "url")
	req4.WithForm(map[string]interface{}{"a": "b"})
	req4.WithHeader("Content-Type", "foo")
	req4.WithHeader("Content-Type", "bar")
	req4.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)

	req5 := NewRequest(config, "METHOD", "url")
	req5.WithMultipart()
	req5.WithForm(map[string]interface{}{"a": "b"})
	req5.WithHeader("Content-Type", "foo")
	req5.WithHeader("Content-Type", "bar")
	req5.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)
}

func TestRequestErrorMarshalForm(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithForm(func() {})

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorMarshalJSON(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithJSON(func() {})

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorReadFile(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithMultipart()
	req.WithFile("", "")

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorSend(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorConflictBody(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithChunked(nil)
	req1.chain.assertOK(t)
	req1.WithChunked(nil)
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithChunked(nil)
	req2.chain.assertOK(t)
	req2.WithBytes(nil)
	req2.chain.assertFailed(t)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithChunked(nil)
	req3.chain.assertOK(t)
	req3.WithText("")
	req3.chain.assertFailed(t)

	req4 := NewRequest(config, "METHOD", "url")
	req4.WithChunked(nil)
	req4.chain.assertOK(t)
	req4.WithJSON(map[string]interface{}{"a": "b"})
	req4.chain.assertFailed(t)

	req5 := NewRequest(config, "METHOD", "url")
	req5.WithChunked(nil)
	req5.chain.assertOK(t)
	req5.WithForm(map[string]interface{}{"a": "b"})
	req5.Expect()
	req5.chain.assertFailed(t)

	req6 := NewRequest(config, "METHOD", "url")
	req6.WithChunked(nil)
	req6.chain.assertOK(t)
	req6.WithFormField("a", "b")
	req6.Expect()
	req6.chain.assertFailed(t)

	req7 := NewRequest(config, "METHOD", "url")
	req7.WithChunked(nil)
	req7.chain.assertOK(t)
	req7.WithMultipart()
	req7.chain.assertFailed(t)
}

func TestRequestErrorConflictType(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithText("")
	req1.chain.assertOK(t)
	req1.WithJSON(map[string]interface{}{"a": "b"})
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithText("")
	req2.chain.assertOK(t)
	req2.WithForm(map[string]interface{}{"a": "b"})
	req2.chain.assertFailed(t)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithText("")
	req3.chain.assertOK(t)
	req3.WithFormField("a", "b")
	req3.chain.assertFailed(t)

	req4 := NewRequest(config, "METHOD", "url")
	req4.WithText("")
	req4.chain.assertOK(t)
	req4.WithMultipart()
	req4.chain.assertFailed(t)
}

func TestRequestErrorConflictMultipart(t *testing.T) {
	factory := DefaultRequestFactory{}

	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		RequestFactory: factory,
		Client:         client,
		Reporter:       reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithForm(map[string]interface{}{"a": "b"})
	req1.chain.assertOK(t)
	req1.WithMultipart()
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithFormField("a", "b")
	req2.chain.assertOK(t)
	req2.WithMultipart()
	req2.chain.assertFailed(t)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithFileBytes("a", "a", []byte("a"))
	req3.chain.assertFailed(t)
}

func TestRequestRedirect(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("dont follow redirects policy", func(t *testing.T) {
		t.Run("request has no body", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					return assert.Equal(t, http.NoBody, r.Body)
				})

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(DontFollowRedirects)
			req.chain.assertOK(t)

			// Should return redirection response
			resp := req.Expect().
				Status(tp.redirectHTTPStatusCode).
				Header("Location").
				Equal("/redirect")
			resp.chain.assertOK(t)

			// Should set GetBody
			assert.Nil(t, req.httpReq.GetBody)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Equal(t, http.ErrUseLastResponse, httpClient.CheckRedirect(req.httpReq, nil))

			// Should do round trip
			assert.Equal(t, 1, tp.tripCount)
		})

		t.Run("request has body", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					b, err := ioutil.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, "test body", string(b))

					return true
				})

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(DontFollowRedirects).
				WithText("test body")
			req.chain.assertOK(t)

			// Should return redirection response
			resp := req.Expect().
				Status(tp.redirectHTTPStatusCode).
				Header("Location").
				Equal("/redirect")
			resp.chain.assertOK(t)

			// Should set GetBody
			assert.Nil(t, req.httpReq.GetBody)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Equal(t, http.ErrUseLastResponse, httpClient.CheckRedirect(req.httpReq, nil))

			// Should do round trip
			assert.Equal(t, 1, tp.tripCount)
		})
	})

	t.Run("follow all redirects policy", func(t *testing.T) {
		t.Run("request has no body", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					return assert.Equal(t, http.NoBody, r.Body)
				}).WithMaxRedirect(1)

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1)
			req.chain.assertOK(t)

			// Should return OK response
			resp := req.Expect().
				Status(http.StatusOK)
			resp.chain.assertOK(t)

			// Should set GetBody
			gb, err := req.httpReq.GetBody()
			assert.NoError(t, err)
			assert.Equal(t, http.NoBody, gb)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
			assert.Equal(t,
				errors.New("stopped after 1 redirects"),
				httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
			)

			// Should do round trip
			assert.Equal(t, 2, tp.tripCount)
		})

		t.Run("request has no body and too many redirects", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					return assert.Equal(t, http.NoBody, r.Body)
				})

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1)
			req.chain.assertOK(t)

			// Should error
			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should set GetBody
			gb, err := req.httpReq.GetBody()
			assert.NoError(t, err)
			assert.Equal(t, http.NoBody, gb)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
			assert.Equal(t,
				errors.New("stopped after 1 redirects"),
				httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
			)

			// Should do round trip
			assert.Equal(t, 2, tp.tripCount)
		})

		t.Run("request has body", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					b, err := ioutil.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, "test body", string(b))

					return true
				}).WithMaxRedirect(1)

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1).
				WithText("test body")
			req.chain.assertOK(t)

			// Should return OK response
			resp := req.Expect().
				Status(http.StatusOK)
			resp.chain.assertOK(t)

			// Should set GetBody
			gb, err := req.httpReq.GetBody()
			assert.NoError(t, err)
			b, err := ioutil.ReadAll(gb)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
			assert.Equal(t,
				errors.New("stopped after 1 redirects"),
				httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
			)

			// Should do round trip
			assert.Equal(t, 2, tp.tripCount)
		})

		t.Run("request has body and too many redirects", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					b, err := ioutil.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, "test body", string(b))

					return true
				})

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1).
				WithText("test body")
			req.chain.assertOK(t)

			// Should error
			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should set GetBody
			gb, err := req.httpReq.GetBody()
			assert.NoError(t, err)
			b, err := ioutil.ReadAll(gb)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
			assert.Equal(t,
				errors.New("stopped after 1 redirects"),
				httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
			)

			// Should do round trip
			assert.Equal(t, 2, tp.tripCount)
		})
	})

	t.Run("follow redirects without body policy", func(t *testing.T) {
		t.Run("request has no body", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					return assert.Contains(t, []interface{}{nil, http.NoBody}, r.Body)
				}).WithMaxRedirect(1)

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowRedirectsWithoutBody).
				WithMaxRedirects(1)
			req.chain.assertOK(t)

			// Should return OK response
			resp := req.Expect().
				Status(http.StatusOK)
			resp.chain.assertOK(t)

			// Should set GetBody
			assert.Nil(t, req.httpReq.GetBody)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
			assert.Equal(t,
				errors.New("stopped after 1 redirects"),
				httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
			)

			// Should do round trip
			assert.Equal(t, 2, tp.tripCount)
		})

		t.Run("request has no body and too many redirects", func(t *testing.T) {
			tp := newMockTransportRedirect(t).
				WithAssertFn(func(r *http.Request) bool {
					return assert.Contains(t, []interface{}{nil, http.NoBody}, r.Body)
				})

			client := &http.Client{
				Transport: tp,
			}

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowRedirectsWithoutBody).
				WithMaxRedirects(1)
			req.chain.assertOK(t)

			// Should error
			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should set GetBody
			assert.Nil(t, req.httpReq.GetBody)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
			assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
			assert.Equal(t,
				errors.New("stopped after 1 redirects"),
				httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
			)

			// Should do round trip
			assert.Equal(t, 2, tp.tripCount)
		})

		t.Run("request has body and redirected with status permanent redirect",
			func(t *testing.T) {
				tp := newMockTransportRedirect(t).
					WithAssertFn(func(r *http.Request) bool {
						b, err := ioutil.ReadAll(r.Body)
						assert.NoError(t, err)
						assert.Equal(t, "test body", string(b))

						return true
					}).WithMaxRedirect(1)

				client := &http.Client{
					Transport: tp,
				}

				config := Config{
					Client:   client,
					Reporter: reporter,
				}

				req := NewRequest(config, http.MethodPut, "/url").
					WithRedirectPolicy(FollowRedirectsWithoutBody).
					WithMaxRedirects(1).
					WithText("test body")
				req.chain.assertOK(t)

				// Should return redirection response
				resp := req.Expect().
					Status(tp.redirectHTTPStatusCode)
				resp.chain.assertOK(t)

				// Should set GetBody
				assert.Nil(t, req.httpReq.GetBody)

				// Should set CheckRedirect
				httpClient, _ := req.config.Client.(*http.Client)
				assert.NotNil(t, httpClient.CheckRedirect)
				assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
				assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
				assert.Equal(t,
					errors.New("stopped after 1 redirects"),
					httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
				)

				// Should do round trip
				assert.Equal(t, 1, tp.tripCount)
			})

		t.Run("request has body and redirected with status moved permanently",
			func(t *testing.T) {
				tp := newMockTransportRedirect(t).
					WithAssertFn(func(r *http.Request) bool {
						if r.URL.String() == "/url" {
							assert.Equal(t, r.Method, http.MethodPut)

							b, err := ioutil.ReadAll(r.Body)
							assert.NoError(t, err)
							assert.Equal(t, "test body", string(b))
						} else if r.URL.String() == "/redirect" {
							assert.Equal(t, r.Method, http.MethodGet)
							assert.Nil(t, r.Body)
						} else {
							t.Fatalf("invalid request URL")
						}

						return true
					}).WithRedirectHTTPStatusCode(http.StatusMovedPermanently).WithMaxRedirect(1)

				client := &http.Client{
					Transport: tp,
				}

				config := Config{
					Client:   client,
					Reporter: reporter,
				}

				req := NewRequest(config, http.MethodPut, "/url").
					WithRedirectPolicy(FollowRedirectsWithoutBody).
					WithMaxRedirects(1).
					WithText("test body")
				req.chain.assertOK(t)

				// Should return OK response
				resp := req.Expect().
					Status(http.StatusOK)
				resp.chain.assertOK(t)

				// Should set GetBody
				assert.Nil(t, req.httpReq.GetBody)

				// Should set CheckRedirect
				httpClient, _ := req.config.Client.(*http.Client)
				assert.NotNil(t, httpClient.CheckRedirect)
				assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
				assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
				assert.Equal(t,
					errors.New("stopped after 1 redirects"),
					httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
				)

				// Should do round trip
				assert.Equal(t, 2, tp.tripCount)
			})

		t.Run(`request has body 
		and redirected with status moved permanently 
		and too many redirects`,
			func(t *testing.T) {
				tp := newMockTransportRedirect(t).
					WithAssertFn(func(r *http.Request) bool {
						if r.URL.String() == "/url" {
							assert.Equal(t, r.Method, http.MethodPut)

							b, err := ioutil.ReadAll(r.Body)
							assert.NoError(t, err)
							assert.Equal(t, "test body", string(b))
						} else if r.URL.String() == "/redirect" {
							assert.Equal(t, r.Method, http.MethodGet)
							assert.Nil(t, r.Body)
						} else {
							t.Fatalf("invalid request URL")
						}

						return true
					}).WithRedirectHTTPStatusCode(http.StatusMovedPermanently)

				client := &http.Client{
					Transport: tp,
				}

				config := Config{
					Client:   client,
					Reporter: reporter,
				}

				req := NewRequest(config, http.MethodPut, "/url").
					WithRedirectPolicy(FollowRedirectsWithoutBody).
					WithMaxRedirects(1).
					WithText("test body")
				req.chain.assertOK(t)

				// Should error
				resp := req.Expect()
				resp.chain.assertFailed(t)

				// Should set GetBody
				assert.Nil(t, req.httpReq.GetBody)

				// Should set CheckRedirect
				httpClient, _ := req.config.Client.(*http.Client)
				assert.NotNil(t, httpClient.CheckRedirect)
				assert.Nil(t, httpClient.CheckRedirect(req.httpReq, nil))
				assert.Nil(t, httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 1)))
				assert.Equal(t,
					errors.New("stopped after 1 redirects"),
					httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)),
				)

				// Should do round trip
				assert.Equal(t, 2, tp.tripCount)
			})
	})
}

// mockTransportRedirect mocks a transport that implements RoundTripper
//
// When tripCount < maxRedirect,
// mockTransportRedirect responses with redirectHTTPStatusCode
//
// When tripCount = maxRedirect,
// mockTransportRedirect responses with HTTP 200 OK
type mockTransportRedirect struct {
	t *testing.T

	// assertFn asserts the HTTP request
	assertFn func(*http.Request) bool

	// redirectHTTPStatusCode indicates the HTTP status code of redirection response
	redirectHTTPStatusCode int

	// tripCount tracks the number of trip that has been done
	tripCount int

	// maxRedirect indicates the number of trip that can be done for redirection.
	// -1 means always redirect.
	maxRedirect int
}

func newMockTransportRedirect(
	t *testing.T,
) *mockTransportRedirect {
	return &mockTransportRedirect{
		t:                      t,
		assertFn:               nil,
		redirectHTTPStatusCode: http.StatusPermanentRedirect,
		maxRedirect:            -1,
	}
}

func (mt *mockTransportRedirect) RoundTrip(origReq *http.Request) (
	*http.Response, error,
) {
	mt.tripCount++

	if mt.assertFn != nil {
		assert.True(mt.t, mt.assertFn(origReq))
	}

	res := httptest.NewRecorder()

	if mt.maxRedirect == -1 || mt.tripCount <= mt.maxRedirect {
		res.Result().StatusCode = mt.redirectHTTPStatusCode
		res.Result().Header.Set("Location", "/redirect")
	} else {
		res.Result().StatusCode = http.StatusOK
	}

	return res.Result(), nil
}

func (mt *mockTransportRedirect) WithAssertFn(
	fn func(*http.Request) bool,
) *mockTransportRedirect {
	mt.assertFn = fn

	return mt
}

func (mt *mockTransportRedirect) WithRedirectHTTPStatusCode(
	statusCode int,
) *mockTransportRedirect {
	if !(statusCode >= 300 && statusCode < 400) {
		mt.t.Fatal("invalid redirect status code")
	}

	mt.redirectHTTPStatusCode = statusCode

	return mt
}

func (mt *mockTransportRedirect) WithMaxRedirect(
	maxRedirect int,
) *mockTransportRedirect {
	if maxRedirect != -1 && maxRedirect < 0 {
		mt.t.Fatal("max redirect less than 0")
	}

	mt.maxRedirect = maxRedirect

	return mt
}

func TestRequestRetry(t *testing.T) {
	reporter := newMockReporter(t)

	newNoErrClient := func(cb func(req *http.Request)) *mockClient {
		return &mockClient{
			resp: http.Response{
				StatusCode: http.StatusOK,
			},
			cb: cb,
		}
	}

	newTempNetErrClient := func(cb func(req *http.Request)) *mockClient {
		return &mockClient{
			err: &mockNetError{
				isTemporary: true,
			},
			cb: cb,
		}
	}

	newTempServerErrClient := func(cb func(req *http.Request)) *mockClient {
		return &mockClient{
			resp: http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			cb: cb,
		}
	}

	newHTTPErrClient := func(cb func(req *http.Request)) *mockClient {
		return &mockClient{
			resp: http.Response{
				StatusCode: http.StatusBadRequest,
			},
			cb: cb,
		}
	}

	noopSleepFn := func(time.Duration) {}

	t.Run("dont retry policy", func(t *testing.T) {
		t.Run("no error", func(t *testing.T) {
			callCount := 0

			client := newNoErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("temporary network error", func(t *testing.T) {
			callCount := 0

			client := newTempNetErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("temporary server error", func(t *testing.T) {
			callCount := 0

			client := newTempServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("http error", func(t *testing.T) {
			callCount := 0

			client := newHTTPErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

	})

	t.Run("retry temporary network errors policy", func(t *testing.T) {
		t.Run("no error", func(t *testing.T) {
			callCount := 0

			client := newNoErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkErrors)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("temporary network error", func(t *testing.T) {
			callCount := 0

			client := newTempNetErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("temporary server error", func(t *testing.T) {
			callCount := 0

			client := newTempServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkErrors).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("http error", func(t *testing.T) {
			callCount := 0

			client := newHTTPErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkErrors).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})
	})

	t.Run("retry temporary network and server errors policy", func(t *testing.T) {
		t.Run("no error", func(t *testing.T) {
			callCount := 0

			client := newNoErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkAndServerErrors)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("temporary network error", func(t *testing.T) {
			callCount := 0

			client := newTempNetErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkAndServerErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("temporary server error", func(t *testing.T) {
			callCount := 0

			client := newTempServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkAndServerErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertOK(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("http error", func(t *testing.T) {
			callCount := 0

			client := newHTTPErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTemporaryNetworkAndServerErrors).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})
	})

	t.Run("retry all errors policy", func(t *testing.T) {
		t.Run("no error", func(t *testing.T) {
			callCount := 0

			client := newNoErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertOK(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("temporary network error", func(t *testing.T) {
			callCount := 0

			client := newTempNetErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("temporary server error", func(t *testing.T) {
			callCount := 0

			client := newTempServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertOK(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("http error", func(t *testing.T) {
			callCount := 0

			client := newHTTPErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertOK(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})
	})

	t.Run("max retries", func(t *testing.T) {
		callCount := 0

		client := newHTTPErrClient(func(req *http.Request) {
			callCount++

			b, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))
		})

		config := Config{
			Client:   client,
			Reporter: reporter,
		}

		req := NewRequest(config, http.MethodPost, "/url").
			WithText("test body").
			WithRetryPolicy(RetryAllErrors).
			WithMaxRetries(3).
			WithRetryDelay(0, 0)
		req.sleepFn = noopSleepFn
		req.chain.assertOK(t)

		resp := req.Expect().
			Status(http.StatusBadRequest)
		resp.chain.assertOK(t)

		// Should retry until max retries is reached
		assert.Equal(t, 1+3, callCount)
	})

	t.Run("retry delay", func(t *testing.T) {
		t.Run("not exceeding max retry delay", func(t *testing.T) {
			callCount := 0

			client := newHTTPErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			var totalSleepTime time.Duration

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(3).
				WithRetryDelay(100*time.Millisecond, 1000*time.Millisecond)
			req.sleepFn = func(d time.Duration) {
				totalSleepTime += d
			}
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertOK(t)

			// Should retry with delay
			assert.Equal(t, int64(100+200+400), totalSleepTime.Milliseconds())
		})

		t.Run("exceeding max retry delay", func(t *testing.T) {
			callCount := 0

			client := newHTTPErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			var totalSleepTime time.Duration

			req := NewRequest(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(3).
				WithRetryDelay(100*time.Millisecond, 300*time.Millisecond)
			req.sleepFn = func(d time.Duration) {
				totalSleepTime += d
			}
			req.chain.assertOK(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertOK(t)

			// Should retry with delay
			assert.Equal(t, int64(100+200+300), totalSleepTime.Milliseconds())
		})
	})
}
