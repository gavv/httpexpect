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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest_FailedChain(t *testing.T) {
	reporter := newMockReporter(t)
	chain := newChainWithDefaults("test", reporter)
	config := newMockConfig(reporter)

	chain.setFailed()

	req := newRequest(chain, config, "GET", "")
	req.chain.assertFailed(t)

	req.Alias("foo")
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
	resp.chain.assertFailed(t)
}

func TestRequest_Constructors(t *testing.T) {
	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		config := newMockConfig(reporter)
		req := NewRequestC(config, "GET", "")
		req.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		reporter := newMockReporter(t)
		config := newMockConfig(reporter)
		req := newRequest(chain, config, "GET", "")
		assert.NotSame(t, req.chain, chain)
		assert.Equal(t, req.chain.context.Path, chain.context.Path)
	})
}

func TestRequest_Reentrancy(t *testing.T) {
	t.Run("call from reporter", func(t *testing.T) {
		reporter := newMockReporter(t)

		config := Config{
			Client:   &mockClient{err: errors.New("test")},
			Reporter: reporter,
		}

		req := NewRequestC(config, "METHOD", "/")

		callCount := 0
		reporter.reportCb = func() {
			callCount++
			if callCount == 1 {
				req.WithName("test")
			}
		}

		resp := req.Expect()
		assert.Equal(t, 2, callCount)

		req.chain.assertFailed(t)
		resp.chain.assertFailed(t)
	})

	t.Run("call from client", func(t *testing.T) {
		client := &mockClient{}

		config := Config{
			Client:   client,
			Reporter: newMockReporter(t),
		}

		req := NewRequestC(config, "METHOD", "/")

		callCount := 0
		client.cb = func(_ *http.Request) {
			callCount++
			req.WithName("test")
		}

		resp := req.Expect()
		assert.Equal(t, 1, callCount)

		req.chain.assertFailed(t)
		resp.chain.assertNotFailed(t)
	})

	t.Run("call from transformer", func(t *testing.T) {
		config := Config{
			Client:   &mockClient{},
			Reporter: newMockReporter(t),
		}

		req := NewRequestC(config, "METHOD", "/")

		callCount := 0
		req.WithTransformer(func(_ *http.Request) {
			callCount++
			req.WithName("test")
		})

		resp := req.Expect()
		assert.Equal(t, 1, callCount)

		req.chain.assertFailed(t)
		resp.chain.assertNotFailed(t)
	})

	t.Run("call from matcher", func(t *testing.T) {
		config := Config{
			Client:   &mockClient{},
			Reporter: newMockReporter(t),
		}

		req := NewRequestC(config, "METHOD", "/")

		callCount := 0
		req.WithMatcher(func(_ *Response) {
			callCount++
			req.WithName("test")
		})

		resp := req.Expect()
		assert.Equal(t, 1, callCount)

		req.chain.assertFailed(t)
		resp.chain.assertNotFailed(t)
	})
}

func TestRequest_Alias(t *testing.T) {
	config := Config{
		Client:   &mockClient{},
		Reporter: newMockReporter(t),
	}

	value := NewRequestC(config, "GET", "")
	assert.Equal(t, []string{`Request("GET")`}, value.chain.context.Path)
	assert.Equal(t, []string{`Request("GET")`}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{`Request("GET")`}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestRequest_Basic(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		client := &mockClient{}

		config := Config{
			Client:   client,
			Reporter: newMockReporter(t),
		}

		req := NewRequestC(config, "GET", "/path")
		resp := req.Expect()

		req.chain.assertNotFailed(t)
		resp.chain.assertNotFailed(t)
	})

	t.Run("empty path", func(t *testing.T) {
		client := &mockClient{}

		config := Config{
			Client:   client,
			Reporter: newMockReporter(t),
		}

		req := NewRequestC(config, "GET", "")
		resp := req.Expect()

		req.chain.assertNotFailed(t)
		resp.chain.assertNotFailed(t)
	})

	t.Run("round trip time", func(t *testing.T) {
		client := &mockClient{}

		config := Config{
			Client:   client,
			Reporter: newMockReporter(t),
		}

		for n := 0; n < 10; n++ {
			req := NewRequestC(config, "GET", "/path")
			resp := req.Expect()
			require.NotNil(t, resp.rtt)
			assert.True(t, *resp.rtt >= 0)
		}
	})
}

func TestRequest_Matchers(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "METHOD", "/")

	var resps []*Response

	req.WithMatcher(func(r *Response) {
		resps = append(resps, r)
	})

	assert.Equal(t, 0, len(resps))

	resp := req.Expect()

	assert.Equal(t, 1, len(resps))
	assert.Same(t, resp, resps[0])
}

func TestRequest_Transformers(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("save ptr", func(t *testing.T) {
		var savedReq *http.Request
		transform := func(r *http.Request) {
			savedReq = r
		}

		req := NewRequestC(config, "METHOD", "/")
		req.WithTransformer(transform)
		req.Expect().chain.assertNotFailed(t)

		assert.NotNil(t, savedReq)
	})

	t.Run("append header", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "/")

		req.WithTransformer(func(r *http.Request) {
			r.Header.Add("foo", "11")
		})

		req.WithTransformer(func(r *http.Request) {
			r.Header.Add("bar", "22")
		})

		req.Expect().chain.assertNotFailed(t)

		assert.Equal(t, []string{"11"}, client.req.Header["Foo"])
		assert.Equal(t, []string{"22"}, client.req.Header["Bar"])
	})

	t.Run("append url", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "/{arg1}/{arg2}")

		req.WithPath("arg1", "11")
		req.WithPath("arg2", "22")

		req.WithTransformer(func(r *http.Request) {
			r.URL.Path += "/33"
		})

		req.WithTransformer(func(r *http.Request) {
			r.URL.Path += "/44"
		})

		req.Expect().chain.assertNotFailed(t)

		assert.Equal(t, "/11/22/33/44", client.req.URL.Path)
	})

	t.Run("nil func", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "/")
		req.WithTransformer(nil)
		req.chain.assertFailed(t)
	})
}

func TestRequest_Client(t *testing.T) {
	client1 := &mockClient{}
	client2 := &mockClient{}

	config := Config{
		Reporter: newMockReporter(t),
		Client:   client1,
	}

	var req *Request

	req = NewRequestC(config, "METHOD", "/")
	req.Expect().chain.assertNotFailed(t)
	assert.NotNil(t, client1.req)

	req = NewRequestC(config, "METHOD", "/")
	req.WithClient(client2)
	req.Expect().chain.assertNotFailed(t)
	assert.NotNil(t, client2.req)

	req = NewRequestC(config, "METHOD", "/")
	req.WithClient(nil)
	req.chain.assertFailed(t)
}

func TestRequest_Handler(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		var hr1 *http.Request
		handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hr1 = r
		})

		var hr2 *http.Request
		handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hr2 = r
		})

		config := Config{
			Reporter: reporter,
			Client: &http.Client{
				Transport: NewBinder(handler1),
			},
		}

		var req *Request

		req = NewRequestC(config, "METHOD", "/")
		req.Expect().chain.assertNotFailed(t)
		assert.NotNil(t, hr1)

		req = NewRequestC(config, "METHOD", "/")
		req.WithHandler(handler2)
		req.Expect().chain.assertNotFailed(t)
		assert.NotNil(t, hr2)
	})

	t.Run("nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		config := Config{
			Reporter: reporter,
		}

		req := NewRequestC(config, "METHOD", "/")
		req.WithHandler(nil)
		req.chain.assertFailed(t)
	})

	t.Run("reset client", func(t *testing.T) {
		reporter := newMockReporter(t)

		var hr *http.Request
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hr = r
		})

		client := &mockClient{}

		config := Config{
			Reporter: reporter,
			Client:   client,
		}

		req := NewRequestC(config, "METHOD", "/")
		req.WithHandler(handler)
		req.Expect().chain.assertNotFailed(t)
		assert.NotNil(t, hr)
		assert.Nil(t, client.req)
	})

	t.Run("reuse client", func(t *testing.T) {
		reporter := newMockReporter(t)

		handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		client := &http.Client{
			Transport: NewBinder(handler1),
			Jar:       NewCookieJar(),
		}

		config := Config{
			Reporter: reporter,
			Client:   client,
		}

		req := NewRequestC(config, "METHOD", "/")
		req.WithHandler(handler2)
		assert.Same(t, client.Jar, req.config.Client.(*http.Client).Jar)
	})
}

func TestRequest_Proto(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "METHOD", "/")

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

func TestRequest_URLConcatenate(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config1 := Config{
		BaseURL:  "",
		Client:   client,
		Reporter: reporter,
	}

	config2 := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	config3 := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs := []*Request{
		NewRequestC(config2, "METHOD", "path"),
		NewRequestC(config2, "METHOD", "/path"),
		NewRequestC(config3, "METHOD", "path"),
		NewRequestC(config3, "METHOD", "/path"),
		NewRequestC(config3, "METHOD", "{arg}", "/path"),
		NewRequestC(config3, "METHOD", "{arg}").WithPath("arg", "/path"),
	}

	for _, req := range reqs {
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, "http://example.com/path", client.req.URL.String())
	}

	empty1 := NewRequestC(config1, "METHOD", "")
	empty2 := NewRequestC(config2, "METHOD", "")
	empty3 := NewRequestC(config3, "METHOD", "")

	empty1.Expect().chain.assertNotFailed(t)
	empty2.Expect().chain.assertNotFailed(t)
	empty3.Expect().chain.assertNotFailed(t)

	assert.Equal(t, "", empty1.httpReq.URL.String())
	assert.Equal(t, "http://example.com", empty2.httpReq.URL.String())
	assert.Equal(t, "http://example.com/", empty3.httpReq.URL.String())
}

func TestRequest_URLOverwrite(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config1 := Config{
		BaseURL:  "",
		Client:   client,
		Reporter: reporter,
	}

	config2 := Config{
		BaseURL:  "http://foobar.com",
		Client:   client,
		Reporter: reporter,
	}

	reqs := []*Request{
		NewRequestC(config1, "METHOD", "/path").WithURL("http://example.com"),
		NewRequestC(config1, "METHOD", "path").WithURL("http://example.com"),
		NewRequestC(config1, "METHOD", "/path").WithURL("http://example.com/"),
		NewRequestC(config1, "METHOD", "path").WithURL("http://example.com/"),
		NewRequestC(config2, "METHOD", "/path").WithURL("http://example.com"),
		NewRequestC(config2, "METHOD", "path").WithURL("http://example.com"),
		NewRequestC(config2, "METHOD", "/path").WithURL("http://example.com/"),
		NewRequestC(config2, "METHOD", "path").WithURL("http://example.com/"),
	}

	for _, req := range reqs {
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, "http://example.com/path", client.req.URL.String())
	}
}

func TestRequest_URLInterpolate(t *testing.T) {
	client := &mockClient{}

	config := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: newMockReporter(t),
	}

	var reqs [3]*Request

	reqs[0] = NewRequestC(config, "METHOD", "/foo/{arg}", "bar")
	reqs[1] = NewRequestC(config, "METHOD", "{arg}foo{arg}", "/", "/bar")
	reqs[2] = NewRequestC(config, "METHOD", "{arg}", "/foo/bar")

	for _, req := range reqs {
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, "http://example.com/foo/bar", client.req.URL.String())
	}

	r1 := NewRequestC(config, "METHOD", "/{arg1}/{arg2}", "foo")
	r1.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/foo/%7Barg2%7D",
		client.req.URL.String())

	r2 := NewRequestC(config, "METHOD", "/{arg1}/{arg2}/{arg3}")
	r2.WithPath("ARG3", "foo")
	r2.WithPath("arg2", "bar")
	r2.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/%7Barg1%7D/bar/foo",
		client.req.URL.String())

	r3 := NewRequestC(config, "METHOD", "/{arg1}.{arg2}.{arg3}")
	r3.WithPath("arg2", "bar")
	r3.WithPathObject(map[string]string{"ARG1": "foo", "arg3": "baz"})
	r3.WithPathObject(nil)
	r3.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/foo.bar.baz",
		client.req.URL.String())

	type S struct {
		Arg1 string
		A2   int `path:"arg2"`
		A3   int `path:"-"`
	}

	r4 := NewRequestC(config, "METHOD", "/{arg1}{arg2}")
	r4.WithPathObject(S{"foo", 1, 2})
	r4.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/foo1", client.req.URL.String())

	r5 := NewRequestC(config, "METHOD", "/{arg1}{arg2}")
	r5.WithPathObject(&S{"foo", 1, 2})
	r5.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/foo1", client.req.URL.String())

	r6 := NewRequestC(config, "GET", "{arg}", nil)
	r6.chain.assertFailed(t)

	r7 := NewRequestC(config, "GET", "{arg}")
	r7.chain.assertNotFailed(t)
	r7.WithPath("arg", nil)
	r7.chain.assertFailed(t)

	r8 := NewRequestC(config, "GET", "{arg}")
	r8.chain.assertNotFailed(t)
	r8.WithPath("bad", "value")
	r8.chain.assertFailed(t)

	r9 := NewRequestC(config, "GET", "{arg")
	r9.chain.assertFailed(t)
	r9.WithPath("arg", "foo")
	r9.chain.assertFailed(t)

	r10 := NewRequestC(config, "GET", "{arg}")
	r10.chain.assertNotFailed(t)
	r10.WithPathObject(func() {})
	r10.chain.assertFailed(t)
}

func TestRequest_URLQuery(t *testing.T) {
	client := &mockClient{}

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req1 := NewRequestC(config, "METHOD", "/path").
		WithQuery("aa", "foo").WithQuery("bb", 123).WithQuery("cc", "*&@")

	q := map[string]interface{}{
		"bb": 123,
		"cc": "*&@",
	}

	req2 := NewRequestC(config, "METHOD", "/path").
		WithQuery("aa", "foo").
		WithQueryObject(q)

	type S struct {
		Bb int    `url:"bb"`
		Cc string `url:"cc"`
		Dd string `url:"-"`
	}

	req3 := NewRequestC(config, "METHOD", "/path").
		WithQueryObject(S{123, "*&@", "dummy"}).WithQuery("aa", "foo")

	req4 := NewRequestC(config, "METHOD", "/path").
		WithQueryObject(&S{123, "*&@", "dummy"}).WithQuery("aa", "foo")

	req5 := NewRequestC(config, "METHOD", "/path").
		WithQuery("bb", 123).
		WithQueryString("aa=foo&cc=%2A%26%40")

	req6 := NewRequestC(config, "METHOD", "/path").
		WithQueryString("aa=foo&cc=%2A%26%40").
		WithQuery("bb", 123)

	for _, req := range []*Request{req1, req2, req3, req4, req5, req6} {
		client.req = nil
		req.Expect()
		req.chain.assertNotFailed(t)
		assert.Equal(t, "http://example.com/path?aa=foo&bb=123&cc=%2A%26%40",
			client.req.URL.String())
	}

	req7 := NewRequestC(config, "METHOD", "/path").
		WithQuery("foo", "bar").
		WithQueryObject(nil)

	req7.Expect()
	req7.chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/path?foo=bar", client.req.URL.String())

	NewRequestC(config, "METHOD", "/path").
		WithQueryObject(func() {}).chain.assertFailed(t)

	NewRequestC(config, "METHOD", "/path").
		WithQueryString("%").chain.assertFailed(t)
}

func TestRequest_Headers(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "METHOD", "url")

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
	resp.chain.assertNotFailed(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "example.com", client.req.Host)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequest_Cookies(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "METHOD", "url")

	req.WithCookie("foo", "1")
	req.WithCookie("bar", "2 ")

	req.WithCookies(map[string]string{
		"baz": " 3",
	})

	expectedHeaders := map[string][]string{
		"Cookie": {`foo=1; bar="2 "; baz=" 3"`},
	}

	resp := req.Expect()
	resp.chain.assertNotFailed(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequest_BasicAuth(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "METHOD", "url")

	req.WithBasicAuth("Aladdin", "open sesame")
	req.chain.assertNotFailed(t)

	assert.Equal(t, "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==",
		req.httpReq.Header.Get("Authorization"))
}

func TestRequest_Host(t *testing.T) {
	client1 := &mockClient{}
	reporter1 := newMockReporter(t)

	config1 := Config{
		Client:   client1,
		Reporter: reporter1,
	}

	req1 := NewRequestC(config1, "METHOD", "url")

	req1.WithHost("example.com")

	resp := req1.Expect()
	resp.chain.assertNotFailed(t)

	assert.Equal(t, "METHOD", client1.req.Method)
	assert.Equal(t, "example.com", client1.req.Host)
	assert.Equal(t, "url", client1.req.URL.String())

	assert.Same(t, &client1.resp, resp.Raw())

	client2 := &mockClient{}
	reporter2 := newMockReporter(t)

	config2 := Config{
		Client:   client2,
		Reporter: reporter2,
	}

	req2 := NewRequestC(config2, "METHOD", "url")

	req2.WithHeader("HOST", "example1.com")
	req2.WithHost("example2.com")

	req2.Expect().chain.assertNotFailed(t)

	assert.Equal(t, "example2.com", client2.req.Host)

	client3 := &mockClient{}
	reporter3 := newMockReporter(t)

	config3 := Config{
		Client:   client3,
		Reporter: reporter3,
	}

	req3 := NewRequestC(config3, "METHOD", "url")

	req3.WithHost("example2.com")
	req3.WithHeader("HOST", "example1.com")

	req3.Expect().chain.assertNotFailed(t)

	assert.Equal(t, "example1.com", client3.req.Host)
}

func TestRequest_BodyChunked(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("body", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "url")

		req.WithChunked(bytes.NewBufferString("body"))

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.NotNil(t, client.req.Body)
		assert.Equal(t, int64(-1), client.req.ContentLength)

		assert.Equal(t, "METHOD", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, make(http.Header), client.req.Header)
		assert.Equal(t, "body", resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("nil", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "url")

		req.WithChunked(nil)

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, http.NoBody, client.req.Body)
		assert.Equal(t, int64(0), client.req.ContentLength)
	})

	t.Run("proto", func(t *testing.T) {
		req1 := NewRequestC(config, "METHOD", "url")

		req1.WithProto("HTTP/1.0")
		assert.Equal(t, 1, req1.httpReq.ProtoMajor)
		assert.Equal(t, 0, req1.httpReq.ProtoMinor)

		req1.WithChunked(bytes.NewBufferString("body"))
		req1.chain.assertFailed(t)

		req2 := NewRequestC(config, "METHOD", "url")

		req2.WithProto("HTTP/2.0")
		assert.Equal(t, 2, req2.httpReq.ProtoMajor)
		assert.Equal(t, 0, req2.httpReq.ProtoMinor)

		req2.WithChunked(bytes.NewBufferString("body"))
		assert.Equal(t, 2, req2.httpReq.ProtoMajor)
		assert.Equal(t, 0, req2.httpReq.ProtoMinor)
	})
}

func TestRequest_BodyBytes(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("byte slice", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "/path")

		req.WithBytes([]byte("body"))

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.NotNil(t, client.req.Body)
		assert.Equal(t, int64(len("body")), client.req.ContentLength)

		assert.Equal(t, "METHOD", client.req.Method)
		assert.Equal(t, "/path", client.req.URL.String())
		assert.Equal(t, make(http.Header), client.req.Header)
		assert.Equal(t, "body", resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("nil", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "url")

		req.WithBytes(nil)

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, http.NoBody, client.req.Body)
		assert.Equal(t, int64(0), client.req.ContentLength)
	})
}

func TestRequest_BodyText(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"text/plain; charset=utf-8"},
		"Some-Header":  {"foo"},
	}

	req := NewRequestC(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithText("some text")

	resp := req.Expect()
	resp.chain.assertNotFailed(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, "some text", resp.Body().Raw())

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequest_BodyForm(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("form", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
			"Some-Header":  {"foo"},
		}

		req := NewRequestC(config, "METHOD", "url")

		req.WithHeaders(map[string]string{
			"Some-Header": "foo",
		})

		req.WithForm(map[string]interface{}{
			"a": 1,
			"b": "2",
		})

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "METHOD", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("form field", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
			"Some-Header":  {"foo"},
		}

		req := NewRequestC(config, "METHOD", "url")

		req.WithHeaders(map[string]string{
			"Some-Header": "foo",
		})

		req.WithFormField("a", 1)
		req.WithFormField("b", "2")

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "METHOD", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("form struct", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		req := NewRequestC(config, "METHOD", "url")

		type S struct {
			A string `form:"a"`
			B int    `form:"b"`
			C int    `form:"-"`
		}

		req.WithForm(S{"1", 2, 3})

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "METHOD", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("form combined", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		req := NewRequestC(config, "METHOD", "url")

		type S struct {
			A int `form:"a"`
		}

		req.WithForm(S{A: 1})
		req.WithForm(map[string]string{"b": "2"})
		req.WithFormField("c", 3)

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "METHOD", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2&c=3`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})
}

func TestRequest_BodyMultipart(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("multipart", func(t *testing.T) {
		req := NewRequestC(config, "POST", "url")

		req.WithMultipart()
		req.WithForm(map[string]string{"b": "1", "c": "2"})
		req.WithFormField("a", 3)

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "POST", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())

		mediatype, params, err := mime.ParseMediaType(client.req.Header.Get("Content-Type"))

		assert.NoError(t, err)
		assert.Equal(t, "multipart/form-data", mediatype)
		assert.True(t, params["boundary"] != "")

		reader := multipart.NewReader(strings.NewReader(resp.Body().Raw()),
			params["boundary"])

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
		assert.Nil(t, eof)
	})

	t.Run("multipart file", func(t *testing.T) {
		req := NewRequestC(config, "POST", "url")

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
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "POST", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())

		mediatype, params, err := mime.ParseMediaType(client.req.Header.Get("Content-Type"))

		assert.NoError(t, err)
		assert.Equal(t, "multipart/form-data", mediatype)
		assert.True(t, params["boundary"] != "")

		reader := multipart.NewReader(strings.NewReader(resp.Body().Raw()),
			params["boundary"])

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
		assert.Nil(t, eof)
	})
}

func TestRequest_BodyJSON(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/json; charset=utf-8"},
		"Some-Header":  {"foo"},
	}

	req := NewRequestC(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithJSON(map[string]interface{}{"key": "value"})

	resp := req.Expect()
	resp.chain.assertNotFailed(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `{"key":"value"}`, resp.Body().Raw())

	assert.Same(t, &client.resp, resp.Raw())
}

func TestRequest_ContentLength(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req1 := NewRequestC(config, "METHOD", "url")
	req1.WithChunked(bytes.NewReader([]byte("12345")))
	req1.Expect().chain.assertNotFailed(t)
	assert.Equal(t, int64(-1), client.req.ContentLength)

	req2 := NewRequestC(config, "METHOD", "url")
	req2.WithBytes([]byte("12345"))
	req2.Expect().chain.assertNotFailed(t)
	assert.Equal(t, int64(5), client.req.ContentLength)

	req3 := NewRequestC(config, "METHOD", "url")
	req3.WithText("12345")
	req3.Expect().chain.assertNotFailed(t)
	assert.Equal(t, int64(5), client.req.ContentLength)

	j, _ := json.Marshal(map[string]string{"a": "b"})
	req4 := NewRequestC(config, "METHOD", "url")
	req4.WithJSON(map[string]string{"a": "b"})
	req4.Expect().chain.assertNotFailed(t)
	assert.Equal(t, int64(len(j)), client.req.ContentLength)

	f := `a=b`
	req5 := NewRequestC(config, "METHOD", "url")
	req5.WithForm(map[string]string{"a": "b"})
	req5.Expect().chain.assertNotFailed(t)
	assert.Equal(t, int64(len(f)), client.req.ContentLength)

	req6 := NewRequestC(config, "METHOD", "url")
	req6.WithFormField("a", "b")
	req6.Expect().chain.assertNotFailed(t)
	assert.Equal(t, int64(len(f)), client.req.ContentLength)

	req7 := NewRequestC(config, "METHOD", "url")
	req7.WithMultipart()
	req7.WithFileBytes("a", "b", []byte("12345"))
	req7.Expect().chain.assertNotFailed(t)
	assert.True(t, client.req.ContentLength > 0)
}

func TestRequest_ContentType(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req1 := NewRequestC(config, "METHOD", "url")
	req1.WithText("hello")
	req1.WithHeader("Content-Type", "foo")
	req1.Expect().chain.assertNotFailed(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)

	req2 := NewRequestC(config, "METHOD", "url")
	req2.WithHeader("Content-Type", "foo")
	req2.WithText("hello")
	req2.Expect().chain.assertNotFailed(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)

	req3 := NewRequestC(config, "METHOD", "url")
	req3.WithJSON(map[string]interface{}{"a": "b"})
	req3.WithHeader("Content-Type", "foo")
	req3.WithHeader("Content-Type", "bar")
	req3.Expect().chain.assertNotFailed(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)

	req4 := NewRequestC(config, "METHOD", "url")
	req4.WithForm(map[string]interface{}{"a": "b"})
	req4.WithHeader("Content-Type", "foo")
	req4.WithHeader("Content-Type", "bar")
	req4.Expect().chain.assertNotFailed(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)

	req5 := NewRequestC(config, "METHOD", "url")
	req5.WithMultipart()
	req5.WithForm(map[string]interface{}{"a": "b"})
	req5.WithHeader("Content-Type", "foo")
	req5.WithHeader("Content-Type", "bar")
	req5.Expect().chain.assertNotFailed(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)
}

func TestRequest_Errors(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("error marshal form", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "url")

		req.WithForm(func() {})

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
	})

	t.Run("error marshalJSON", func(t *testing.T) {
		req := NewRequestC(config, "METHOD", "url")

		req.WithJSON(func() {})

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
	})

	t.Run("error read file", func(t *testing.T) {
		client.err = errors.New("error")

		req := NewRequestC(config, "METHOD", "url")

		req.WithMultipart()
		req.WithFile("", "")

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
	})

	t.Run("error send", func(t *testing.T) {
		client.err = errors.New("error")

		req := NewRequestC(config, "METHOD", "url")

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
	})
}

func TestRequest_Conflicts(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("body conflict", func(t *testing.T) {
		req1 := NewRequestC(config, "METHOD", "url")
		req1.WithChunked(nil)
		req1.chain.assertNotFailed(t)
		req1.WithChunked(nil)
		req1.chain.assertFailed(t)

		req2 := NewRequestC(config, "METHOD", "url")
		req2.WithChunked(nil)
		req2.chain.assertNotFailed(t)
		req2.WithBytes(nil)
		req2.chain.assertFailed(t)

		req3 := NewRequestC(config, "METHOD", "url")
		req3.WithChunked(nil)
		req3.chain.assertNotFailed(t)
		req3.WithText("")
		req3.chain.assertFailed(t)

		req4 := NewRequestC(config, "METHOD", "url")
		req4.WithChunked(nil)
		req4.chain.assertNotFailed(t)
		req4.WithJSON(map[string]interface{}{"a": "b"})
		req4.chain.assertFailed(t)

		req5 := NewRequestC(config, "METHOD", "url")
		req5.WithChunked(nil)
		req5.chain.assertNotFailed(t)
		req5.WithForm(map[string]interface{}{"a": "b"})
		req5.Expect()
		req5.chain.assertFailed(t)

		req6 := NewRequestC(config, "METHOD", "url")
		req6.WithChunked(nil)
		req6.chain.assertNotFailed(t)
		req6.WithFormField("a", "b")
		req6.Expect()
		req6.chain.assertFailed(t)

		req7 := NewRequestC(config, "METHOD", "url")
		req7.WithChunked(nil)
		req7.chain.assertNotFailed(t)
		req7.WithMultipart()
		req7.chain.assertFailed(t)
	})

	t.Run("type conflict", func(t *testing.T) {
		req1 := NewRequestC(config, "METHOD", "url")
		req1.WithText("")
		req1.chain.assertNotFailed(t)
		req1.WithJSON(map[string]interface{}{"a": "b"})
		req1.chain.assertFailed(t)

		req2 := NewRequestC(config, "METHOD", "url")
		req2.WithText("")
		req2.chain.assertNotFailed(t)
		req2.WithForm(map[string]interface{}{"a": "b"})
		req2.chain.assertFailed(t)

		req3 := NewRequestC(config, "METHOD", "url")
		req3.WithText("")
		req3.chain.assertNotFailed(t)
		req3.WithFormField("a", "b")
		req3.chain.assertFailed(t)

		req4 := NewRequestC(config, "METHOD", "url")
		req4.WithText("")
		req4.chain.assertNotFailed(t)
		req4.WithMultipart()
		req4.chain.assertFailed(t)
	})

	t.Run("multipart conflict", func(t *testing.T) {
		req1 := NewRequestC(config, "METHOD", "url")
		req1.WithForm(map[string]interface{}{"a": "b"})
		req1.chain.assertNotFailed(t)
		req1.WithMultipart()
		req1.chain.assertFailed(t)

		req2 := NewRequestC(config, "METHOD", "url")
		req2.WithFormField("a", "b")
		req2.chain.assertNotFailed(t)
		req2.WithMultipart()
		req2.chain.assertFailed(t)

		req3 := NewRequestC(config, "METHOD", "url")
		req3.WithFileBytes("a", "a", []byte("a"))
		req3.chain.assertFailed(t)
	})
}

func TestRequest_Redirects(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("DontFollowRedirects", func(t *testing.T) {
		t.Run("no body", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(DontFollowRedirects)
			req.chain.assertNotFailed(t)

			// Should return redirection response
			resp := req.Expect().
				Status(tp.redirectHTTPStatusCode).
				Header("Location").
				IsEqual("/redirect")
			resp.chain.assertNotFailed(t)

			// Should set GetBody
			assert.Nil(t, req.httpReq.GetBody)

			// Should set CheckRedirect
			httpClient, _ := req.config.Client.(*http.Client)
			assert.NotNil(t, httpClient.CheckRedirect)
			assert.Equal(t, http.ErrUseLastResponse, httpClient.CheckRedirect(req.httpReq, nil))

			// Should do round trip
			assert.Equal(t, 1, tp.tripCount)
		})

		t.Run("has body", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(DontFollowRedirects).
				WithText("test body")
			req.chain.assertNotFailed(t)

			// Should return redirection response
			resp := req.Expect().
				Status(tp.redirectHTTPStatusCode).
				Header("Location").
				IsEqual("/redirect")
			resp.chain.assertNotFailed(t)

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

	t.Run("FollowAllRedirects", func(t *testing.T) {
		t.Run("no body", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1)
			req.chain.assertNotFailed(t)

			// Should return OK response
			resp := req.Expect().
				Status(http.StatusOK)
			resp.chain.assertNotFailed(t)

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

		t.Run("no body, too many redirects", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1)
			req.chain.assertNotFailed(t)

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

		t.Run("has body", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1).
				WithText("test body")
			req.chain.assertNotFailed(t)

			// Should return OK response
			resp := req.Expect().
				Status(http.StatusOK)
			resp.chain.assertNotFailed(t)

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

		t.Run("has body, too many redirects", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowAllRedirects).
				WithMaxRedirects(1).
				WithText("test body")
			req.chain.assertNotFailed(t)

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

	t.Run("FollowRedirectsWithoutBody", func(t *testing.T) {
		t.Run("no body", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowRedirectsWithoutBody).
				WithMaxRedirects(1)
			req.chain.assertNotFailed(t)

			// Should return OK response
			resp := req.Expect().
				Status(http.StatusOK)
			resp.chain.assertNotFailed(t)

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

		t.Run("no body, too many redirects", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowRedirectsWithoutBody).
				WithMaxRedirects(1)
			req.chain.assertNotFailed(t)

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

		t.Run("has body, status permanent redirect", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowRedirectsWithoutBody).
				WithMaxRedirects(1).
				WithText("test body")
			req.chain.assertNotFailed(t)

			// Should return redirection response
			resp := req.Expect().
				Status(tp.redirectHTTPStatusCode)
			resp.chain.assertNotFailed(t)

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

		t.Run("has body, status moved permanently",
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
					}).
					WithRedirectHTTPStatusCode(http.StatusMovedPermanently).
					WithMaxRedirect(1)

				client := &http.Client{
					Transport: tp,
				}

				config := Config{
					Client:   client,
					Reporter: reporter,
				}

				req := NewRequestC(config, http.MethodPut, "/url").
					WithRedirectPolicy(FollowRedirectsWithoutBody).
					WithMaxRedirects(1).
					WithText("test body")
				req.chain.assertNotFailed(t)

				// Should return OK response
				resp := req.Expect().
					Status(http.StatusOK)
				resp.chain.assertNotFailed(t)

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

		t.Run("has body, status moved permanently, too many redirects", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPut, "/url").
				WithRedirectPolicy(FollowRedirectsWithoutBody).
				WithMaxRedirects(1).
				WithText("test body")
			req.chain.assertNotFailed(t)

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

func TestRequest_Retries(t *testing.T) {
	reporter := newMockReporter(t)

	newNoErrClient := func(cb func(req *http.Request)) *mockClient {
		return &mockClient{
			resp: http.Response{
				StatusCode: http.StatusOK,
			},
			cb: cb,
		}
	}

	newTimeoutErrClient := func(cb func(req *http.Request)) *mockClient {
		return &mockClient{
			err: &mockNetError{
				isTimeout: true,
			},
			cb: cb,
		}
	}

	newServerErrClient := func(cb func(req *http.Request)) *mockClient {
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

	noopSleepFn := func(time.Duration) <-chan time.Time {
		return time.After(0)
	}

	t.Run("DontRetry", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("timeout error", func(t *testing.T) {
			callCount := 0

			client := newTimeoutErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("server error", func(t *testing.T) {
			callCount := 0

			client := newServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertNotFailed(t)

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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(DontRetry).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

	})

	t.Run("RetryTimeoutErrors", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutErrors)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("timeout error", func(t *testing.T) {
			callCount := 0

			client := newTimeoutErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("server error", func(t *testing.T) {
			callCount := 0

			client := newServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutErrors).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertNotFailed(t)

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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutErrors).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})
	})

	t.Run("RetryTimeoutAndServerErrors", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutAndServerErrors)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("timeout error", func(t *testing.T) {
			callCount := 0

			client := newTimeoutErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutAndServerErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("server error", func(t *testing.T) {
			callCount := 0

			client := newServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutAndServerErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertNotFailed(t)

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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryTimeoutAndServerErrors).
				WithMaxRetries(1)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})
	})

	t.Run("RetryAllErrors", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertNotFailed(t)

			// Should not retry
			assert.Equal(t, 1, callCount)
		})

		t.Run("timeout error", func(t *testing.T) {
			callCount := 0

			client := newTimeoutErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			resp.chain.assertFailed(t)

			// Should retry
			assert.Equal(t, 2, callCount)
		})

		t.Run("server error", func(t *testing.T) {
			callCount := 0

			client := newServerErrClient(func(req *http.Request) {
				callCount++

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, "test body", string(b))
			})

			config := Config{
				Client:   client,
				Reporter: reporter,
			}

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusInternalServerError)
			resp.chain.assertNotFailed(t)

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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(1).
				WithRetryDelay(0, 0)
			req.sleepFn = noopSleepFn
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertNotFailed(t)

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

		req := NewRequestC(config, http.MethodPost, "/url").
			WithText("test body").
			WithRetryPolicy(RetryAllErrors).
			WithMaxRetries(3).
			WithRetryDelay(0, 0)
		req.sleepFn = noopSleepFn
		req.chain.assertNotFailed(t)

		resp := req.Expect().
			Status(http.StatusBadRequest)
		resp.chain.assertNotFailed(t)

		// Should retry until max retries is reached
		assert.Equal(t, 1+3, callCount)
	})

	t.Run("retry delay", func(t *testing.T) {
		t.Run("not exceeded", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(3).
				WithRetryDelay(100*time.Millisecond, 1000*time.Millisecond)
			req.sleepFn = func(d time.Duration) <-chan time.Time {
				totalSleepTime += d
				return time.After(0)
			}
			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertNotFailed(t)

			// Should retry with delay
			assert.Equal(t, int64(100+200+400), totalSleepTime.Milliseconds())
		})

		t.Run("exceeded", func(t *testing.T) {
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

			req := NewRequestC(config, http.MethodPost, "/url").
				WithText("test body").
				WithRetryPolicy(RetryAllErrors).
				WithMaxRetries(3).
				WithRetryDelay(100*time.Millisecond, 300*time.Millisecond)
			req.sleepFn = func(d time.Duration) <-chan time.Time {
				totalSleepTime += d
				return time.After(0)
			}

			req.chain.assertNotFailed(t)

			resp := req.Expect().
				Status(http.StatusBadRequest)
			resp.chain.assertNotFailed(t)

			// Should retry with delay
			assert.Equal(t, int64(100+200+300), totalSleepTime.Milliseconds())
		})
	})

	t.Run("cancelled retries", func(t *testing.T) {
		callCount := 0

		client := newHTTPErrClient(func(req *http.Request) {
			callCount++

			assert.Error(t, req.Context().Err(), context.Canceled.Error())

			b, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))
		})

		config := Config{
			Client:   client,
			Reporter: reporter,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately to trigger error

		req := NewRequestC(config, http.MethodPost, "/url").
			WithText("test body").
			WithRetryPolicy(RetryAllErrors).
			WithMaxRetries(1).
			WithContext(ctx).
			WithRetryDelay(1*time.Minute, 5*time.Minute)
		req.chain.assertNotFailed(t)

		resp := req.Expect()
		resp.chain.assertFailed(t)

		// Should not retry
		assert.Equal(t, 1, callCount)
	})
}

func TestRequest_Usage(t *testing.T) {
	tests := []struct {
		name        string
		client      Client
		prepFunc    func(req *Request)
		prepFails   bool
		expectFails bool
	}{
		{
			name:   "with matcher",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithMatcher(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with transformer",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithTransformer(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with client",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithClient(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with handler",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithHandler(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with context",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithContext(nil) //nolint
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "max redirects method fail",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithMaxRedirects(-1)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with max retries",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithMaxRetries(-1)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with retry delay",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithRetryDelay(10, 5)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with websocket dialer",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithWebsocketDialer(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with path",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithPath("test-path", nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with query",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithPath("test-query", nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with url",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithURL("%-invalid-url")
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name:   "with file",
			client: nil,
			prepFunc: func(req *Request) {
				req.WithFile("test-key", "test-path", nil, nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		// WithRedirectPolicy and WithMaxRedirects requires
		// Client to be http.Client, but we use another one.
		{
			name:   "with redirect policy",
			client: &mockClient{},
			prepFunc: func(req *Request) {
				req.WithRedirectPolicy(FollowAllRedirects)
			},
			prepFails:   false,
			expectFails: true,
		},
		{
			name:   "max redirects Expect() fail",
			client: &mockClient{},
			prepFunc: func(req *Request) {
				req.WithMaxRedirects(1)
			},
			prepFails:   false,
			expectFails: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Client:   tt.client,
				Reporter: newMockReporter(t),
			}

			req := NewRequestC(config, "METHOD", "/")

			if !(tt.prepFails && tt.expectFails) {
				tt.prepFunc(req)
				req.Expect().chain.assertFailed(t)
			} else {
				tt.prepFunc(req)
				req.chain.assertFailed(t)
			}
		})
	}
}

func TestRequest_Order(t *testing.T) {
	tests := []struct {
		name         string
		prepFunc     func(req *Request) *Request
		expectFunc   func(req *Request) *Response
		reqFunc      func(req *Request)
		expectSame   bool
		expectNotNil bool
	}{
		{
			name:     "Expect after Expect",
			prepFunc: nil,
			expectFunc: func(req *Request) *Response {
				return req.Expect()
			},
			reqFunc:      nil,
			expectSame:   false,
			expectNotNil: true,
		},
		{
			name: "WithName after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithName("Test")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithMatcher after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithMatcher(func(resp *Response) {
					resp.Header("API-Version").NotEmpty()
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithTransformer after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithTransformer(func(r *http.Request) {
					r.Header.Add("foo", "bar")
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithClient after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithClient(&mockClient{})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithHandler after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithHandler(http.NotFoundHandler())
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithContext after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithContext(context.Background())
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithTimeout after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithTimeout(3 * time.Second)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithRedirectPolicy after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithRedirectPolicy(FollowAllRedirects)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithMaxRedirects after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithMaxRedirects(3)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithRetryPolicy after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithRetryPolicy(DontRetry)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithMaxRetries after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithMaxRetries(10)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithRetryDelay after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithRetryDelay(time.Second, 5*time.Second)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithWebsocketUpgrade after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithWebsocketUpgrade()
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithWebsocketDialer after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithWebsocketDialer(&websocket.Dialer{})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithPath after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithPath("repo", "repo1")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithPathObject after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithPathObject(map[string]string{
					"repo": "repo1",
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithQuery after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithQuery("a", 123)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithQueryObject after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithQueryObject(map[string]string{
					"a": "val",
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithQueryString after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithQueryString("a=123&b=hello")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithURL after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithURL("https://www.github.com")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithHeaders after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithHeaders(map[string]string{
					"Content-Type": "application/json",
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithHeader after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithHeader("Content-Type", "application/json")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithCookies after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithCookies(map[string]string{
					"key1": "val1",
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithCookie after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithCookie("key1", "val1")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithBasicAuth after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithBasicAuth("user", "pass")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithHost after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithHost("localhost")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithProto after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithProto("HTTP/1.1")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithChunked after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithChunked(bytes.NewReader(nil))
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithBytes after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithBytes(nil)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithText after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithText("hello")
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithJSON after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithJSON(map[string]string{
					"key1": "val1",
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithForm after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithForm(map[string]string{
					"key1": "val1",
				})
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithFormField after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithFormField("key1", 123)
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithFile after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithFile("foo", "bar", strings.NewReader("baz"))
			},
			expectFunc: nil,
			reqFunc: func(req *Request) {
				req.WithMultipart()
			},
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithFileBytes after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithFileBytes("foo", "bar", []byte("baz"))
			},
			expectFunc: nil,
			reqFunc: func(req *Request) {
				req.WithMultipart()
			},
			expectSame:   true,
			expectNotNil: false,
		},
		{
			name: "WithMultipart after Expect",
			prepFunc: func(req *Request) *Request {
				return req.WithMultipart()
			},
			expectFunc:   nil,
			reqFunc:      nil,
			expectSame:   true,
			expectNotNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Client:   &mockClient{},
				Reporter: newMockReporter(t),
			}

			req := NewRequestC(config, "GET", "/")

			if tt.reqFunc != nil {
				tt.reqFunc(req)
			}

			req.Expect()

			if tt.expectSame {
				assert.Same(t, req, tt.prepFunc(req))
			}

			if tt.expectNotNil {
				assert.NotNil(t, tt.expectFunc(req))
			}

			req.chain.assertFailed(t)
		})
	}
}

func TestRequest_Panics(t *testing.T) {
	t.Run("RequestFactory is nil", func(t *testing.T) {
		config := Config{
			RequestFactory:   nil,
			Client:           &mockClient{},
			AssertionHandler: &mockAssertionHandler{},
		}

		assert.Panics(t, func() { newRequest(newMockChain(t), config, "METHOD", "") })
	})

	t.Run("Client is nil", func(t *testing.T) {
		config := Config{
			RequestFactory:   DefaultRequestFactory{},
			Client:           nil,
			AssertionHandler: &mockAssertionHandler{},
		}

		assert.Panics(t, func() { newRequest(newMockChain(t), config, "METHOD", "") })
	})

	t.Run("AssertionHandler is nil", func(t *testing.T) {
		config := Config{
			RequestFactory:   DefaultRequestFactory{},
			Client:           &mockClient{},
			AssertionHandler: nil,
		}

		assert.Panics(t, func() { newRequest(newMockChain(t), config, "METHOD", "") })
	})
}
