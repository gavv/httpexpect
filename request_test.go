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
	neturl "net/url"
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
	req.WithName("foo")
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

		req := NewRequestC(config, "GET", "/")

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

		req := NewRequestC(config, "GET", "/")

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

		req := NewRequestC(config, "GET", "/")

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

		req := NewRequestC(config, "GET", "/")

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

	t.Run("client error", func(t *testing.T) {
		client := &mockClient{
			err: errors.New("error"),
		}

		config := Config{
			Client:   client,
			Reporter: newMockReporter(t),
		}

		req := NewRequestC(config, "GET", "url")

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
	})
}

func TestRequest_Matchers(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "GET", "/")

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

		req := NewRequestC(config, "GET", "/")
		req.WithTransformer(transform)
		req.Expect().chain.assertNotFailed(t)

		assert.NotNil(t, savedReq)
	})

	t.Run("append header", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/")

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
		req := NewRequestC(config, "GET", "/{arg1}/{arg2}")

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
		req := NewRequestC(config, "GET", "/")
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

	req = NewRequestC(config, "GET", "/")
	req.Expect().chain.assertNotFailed(t)
	assert.NotNil(t, client1.req)

	req = NewRequestC(config, "GET", "/")
	req.WithClient(client2)
	req.Expect().chain.assertNotFailed(t)
	assert.NotNil(t, client2.req)

	req = NewRequestC(config, "GET", "/")
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

		req = NewRequestC(config, "GET", "/")
		req.Expect().chain.assertNotFailed(t)
		assert.NotNil(t, hr1)

		req = NewRequestC(config, "GET", "/")
		req.WithHandler(handler2)
		req.Expect().chain.assertNotFailed(t)
		assert.NotNil(t, hr2)
	})

	t.Run("nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		config := Config{
			Reporter: reporter,
		}

		req := NewRequestC(config, "GET", "/")
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

		req := NewRequestC(config, "GET", "/")
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

		req := NewRequestC(config, "GET", "/")
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

	req := NewRequestC(config, "GET", "/")

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
	cases := []struct {
		name        string
		baseURL     string
		method      string
		path        string
		pathArgs    []interface{}
		setupFunc   func(req *Request)
		expectedURL string
	}{
		{
			name:        "empty url, empty path",
			baseURL:     "",
			method:      "GET",
			path:        "",
			expectedURL: "",
		},
		{
			name:        "empty path",
			baseURL:     "http://example.com",
			method:      "GET",
			path:        "",
			expectedURL: "http://example.com",
		},
		{
			name:        "url with slash, empty path",
			baseURL:     "http://example.com/",
			method:      "GET",
			path:        "",
			expectedURL: "http://example.com/",
		},
		{
			name:        "url with path",
			baseURL:     "http://example.com",
			method:      "GET",
			path:        "path",
			expectedURL: "http://example.com/path",
		},
		{
			name:        "url with path, path without slash",
			baseURL:     "http://example.com",
			method:      "GET",
			path:        "path",
			expectedURL: "http://example.com/path",
		},
		{
			name:        "url with path, path with slash",
			baseURL:     "http://example.com",
			method:      "GET",
			path:        "/path",
			expectedURL: "http://example.com/path",
		},
		{
			name:        "url with slash and path, path without slash",
			baseURL:     "http://example.com/",
			method:      "GET",
			path:        "path",
			expectedURL: "http://example.com/path",
		},
		{
			name:        "url with slash and path, path with slash",
			baseURL:     "http://example.com/",
			method:      "GET",
			path:        "/path",
			expectedURL: "http://example.com/path",
		},
		{
			name:        "url with path arg",
			baseURL:     "http://example.com/",
			method:      "GET",
			path:        "{arg}",
			pathArgs:    []interface{}{"/path"},
			expectedURL: "http://example.com/path",
		},
		{
			name:    "url with arg setup func",
			baseURL: "http://example.com/",
			method:  "GET",
			path:    "{arg}",
			setupFunc: func(req *Request) {
				req.WithPath("arg", "/path")
			},
			expectedURL: "http://example.com/path",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockClient{}
			reporter := NewAssertReporter(t)

			req := NewRequestC(
				Config{
					BaseURL:  tc.baseURL,
					Client:   client,
					Reporter: reporter,
				},
				tc.method,
				tc.path,
				tc.pathArgs...)

			if tc.setupFunc != nil {
				tc.setupFunc(req)
			}

			req.Expect().chain.assertNotFailed(t)
			req.chain.assertNotFailed(t)
			assert.Equal(t, tc.expectedURL, req.httpReq.URL.String())

		})
	}
}

func TestRequest_URLOverwrite(t *testing.T) {
	cases := []struct {
		name         string
		baseURL      string
		method       string
		path         string
		overwriteURL string
		expectedURL  string
	}{
		{
			name:         "without slash on url",
			baseURL:      "",
			method:       "GET",
			path:         "/path",
			overwriteURL: "http://example.com",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "without slash on url and path",
			baseURL:      "",
			method:       "GET",
			path:         "path",
			overwriteURL: "http://example.com",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "with slash on url and path",
			baseURL:      "",
			method:       "GET",
			path:         "/path",
			overwriteURL: "http://example.com/",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "without slash on path",
			baseURL:      "",
			method:       "GET",
			path:         "path",
			overwriteURL: "http://example.com/",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "without slash on url",
			baseURL:      "http://foobar.com",
			method:       "GET",
			path:         "/path",
			overwriteURL: "http://example.com",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "without slash on url and path",
			baseURL:      "http://foobar.com",
			method:       "GET",
			path:         "path",
			overwriteURL: "http://example.com",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "with slash on url and path",
			baseURL:      "http://foobar.com",
			method:       "GET",
			path:         "/path",
			overwriteURL: "http://example.com/",
			expectedURL:  "http://example.com/path",
		},
		{
			name:         "without slash on path",
			baseURL:      "http://foobar.com",
			method:       "GET",
			path:         "path",
			overwriteURL: "http://example.com/",
			expectedURL:  "http://example.com/path",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockClient{}
			reporter := NewAssertReporter(t)

			req := NewRequestC(
				Config{
					BaseURL:  tc.baseURL,
					Client:   client,
					Reporter: reporter,
				},
				tc.method,
				tc.path)

			req.WithURL(tc.overwriteURL)

			req.Expect().chain.assertNotFailed(t)
			req.chain.assertNotFailed(t)
			assert.Equal(t, tc.expectedURL, client.req.URL.String())
		})
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

	reqs[0] = NewRequestC(config, "GET", "/foo/{arg}", "bar")
	reqs[1] = NewRequestC(config, "GET", "{arg}foo{arg}", "/", "/bar")
	reqs[2] = NewRequestC(config, "GET", "{arg}", "/foo/bar")

	for _, req := range reqs {
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, "http://example.com/foo/bar", client.req.URL.String())
	}

	r1 := NewRequestC(config, "GET", "/{arg1}/{arg2}", "foo")
	r1.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/foo/%7Barg2%7D",
		client.req.URL.String())

	r2 := NewRequestC(config, "GET", "/{arg1}/{arg2}/{arg3}")
	r2.WithPath("ARG3", "foo")
	r2.WithPath("arg2", "bar")
	r2.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/%7Barg1%7D/bar/foo",
		client.req.URL.String())

	r3 := NewRequestC(config, "GET", "/{arg1}.{arg2}.{arg3}")
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

	r4 := NewRequestC(config, "GET", "/{arg1}{arg2}")
	r4.WithPathObject(S{"foo", 1, 2})
	r4.Expect().chain.assertNotFailed(t)
	assert.Equal(t, "http://example.com/foo1", client.req.URL.String())

	r5 := NewRequestC(config, "GET", "/{arg1}{arg2}")
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

	checkOK := func(req *Request, url string) {
		client.req = nil
		req.Expect()
		req.chain.assertNotFailed(t)
		assert.Equal(t, url, client.req.URL.String())
	}

	checkFailed := func(req *Request) {
		req.chain.assertFailed(t)
	}

	t.Run("WithQuery", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQuery("aa", "foo").WithQuery("bb", 123).WithQuery("cc", "*&@")
		checkOK(req,
			"http://example.com/path?aa=foo&bb=123&cc=%2A%26%40")
	})

	t.Run("WithQueryObject map[string]interface", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQuery("aa", "foo").
			WithQueryObject(map[string]interface{}{
				"bb": 123,
				"cc": "*&@",
			})
		checkOK(req,
			"http://example.com/path?aa=foo&bb=123&cc=%2A%26%40")
	})

	type S struct {
		Bb int    `url:"bb"`
		Cc string `url:"cc"`
		Dd string `url:"-"`
	}

	t.Run("WithQueryObject struct", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQueryObject(S{123, "*&@", "dummy"}).WithQuery("aa", "foo")
		checkOK(req,
			"http://example.com/path?aa=foo&bb=123&cc=%2A%26%40")
	})

	t.Run("WithQueryObject pointer to struct", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQueryObject(&S{123, "*&@", "dummy"}).WithQuery("aa", "foo")
		checkOK(req,
			"http://example.com/path?aa=foo&bb=123&cc=%2A%26%40")
	})

	t.Run("WithQuery and WithQueryString", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQuery("bb", 123).
			WithQueryString("aa=foo&cc=%2A%26%40")
		checkOK(req,
			"http://example.com/path?aa=foo&bb=123&cc=%2A%26%40")
	})

	t.Run("WithQueryString and WithQuery", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQueryString("aa=foo&cc=%2A%26%40").
			WithQuery("bb", 123)
		checkOK(req,
			"http://example.com/path?aa=foo&bb=123&cc=%2A%26%40")
	})

	t.Run("WithQueryObject nil", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQuery("foo", "bar").
			WithQueryObject(nil)
		checkOK(req,
			"http://example.com/path?foo=bar")
	})

	t.Run("WithQueryString invalid", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQueryString("%")
		checkFailed(req)
	})

	t.Run("WithQueryObject invalid", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path").
			WithQueryObject(func() {})
		checkFailed(req)
	})

	t.Run("WithQueryObject invalid struct", func(t *testing.T) {
		type invalidSt struct {
			Str mockQueryEncoder
		}
		queryObj := invalidSt{
			Str: mockQueryEncoder("err"),
		}
		req := NewRequestC(config, "GET", "/path").
			WithQueryObject(queryObj)
		checkFailed(req)
	})
}

func TestRequest_Headers(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	req := NewRequestC(config, "GET", "url")

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

	assert.Equal(t, "GET", client.req.Method)
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

	req := NewRequestC(config, "GET", "url")

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

	assert.Equal(t, "GET", client.req.Method)
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

	req := NewRequestC(config, "GET", "url")

	req.WithBasicAuth("Aladdin", "open sesame")
	req.chain.assertNotFailed(t)

	assert.Equal(t, "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==",
		req.httpReq.Header.Get("Authorization"))
}

func TestRequest_Host(t *testing.T) {
	cases := []struct {
		name         string
		method       string
		path         string
		expectedHost string
		setupFunc    func(req *Request)
	}{
		{
			name:         "request with host and without host header",
			method:       "GET",
			path:         "url",
			expectedHost: "example.com",
			setupFunc: func(req *Request) {
				req.WithHost("example.com")
			},
		},
		{
			name:         "request with header before with host",
			method:       "GET",
			path:         "url",
			expectedHost: "example2.com",
			setupFunc: func(req *Request) {
				req.WithHost("example1.com")
				req.withHeader("HOST", "example2.com")
			},
		},
		{
			name:         "request with host before with header",
			method:       "GET",
			path:         "url",
			expectedHost: "example1.com",
			setupFunc: func(req *Request) {
				req.WithHost("example2.com")
				req.withHeader("HOST", "example1.com")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockClient{}
			reporter := newMockReporter(t)

			req := NewRequestC(
				Config{
					Client:   client,
					Reporter: reporter,
				},
				tc.method,
				tc.path)

			if tc.setupFunc != nil {
				tc.setupFunc(req)
			}

			resp := req.Expect()
			req.chain.assertNotFailed(t)
			resp.chain.assertNotFailed(t)

			assert.Equal(t, tc.method, client.req.Method)
			assert.Equal(t, tc.expectedHost, client.req.Host)
			assert.Equal(t, tc.path, client.req.URL.String())

			assert.Same(t, &client.resp, resp.Raw())
		})
	}
}

func TestRequest_BodyChunked(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("body", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

		req.WithChunked(bytes.NewBufferString("body"))

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.NotNil(t, client.req.Body)
		assert.Equal(t, int64(-1), client.req.ContentLength)

		assert.Equal(t, "GET", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, make(http.Header), client.req.Header)
		assert.Equal(t, "body", resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("nil", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

		req.WithChunked(nil)

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, http.NoBody, client.req.Body)
		assert.Equal(t, int64(0), client.req.ContentLength)
	})

	t.Run("proto 1.0", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

		req.WithProto("HTTP/1.0")
		assert.Equal(t, 1, req.httpReq.ProtoMajor)
		assert.Equal(t, 0, req.httpReq.ProtoMinor)

		req.WithChunked(bytes.NewBufferString("body"))
		req.chain.assertFailed(t)
	})

	t.Run("proto 2.0", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

		req.WithProto("HTTP/2.0")
		assert.Equal(t, 2, req.httpReq.ProtoMajor)
		assert.Equal(t, 0, req.httpReq.ProtoMinor)

		req.WithChunked(bytes.NewBufferString("body"))
		assert.Equal(t, 2, req.httpReq.ProtoMajor)
		assert.Equal(t, 0, req.httpReq.ProtoMinor)
	})
}

func TestRequest_BodyBytes(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("byte slice", func(t *testing.T) {
		req := NewRequestC(config, "GET", "/path")

		req.WithBytes([]byte("body"))

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.NotNil(t, client.req.Body)
		assert.Equal(t, int64(len("body")), client.req.ContentLength)

		assert.Equal(t, "GET", client.req.Method)
		assert.Equal(t, "/path", client.req.URL.String())
		assert.Equal(t, make(http.Header), client.req.Header)
		assert.Equal(t, "body", resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("nil", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

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

	req := NewRequestC(config, "GET", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithText("some text")

	resp := req.Expect()
	resp.chain.assertNotFailed(t)

	assert.Equal(t, "GET", client.req.Method)
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

		req := NewRequestC(config, "GET", "url")

		req.WithHeaders(map[string]string{
			"Some-Header": "foo",
		})

		req.WithForm(map[string]interface{}{
			"a": 1,
			"b": "2",
		})

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "GET", client.req.Method)
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

		req := NewRequestC(config, "GET", "url")

		req.WithHeaders(map[string]string{
			"Some-Header": "foo",
		})

		req.WithFormField("a", 1)
		req.WithFormField("b", "2")

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "GET", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("form struct", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		req := NewRequestC(config, "GET", "url")

		type S struct {
			A string `form:"a"`
			B int    `form:"b"`
			C int    `form:"-"`
		}

		req.WithForm(S{"1", 2, 3})

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "GET", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("form combined", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		req := NewRequestC(config, "GET", "url")

		type S struct {
			A int `form:"a"`
		}

		req.WithForm(S{A: 1})
		req.WithForm(map[string]string{"b": "2"})
		req.WithFormField("c", 3)

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "GET", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `a=1&b=2&c=3`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("marshal error", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

		req.WithForm(func() {})

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
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

	t.Run("json", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/json; charset=utf-8"},
			"Some-Header":  {"foo"},
		}

		req := NewRequestC(config, "GET", "url")

		req.WithHeaders(map[string]string{
			"Some-Header": "foo",
		})

		req.WithJSON(map[string]interface{}{"key": "value"})

		resp := req.Expect()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, "GET", client.req.Method)
		assert.Equal(t, "url", client.req.URL.String())
		assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
		assert.Equal(t, `{"key":"value"}`, resp.Body().Raw())

		assert.Same(t, &client.resp, resp.Raw())
	})

	t.Run("marshal error", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")

		req.WithJSON(func() {})

		resp := req.Expect()
		resp.chain.assertFailed(t)

		assert.Nil(t, resp.Raw())
	})
}

func TestRequest_ContentLength(t *testing.T) {
	client := &mockClient{}
	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("chunked", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithChunked(bytes.NewReader([]byte("12345")))
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, int64(-1), client.req.ContentLength)
	})

	t.Run("bytes", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithBytes([]byte("12345"))
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, int64(5), client.req.ContentLength)
	})

	t.Run("text", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithText("12345")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, int64(5), client.req.ContentLength)
	})

	t.Run("json", func(t *testing.T) {
		j, _ := json.Marshal(map[string]string{"a": "b"})
		req := NewRequestC(config, "GET", "url")
		req.WithJSON(map[string]string{"a": "b"})
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, int64(len(j)), client.req.ContentLength)
	})

	t.Run("form", func(t *testing.T) {
		f := `a=b`
		req := NewRequestC(config, "GET", "url")
		req.WithForm(map[string]string{"a": "b"})
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, int64(len(f)), client.req.ContentLength)
	})

	t.Run("form field", func(t *testing.T) {
		f := `a=b`
		req := NewRequestC(config, "GET", "url")
		req.WithFormField("a", "b")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, int64(len(f)), client.req.ContentLength)
	})

	t.Run("multipart", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithMultipart()
		req.WithFileBytes("a", "b", []byte("12345"))
		req.Expect().chain.assertNotFailed(t)
		assert.True(t, client.req.ContentLength > 0)
	})
}

func TestRequest_ContentType(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("WithText sets Content-Type header", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithText("hello")
		req.WithHeader("Content-Type", "foo")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)
	})

	t.Run("WithHeader sets Content-Type header", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithHeader("Content-Type", "foo")
		req.WithText("hello")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)
	})

	t.Run("WithJSON overrides Content-Type header", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithJSON(map[string]interface{}{"a": "b"})
		req.WithHeader("Content-Type", "foo")
		req.WithHeader("Content-Type", "bar")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)
	})

	t.Run("WithForm overrides Content-Type header", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithForm(map[string]interface{}{"a": "b"})
		req.WithHeader("Content-Type", "foo")
		req.WithHeader("Content-Type", "bar")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)
	})

	t.Run("WithMultipart overrides Content-Type header", func(t *testing.T) {
		req := NewRequestC(config, "GET", "url")
		req.WithMultipart()
		req.WithForm(map[string]interface{}{"a": "b"})
		req.WithHeader("Content-Type", "foo")
		req.WithHeader("Content-Type", "bar")
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)
	})
}

func TestRequest_Websocket(t *testing.T) {
	t.Run("ws successful", func(t *testing.T) {
		scheme := ""
		dialer := WebsocketDialerFunc(func(
			url string, _ http.Header,
		) (*websocket.Conn, *http.Response, error) {
			u, _ := neturl.Parse(url)
			scheme = u.Scheme
			return &websocket.Conn{}, &http.Response{}, nil
		})
		config := Config{
			Reporter:        newMockReporter(t),
			WebsocketDialer: dialer,
			BaseURL:         "http://example.com",
		}
		req := NewRequestC(config, "GET", "url").WithWebsocketUpgrade()
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, "ws", scheme)
	})
	t.Run("wss successful", func(t *testing.T) {
		scheme := ""
		dialer := WebsocketDialerFunc(func(
			url string, _ http.Header,
		) (*websocket.Conn, *http.Response, error) {
			u, _ := neturl.Parse(url)
			scheme = u.Scheme
			return &websocket.Conn{}, &http.Response{}, nil
		})
		config := Config{
			Reporter:        newMockReporter(t),
			WebsocketDialer: dialer,
			BaseURL:         "https://example.com",
		}
		req := NewRequestC(config, "GET", "url").WithWebsocketUpgrade()
		req.Expect().chain.assertNotFailed(t)
		assert.Equal(t, "wss", scheme)
	})
	t.Run("bad handshake", func(t *testing.T) {
		dialer := WebsocketDialerFunc(func(
			_ string, _ http.Header,
		) (*websocket.Conn, *http.Response, error) {
			return &websocket.Conn{}, &http.Response{}, websocket.ErrBadHandshake
		})
		config := Config{
			Reporter:        newMockReporter(t),
			WebsocketDialer: dialer,
		}
		req := NewRequestC(config, "GET", "url").WithWebsocketUpgrade()
		req.Expect().chain.assertNotFailed(t)
	})
	t.Run("request body not allowed", func(t *testing.T) {
		dialer := WebsocketDialerFunc(func(
			_ string, _ http.Header,
		) (*websocket.Conn, *http.Response, error) {
			return &websocket.Conn{}, &http.Response{}, nil
		})
		config := Config{
			Reporter:        newMockReporter(t),
			WebsocketDialer: dialer,
		}
		req := NewRequestC(config, "GET", "url").
			WithJSON("").
			WithWebsocketUpgrade()
		req.Expect().chain.assertFailed(t)
	})
}

func TestRequest_RedirectsDontFollow(t *testing.T) {
	t.Run("no body", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.assertFn = func(r *http.Request) {
			assert.Equal(t, http.NoBody, r.Body)
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.assertFn = func(r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
}

func TestRequest_RedirectsFollowAll(t *testing.T) {
	t.Run("no body", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.maxRedirect = 1
		tp.assertFn = func(r *http.Request) {
			assert.Equal(t, http.NoBody, r.Body)
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})

	t.Run("no body, too many redirects", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.assertFn = func(r *http.Request) {
			assert.Equal(t, http.NoBody, r.Body)
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})

	t.Run("has body", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.maxRedirect = 1
		tp.assertFn = func(r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})

	t.Run("has body, too many redirects", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.assertFn = func(r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})
}

func TestRequest_RedirectsFollowWithoutBody(t *testing.T) {
	t.Run("no body", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.maxRedirect = 1
		tp.assertFn = func(r *http.Request) {
			assert.Contains(t, []interface{}{nil, http.NoBody}, r.Body)
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})

	t.Run("no body, too many redirects", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.assertFn = func(r *http.Request) {
			assert.Contains(t, []interface{}{nil, http.NoBody}, r.Body)
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})

	t.Run("has body, status permanent redirect", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.maxRedirect = 1
		tp.assertFn = func(r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test body", string(b))
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 1, tp.tripCount)
	})

	t.Run("has body, status moved permanently", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.maxRedirect = 1
		tp.redirectHTTPStatusCode = http.StatusMovedPermanently
		tp.assertFn = func(r *http.Request) {
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
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
	})

	t.Run("has body, status moved permanently, too many redirects", func(t *testing.T) {
		reporter := newMockReporter(t)

		tp := newMockTransportRedirect()
		tp.redirectHTTPStatusCode = http.StatusMovedPermanently
		tp.assertFn = func(r *http.Request) {
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
		}

		config := Config{
			Client:   &http.Client{Transport: tp},
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
			httpClient.CheckRedirect(req.httpReq, make([]*http.Request, 2)))

		// Should do round trip
		assert.Equal(t, 2, tp.tripCount)
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

func TestRequest_Conflicts(t *testing.T) {
	client := &mockClient{}

	config := Config{
		Client:   client,
		Reporter: newMockReporter(t),
	}

	t.Run("body conflict", func(t *testing.T) {
		var req *Request

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithChunked(nil)
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithBytes(nil)
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithText("")
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithJSON(map[string]interface{}{"a": "b"})
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithForm(map[string]interface{}{"a": "b"})
		req.Expect()
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithFormField("a", "b")
		req.Expect()
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithChunked(nil)
		req.chain.assertNotFailed(t)
		req.WithMultipart()
		req.chain.assertFailed(t)
	})

	t.Run("type conflict", func(t *testing.T) {
		var req *Request

		req = NewRequestC(config, "GET", "url")
		req.WithText("")
		req.chain.assertNotFailed(t)
		req.WithJSON(map[string]interface{}{"a": "b"})
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithText("")
		req.chain.assertNotFailed(t)
		req.WithForm(map[string]interface{}{"a": "b"})
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithText("")
		req.chain.assertNotFailed(t)
		req.WithFormField("a", "b")
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithText("")
		req.chain.assertNotFailed(t)
		req.WithMultipart()
		req.chain.assertFailed(t)
	})

	t.Run("multipart conflict", func(t *testing.T) {
		var req *Request

		req = NewRequestC(config, "GET", "url")
		req.WithForm(map[string]interface{}{"a": "b"})
		req.chain.assertNotFailed(t)
		req.WithMultipart()
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithFormField("a", "b")
		req.chain.assertNotFailed(t)
		req.WithMultipart()
		req.chain.assertFailed(t)

		req = NewRequestC(config, "GET", "url")
		req.WithFileBytes("a", "a", []byte("a"))
		req.chain.assertFailed(t)
	})
}

func TestRequest_Usage(t *testing.T) {
	cases := []struct {
		name        string
		client      Client
		prepFunc    func(req *Request)
		prepFails   bool
		expectFails bool
	}{
		{
			name: "WithMatcher - nil argument",
			prepFunc: func(req *Request) {
				req.WithMatcher(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithTransformer - nil argument",
			prepFunc: func(req *Request) {
				req.WithTransformer(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithClient - nil argument",
			prepFunc: func(req *Request) {
				req.WithClient(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithHandler - nil argument",
			prepFunc: func(req *Request) {
				req.WithHandler(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithContext - nil argument",
			prepFunc: func(req *Request) {
				req.WithContext(nil) //nolint
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithMaxRedirects - negative argument",
			prepFunc: func(req *Request) {
				req.WithMaxRedirects(-1)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithMaxRetries - negative argument",
			prepFunc: func(req *Request) {
				req.WithMaxRetries(-1)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithRetryDelay - invalid range",
			prepFunc: func(req *Request) {
				req.WithRetryDelay(10, 5)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithWebsocketDialer - nil argument",
			prepFunc: func(req *Request) {
				req.WithWebsocketDialer(nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithPath - nil argument",
			prepFunc: func(req *Request) {
				req.WithPath("test-key", nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithQuery - nil argument",
			prepFunc: func(req *Request) {
				req.WithQuery("test-query", nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithURL - invalid url",
			prepFunc: func(req *Request) {
				req.WithURL("%-invalid-url")
			},
			prepFails:   true,
			expectFails: true,
		},
		{
			name: "WithFile - multiple readers",
			prepFunc: func(req *Request) {
				req.WithFile("test-key", "test-path", nil, nil)
			},
			prepFails:   true,
			expectFails: true,
		},
		// WithRedirectPolicy and WithMaxRedirects require Client
		// to be http.Client, but we use another one
		{
			name:   "WithRedirectPolicy - incompatible client",
			client: &mockClient{},
			prepFunc: func(req *Request) {
				req.WithRedirectPolicy(FollowAllRedirects)
			},
			prepFails:   false,
			expectFails: true,
		},
		{
			name:   "WithMaxRedirects - incompatible client",
			client: &mockClient{},
			prepFunc: func(req *Request) {
				req.WithMaxRedirects(1)
			},
			prepFails:   false,
			expectFails: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				Client:   tc.client,
				Reporter: newMockReporter(t),
			}

			req := NewRequestC(config, "GET", "/")

			tc.prepFunc(req)

			if tc.prepFails {
				req.chain.assertFailed(t)
			} else {
				req.chain.assertNotFailed(t)

				resp := req.Expect()

				if tc.expectFails {
					req.chain.assertFailed(t)
					resp.chain.assertFailed(t)
				} else {
					req.chain.assertNotFailed(t)
					resp.chain.assertNotFailed(t)
				}
			}
		})
	}
}

func TestRequest_Order(t *testing.T) {
	cases := []struct {
		name       string
		beforeFunc func(req *Request)
		afterFunc  func(req *Request)
	}{
		{
			name: "Expect after Expect",
			afterFunc: func(req *Request) {
				req.Expect()
			},
		},
		{
			name: "WithName after Expect",
			afterFunc: func(req *Request) {
				req.WithName("Test")
			},
		},
		{
			name: "WithMatcher after Expect",
			afterFunc: func(req *Request) {
				req.WithMatcher(func(*Response) {
				})
			},
		},
		{
			name: "WithTransformer after Expect",
			afterFunc: func(req *Request) {
				req.WithTransformer(func(*http.Request) {
				})
			},
		},
		{
			name: "WithClient after Expect",
			afterFunc: func(req *Request) {
				req.WithClient(&mockClient{})
			},
		},
		{
			name: "WithHandler after Expect",
			afterFunc: func(req *Request) {
				req.WithHandler(http.NotFoundHandler())
			},
		},
		{
			name: "WithContext after Expect",
			afterFunc: func(req *Request) {
				req.WithContext(context.Background())
			},
		},
		{
			name: "WithTimeout after Expect",
			afterFunc: func(req *Request) {
				req.WithTimeout(3 * time.Second)
			},
		},
		{
			name: "WithRedirectPolicy after Expect",
			afterFunc: func(req *Request) {
				req.WithRedirectPolicy(FollowAllRedirects)
			},
		},
		{
			name: "WithMaxRedirects after Expect",
			afterFunc: func(req *Request) {
				req.WithMaxRedirects(3)
			},
		},
		{
			name: "WithRetryPolicy after Expect",
			afterFunc: func(req *Request) {
				req.WithRetryPolicy(DontRetry)
			},
		},
		{
			name: "WithMaxRetries after Expect",
			afterFunc: func(req *Request) {
				req.WithMaxRetries(10)
			},
		},
		{
			name: "WithRetryDelay after Expect",
			afterFunc: func(req *Request) {
				req.WithRetryDelay(time.Second, 5*time.Second)
			},
		},
		{
			name: "WithWebsocketUpgrade after Expect",
			afterFunc: func(req *Request) {
				req.WithWebsocketUpgrade()
			},
		},
		{
			name: "WithWebsocketDialer after Expect",
			afterFunc: func(req *Request) {
				req.WithWebsocketDialer(&websocket.Dialer{})
			},
		},
		{
			name: "WithPath after Expect",
			afterFunc: func(req *Request) {
				req.WithPath("repo", "repo1")
			},
		},
		{
			name: "WithPathObject after Expect",
			afterFunc: func(req *Request) {
				req.WithPathObject(map[string]string{
					"repo": "repo1",
				})
			},
		},
		{
			name: "WithQuery after Expect",
			afterFunc: func(req *Request) {
				req.WithQuery("a", 123)
			},
		},
		{
			name: "WithQueryObject after Expect",
			afterFunc: func(req *Request) {
				req.WithQueryObject(map[string]string{
					"a": "val",
				})
			},
		},
		{
			name: "WithQueryString after Expect",
			afterFunc: func(req *Request) {
				req.WithQueryString("a=123&b=hello")
			},
		},
		{
			name: "WithURL after Expect",
			afterFunc: func(req *Request) {
				req.WithURL("https://www.github.com")
			},
		},
		{
			name: "WithHeaders after Expect",
			afterFunc: func(req *Request) {
				req.WithHeaders(map[string]string{
					"Content-Type": "application/json",
				})
			},
		},
		{
			name: "WithHeader after Expect",
			afterFunc: func(req *Request) {
				req.WithHeader("Content-Type", "application/json")
			},
		},
		{
			name: "WithCookies after Expect",
			afterFunc: func(req *Request) {
				req.WithCookies(map[string]string{
					"key1": "val1",
				})
			},
		},
		{
			name: "WithCookie after Expect",
			afterFunc: func(req *Request) {
				req.WithCookie("key1", "val1")
			},
		},
		{
			name: "WithBasicAuth after Expect",
			afterFunc: func(req *Request) {
				req.WithBasicAuth("user", "pass")
			},
		},
		{
			name: "WithHost after Expect",
			afterFunc: func(req *Request) {
				req.WithHost("localhost")
			},
		},
		{
			name: "WithProto after Expect",
			afterFunc: func(req *Request) {
				req.WithProto("HTTP/1.1")
			},
		},
		{
			name: "WithChunked after Expect",
			afterFunc: func(req *Request) {
				req.WithChunked(bytes.NewReader(nil))
			},
		},
		{
			name: "WithBytes after Expect",
			afterFunc: func(req *Request) {
				req.WithBytes(nil)
			},
		},
		{
			name: "WithText after Expect",
			afterFunc: func(req *Request) {
				req.WithText("hello")
			},
		},
		{
			name: "WithJSON after Expect",
			afterFunc: func(req *Request) {
				req.WithJSON(map[string]string{
					"key1": "val1",
				})
			},
		},
		{
			name: "WithForm after Expect",
			afterFunc: func(req *Request) {
				req.WithForm(map[string]string{
					"key1": "val1",
				})
			},
		},
		{
			name: "WithFormField after Expect",
			afterFunc: func(req *Request) {
				req.WithFormField("key1", 123)
			},
		},
		{
			name: "WithFile after Expect",
			beforeFunc: func(req *Request) {
				req.WithMultipart()
			},
			afterFunc: func(req *Request) {
				req.WithFile("foo", "bar", strings.NewReader("baz"))
			},
		},
		{
			name: "WithFileBytes after Expect",
			beforeFunc: func(req *Request) {
				req.WithMultipart()
			},
			afterFunc: func(req *Request) {
				req.WithFileBytes("foo", "bar", []byte("baz"))
			},
		},
		{
			name: "WithMultipart after Expect",
			afterFunc: func(req *Request) {
				req.WithMultipart()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				Client:   &mockClient{},
				Reporter: newMockReporter(t),
			}

			req := NewRequestC(config, "GET", "/")

			if tc.beforeFunc != nil {
				tc.beforeFunc(req)
			}
			req.chain.assertNotFailed(t)

			resp := req.Expect()
			req.chain.assertNotFailed(t)
			resp.chain.assertNotFailed(t)

			tc.afterFunc(req)
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

		assert.Panics(t, func() { newRequest(newMockChain(t), config, "GET", "") })
	})

	t.Run("Client is nil", func(t *testing.T) {
		config := Config{
			RequestFactory:   DefaultRequestFactory{},
			Client:           nil,
			AssertionHandler: &mockAssertionHandler{},
		}

		assert.Panics(t, func() { newRequest(newMockChain(t), config, "GET", "") })
	})

	t.Run("AssertionHandler is nil", func(t *testing.T) {
		config := Config{
			RequestFactory:   DefaultRequestFactory{},
			Client:           &mockClient{},
			AssertionHandler: nil,
		}

		assert.Panics(t, func() { newRequest(newMockChain(t), config, "GET", "") })
	})
}
