package httpexpect

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponse_FailedChain(t *testing.T) {
	check := func(resp *Response) {
		resp.chain.assertFailed(t)

		resp.Alias("foo")

		resp.RoundTripTime().chain.assertFailed(t)
		resp.Duration().chain.assertFailed(t)
		resp.Headers().chain.assertFailed(t)
		resp.Header("foo").chain.assertFailed(t)
		resp.Cookies().chain.assertFailed(t)
		resp.Cookie("foo").chain.assertFailed(t)
		resp.Body().chain.assertFailed(t)
		resp.Text().chain.assertFailed(t)
		resp.Form().chain.assertFailed(t)
		resp.JSON().chain.assertFailed(t)
		resp.JSONP("").chain.assertFailed(t)
		resp.Websocket().chain.assertFailed(t)

		resp.Status(123)
		resp.StatusRange(Status2xx)
		resp.StatusList(http.StatusOK, http.StatusBadGateway)
		resp.NoContent()
		resp.ContentType("", "")
		resp.ContentEncoding("")
		resp.TransferEncoding("")
	}

	t.Run("failed chain", func(t *testing.T) {
		reporter := newMockReporter(t)
		chain := newChainWithDefaults("test", reporter)
		config := newMockConfig(reporter)

		chain.setFailed()

		resp := newResponse(responseOpts{
			config:   config,
			chain:    chain,
			httpResp: &http.Response{},
		})

		check(resp)
	})

	t.Run("nil value", func(t *testing.T) {
		reporter := newMockReporter(t)
		chain := newChainWithDefaults("test", reporter)
		config := newMockConfig(reporter)

		resp := newResponse(responseOpts{
			config:   config,
			chain:    chain,
			httpResp: nil,
		})

		check(resp)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
		reporter := newMockReporter(t)
		chain := newChainWithDefaults("test", reporter)
		config := newMockConfig(reporter)

		chain.setFailed()

		resp := newResponse(responseOpts{
			config:   config,
			chain:    chain,
			httpResp: nil,
		})

		check(resp)
	})
}

func TestResponse_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		resp := NewResponse(reporter, &http.Response{})
		resp.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		resp := NewResponseC(Config{
			Reporter: reporter,
		}, &http.Response{})
		resp.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		reporter := newMockReporter(t)
		config := newMockConfig(reporter)
		value := newResponse(responseOpts{
			config:   config,
			chain:    chain,
			httpResp: &http.Response{},
		})
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestResponse_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewResponse(reporter, &http.Response{}, time.Second)
	assert.Equal(t, []string{"Response()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Response()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Response()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestResponse_RoundTripTime(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("set", func(t *testing.T) {
		duration := time.Second

		resp := NewResponse(reporter, &http.Response{}, duration)
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		rt := resp.RoundTripTime()

		assert.Equal(t, time.Second, rt.Raw())

		rt.IsSet()
		rt.IsEqual(time.Second)
		rt.chain.assertNotFailed(t)
	})

	t.Run("unset", func(t *testing.T) {
		resp := NewResponse(reporter, &http.Response{})
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		rt := resp.RoundTripTime()

		assert.Equal(t, time.Duration(0), rt.Raw())

		rt.NotSet()
		rt.chain.assertNotFailed(t)

		rt.IsSet()
		rt.chain.assertFailed(t)
	})
}

func TestResponse_StatusRange(t *testing.T) {
	reporter := newMockReporter(t)

	ranges := []StatusRange{
		Status1xx,
		Status2xx,
		Status3xx,
		Status4xx,
		Status5xx,
	}

	cases := []struct {
		Status int
		Range  StatusRange
	}{
		{99, StatusRange(-1)},
		{100, Status1xx},
		{199, Status1xx},
		{200, Status2xx},
		{299, Status2xx},
		{300, Status3xx},
		{399, Status3xx},
		{400, Status4xx},
		{499, Status4xx},
		{500, Status5xx},
		{599, Status5xx},
		{600, StatusRange(-1)},
	}

	for _, test := range cases {
		for _, r := range ranges {
			resp := NewResponse(reporter, &http.Response{
				StatusCode: test.Status,
			})

			resp.StatusRange(r)

			if test.Range == r {
				resp.chain.assertNotFailed(t)
			} else {
				resp.chain.assertFailed(t)
			}
		}
	}
}

func TestResponse_StatusList(t *testing.T) {
	reporter := newMockReporter(t)

	cases := []struct {
		Status int
		List   []int
		WantOK bool
	}{
		{
			http.StatusOK,
			[]int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError},
			true,
		},
		{
			http.StatusBadRequest,
			[]int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError},
			true,
		},
		{
			http.StatusOK,
			[]int{http.StatusInternalServerError, http.StatusBadRequest},
			false,
		},
		{
			http.StatusBadGateway,
			[]int{},
			false,
		},
	}

	for _, c := range cases {
		resp := NewResponse(reporter, &http.Response{
			StatusCode: c.Status,
		})
		resp.StatusList(c.List...)
		if c.WantOK {
			resp.chain.assertNotFailed(t)
		} else {
			resp.chain.assertFailed(t)
		}
	}
}

func TestResponse_Headers(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"First-Header":  {"foo"},
		"Second-Header": {"bar"},
	}

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       nil,
	}

	resp := NewResponse(reporter, httpResp)
	resp.chain.assertNotFailed(t)
	resp.chain.clearFailed()

	assert.Same(t, httpResp, resp.Raw())

	resp.Status(http.StatusOK)
	resp.chain.assertNotFailed(t)
	resp.chain.clearFailed()

	resp.Status(http.StatusNotFound)
	resp.chain.assertFailed(t)
	resp.chain.clearFailed()

	resp.Headers().IsEqual(headers).chain.assertNotFailed(t)

	for k, v := range headers {
		for _, h := range []string{k, strings.ToLower(k), strings.ToUpper(k)} {
			resp.Header(h).IsEqual(v[0]).chain.assertNotFailed(t)
		}
	}

	resp.Header("Bad-Header").IsEmpty().chain.assertNotFailed(t)
}

func TestResponse_Cookies(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("cookies", func(t *testing.T) {
		headers := map[string][]string{
			"Set-Cookie": {
				"foo=aaa",
				"bar=bbb; expires=Fri, 31 Dec 2010 23:59:59 GMT; " +
					"path=/xxx; domain=example.com",
			},
		}

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       nil,
		}

		resp := NewResponse(reporter, httpResp)
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t, []interface{}{"foo", "bar"}, resp.Cookies().Raw())
		resp.chain.assertNotFailed(t)

		c1 := resp.Cookie("foo")
		resp.chain.assertNotFailed(t)
		assert.Equal(t, "foo", c1.Raw().Name)
		assert.Equal(t, "aaa", c1.Raw().Value)
		assert.Equal(t, "", c1.Raw().Domain)
		assert.Equal(t, "", c1.Raw().Path)

		c2 := resp.Cookie("bar")
		resp.chain.assertNotFailed(t)
		assert.Equal(t, "bar", c2.Raw().Name)
		assert.Equal(t, "bbb", c2.Raw().Value)
		assert.Equal(t, "example.com", c2.Raw().Domain)
		assert.Equal(t, "/xxx", c2.Raw().Path)
		assert.True(t, time.Date(2010, 12, 31, 23, 59, 59, 0, time.UTC).
			Equal(c2.Raw().Expires))

		c3 := resp.Cookie("baz")
		resp.chain.assertFailed(t)
		c3.chain.assertFailed(t)
		assert.Nil(t, c3.Raw())
	})

	t.Run("no cookies", func(t *testing.T) {
		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     nil,
			Body:       nil,
		}

		resp := NewResponse(reporter, httpResp)
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t, []interface{}{}, resp.Cookies().Raw())
		resp.chain.assertNotFailed(t)

		c := resp.Cookie("foo")
		resp.chain.assertFailed(t)
		c.chain.assertFailed(t)
		assert.Nil(t, c.Raw())
	})
}

func TestResponse_BodyOperations(t *testing.T) {
	t.Run("content", func(t *testing.T) {
		reporter := newMockReporter(t)

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("body")),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "body", resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("read and close", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("test_body")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "test_body", resp.Body().Raw())
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)

		resp.chain.assertNotFailed(t)
	})

	t.Run("read error", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("test_body")
		body.readErr = errors.New("test_error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)
		respBody := resp.Body()

		assert.Equal(t, "", respBody.Raw())
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)

		respBody.chain.assertFailed(t)
		resp.chain.assertFailed(t)
	})

	t.Run("close error", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("test_body")
		body.closeErr = errors.New("test_error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "", resp.Body().Raw())
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)

		resp.chain.assertFailed(t)
	})
}

func TestResponse_BodyDeferred(t *testing.T) {
	t.Run("constructor does not read content", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("body string")
		resp := NewResponse(reporter, &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		})

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentPending, resp.contentState)
	})

	t.Run("content is remembered", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("body string")
		resp := NewResponse(reporter, &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		})

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentPending, resp.contentState)

		// Read body
		resp.Body()
		resp.chain.assertNotFailed(t)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, []byte("body string"), resp.content)
		assert.Equal(t, contentRetreived, resp.contentState)

		// Second call should be no-op
		resp.Body()
		resp.chain.assertNotFailed(t)

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, []byte("body string"), resp.content)
		assert.Equal(t, contentRetreived, resp.contentState)
	})

	t.Run("read error is remembered", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("body string")
		body.readErr = errors.New("test error")

		resp := NewResponse(reporter, &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		})

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentPending, resp.contentState)

		// Read body
		resp.Body()
		resp.chain.assertFailed(t)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)

		// Second call should be no-op
		resp.Body()
		resp.chain.assertFailed(t)

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)
	})

	t.Run("close error is remembered", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("body string")
		body.closeErr = errors.New("test error")

		resp := NewResponse(reporter, &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		})

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentPending, resp.contentState)

		// Read body
		resp.Body()
		resp.chain.assertFailed(t)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)

		// Second call should be no-op
		resp.Body()
		resp.chain.assertFailed(t)

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)
	})

	t.Run("failed state", func(t *testing.T) {
		reporter := newMockReporter(t)

		body := newMockBody("body string")
		body.readErr = errors.New("test error")

		resp := NewResponse(reporter, &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		})

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentPending, resp.contentState)

		// Read body
		resp.Body()
		resp.chain.assertFailed(t)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)

		// Invoke getContent()
		chain := resp.chain.enter("Test()")
		content, ok := resp.getContent(chain)

		chain.assertFailed(t)
		assert.Nil(t, content)
		assert.False(t, ok)

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)
	})
}

func TestResponse_NoContent(t *testing.T) {
	t.Run("empty Content-Type, empty Body", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {""},
		}

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "", resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.NoContent()
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.Text()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.Form()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.JSON()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.JSONP("")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("empty Content-Type, nil Body", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {""},
		}

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       nil,
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "", resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.NoContent()
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.Text()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.Form()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.JSON()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.JSONP("")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("non-empty Content-Type, empty Body", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain; charset=utf-8"},
		}

		body := ``

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.NoContent()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("empty Content-Type, Body read failure", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {""},
		}

		body := newMockBody("")
		body.readErr = errors.New("test_error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		resp.NoContent()
		resp.chain.assertFailed(t)
	})

	t.Run("empty Content-Type, Body close failure", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {""},
		}

		body := newMockBody("")
		body.closeErr = errors.New("test_error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		resp.NoContent()
		resp.chain.assertFailed(t)
	})
}

func TestResponse_ContentType(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain; charset=utf-8"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.ContentType("text/plain")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "utf-8")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "UTF-8")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("bad")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "bad")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("empty type", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"charset=utf-8"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.ContentType("")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("", "")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("empty charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.ContentType("text/plain")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "utf-8")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})

	t.Run("empty type and charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {";"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.ContentType("")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("", "")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})
}

func TestResponse_ContentEncoding(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Encoding": {"gzip", "deflate"},
	}

	resp := NewResponse(reporter, &http.Response{
		Header: http.Header(headers),
	})

	resp.ContentEncoding("gzip", "deflate")
	resp.chain.assertNotFailed(t)
	resp.chain.clearFailed()

	resp.ContentEncoding("deflate", "gzip")
	resp.chain.assertFailed(t)
	resp.chain.clearFailed()

	resp.ContentEncoding("gzip")
	resp.chain.assertFailed(t)
	resp.chain.clearFailed()

	resp.ContentEncoding()
	resp.chain.assertFailed(t)
	resp.chain.clearFailed()
}

func TestResponse_TransferEncoding(t *testing.T) {
	reporter := newMockReporter(t)

	resp := NewResponse(reporter, &http.Response{
		TransferEncoding: []string{"foo", "bar"},
	})

	resp.TransferEncoding("foo", "bar")
	resp.chain.assertNotFailed(t)
	resp.chain.clearFailed()

	resp.TransferEncoding("foo")
	resp.chain.assertFailed(t)
	resp.chain.clearFailed()

	resp.TransferEncoding()
	resp.chain.assertFailed(t)
	resp.chain.clearFailed()
}

func TestResponse_Text(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain; charset=utf-8"},
		}

		body := `hello, world!`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain", "utf-8")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("application/json")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.Text()
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t, "hello, world!", resp.Text().Raw())
	})

	t.Run("read failure", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain; charset=utf-8"},
		}

		body := newMockBody(`hello, world!`)
		body.readErr = errors.New("read error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		respText := resp.Text()

		assert.Equal(t, "", respText.Raw())
		respText.chain.assertFailed(t)
		resp.chain.assertFailed(t)
	})
}

func TestResponse_Form(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		body := `a=1&b=2`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("application/x-www-form-urlencoded")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("application/x-www-form-urlencoded", "")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.Form()
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		expected := map[string]interface{}{
			"a": "1",
			"b": "2",
		}

		assert.Equal(t, expected, resp.Form().Raw())
	})

	t.Run("bad body", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		body := "%"

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.Form()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		assert.Nil(t, resp.Form().Raw())
	})

	t.Run("bad type", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"bad"},
		}

		body := "foo=bar"

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.Form()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		assert.Nil(t, resp.Form().Raw())
	})

	t.Run("read failure", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		body := newMockBody("foo=bar")
		body.readErr = errors.New("read error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		respForm := resp.Form()
		assert.Nil(t, respForm.Raw())

		respForm.chain.assertFailed(t)
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})
}

func TestResponse_JSON(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/json; charset=utf-8"},
		}

		body := `{"key": "value"}`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("application/json")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("application/json", "utf-8")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		resp.ContentType("text/plain")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		resp.JSON()
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t,
			map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
	})

	t.Run("bad body", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/json; charset=utf-8"},
		}

		body := "{"

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSON()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		assert.Nil(t, resp.JSON().Raw())
	})

	t.Run("empty charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/json"},
		}

		body := `{"key": "value"}`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSON()
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t,
			map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
	})

	t.Run("bad charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/json; charset=bad"},
		}

		body := `{"key": "value"}`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSON()
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t, nil, resp.JSON().Raw())
	})

	t.Run("read failure", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/json; charset=utf-8"},
		}

		body := newMockBody(`{"key": "value"}`)
		body.readErr = errors.New("read error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		respJSON := resp.JSON()
		assert.Nil(t, respJSON.Raw())
		respJSON.chain.assertFailed(t)
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})
}

func TestResponse_JSONP(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/javascript; charset=utf-8"},
		}

		body1 := `foo({"key": "value"})`
		body2 := `foo({"key": "value"});`
		body3 := ` foo ( {"key": "value"} ) ; `

		for n, body := range []string{body1, body2, body3} {
			t.Run(fmt.Sprintf("body%d", n+1),
				func(t *testing.T) {
					httpResp := &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header(headers),
						Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
					}

					resp := NewResponse(reporter, httpResp)

					assert.Equal(t, body, resp.Body().Raw())
					resp.chain.assertNotFailed(t)
					resp.chain.clearFailed()

					resp.ContentType("application/javascript")
					resp.chain.assertNotFailed(t)
					resp.chain.clearFailed()

					resp.ContentType("application/javascript", "utf-8")
					resp.chain.assertNotFailed(t)
					resp.chain.clearFailed()

					resp.ContentType("text/plain")
					resp.chain.assertFailed(t)
					resp.chain.clearFailed()

					resp.JSONP("foo")
					resp.chain.assertNotFailed(t)
					resp.chain.clearFailed()

					assert.Equal(t,
						map[string]interface{}{"key": "value"},
						resp.JSONP("foo").Object().Raw())

					resp.JSONP("fo")
					resp.chain.assertFailed(t)
					resp.chain.clearFailed()

					resp.JSONP("")
					resp.chain.assertFailed(t)
					resp.chain.clearFailed()
				})
		}
	})

	t.Run("bad body", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/javascript; charset=utf-8"},
		}

		body1 := `foo`
		body2 := `foo();`
		body3 := `foo(`
		body4 := `foo({);`

		for n, body := range []string{body1, body2, body3, body4} {
			t.Run(fmt.Sprintf("body%d", n+1),
				func(t *testing.T) {
					httpResp := &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header(headers),
						Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
					}

					resp := NewResponse(reporter, httpResp)

					resp.JSONP("foo")
					resp.chain.assertFailed(t)
					resp.chain.clearFailed()

					assert.Nil(t, resp.JSONP("foo").Raw())
				})
		}
	})

	t.Run("empty charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/javascript"},
		}

		body := `foo({"key": "value"})`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSONP("foo")
		resp.chain.assertNotFailed(t)
		resp.chain.clearFailed()

		assert.Equal(t,
			map[string]interface{}{"key": "value"}, resp.JSONP("foo").Object().Raw())
	})

	t.Run("bad charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/javascript; charset=bad"},
		}

		body := `foo({"key": "value"})`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSONP("foo")
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()

		assert.Nil(t, resp.JSONP("foo").Raw())
	})

	t.Run("read failure", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/javascript; charset=utf-8"},
		}

		body := newMockBody(`foo({"key": "value"})`)
		body.readErr = errors.New("read error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		respJSONP := resp.JSONP("foo")
		assert.Nil(t, respJSONP.Raw())
		respJSONP.chain.assertFailed(t)
		resp.chain.assertFailed(t)
		resp.chain.clearFailed()
	})
}

func TestResponse_ContentOpts(t *testing.T) {
	reporter := newMockReporter(t)

	type testCase struct {
		respContentType   string
		respBody          string
		expectedMediaType string
		expectedCharset   string
		match             bool
		chainFunc         func(*Response, ContentOpts) *chain
	}

	runTest := func(tc testCase) {
		headers := map[string][]string{
			"Content-Type": {tc.respContentType},
		}

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(tc.respBody)),
		}

		resp := NewResponse(reporter, httpResp)

		c := tc.chainFunc(resp, ContentOpts{
			MediaType: tc.expectedMediaType,
			Charset:   tc.expectedCharset,
		})

		if tc.match {
			c.assertNotFailed(t)
		} else {
			c.assertFailed(t)
		}
	}

	check := func(
		defaultType, defaultCharset, respBody string,
		chainFunc func(*Response, ContentOpts) *chain,
	) {
		runTest(testCase{
			respContentType:   "test-type; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "test-charset",
			match:             true,
			chainFunc:         chainFunc,
		})
		runTest(testCase{
			respContentType:   "test-type; charset=BAD",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "test-charset",
			match:             false,
			chainFunc:         chainFunc,
		})
		runTest(testCase{
			respContentType:   "BAD; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "test-charset",
			match:             false,
			chainFunc:         chainFunc,
		})
		if defaultCharset != "" {
			runTest(testCase{
				respContentType:   "test-type; charset=" + defaultCharset,
				respBody:          respBody,
				expectedMediaType: "test-type",
				expectedCharset:   defaultCharset,
				match:             true,
				chainFunc:         chainFunc,
			})
			runTest(testCase{
				respContentType:   "test-type; charset=" + defaultCharset,
				respBody:          respBody,
				expectedMediaType: "test-type",
				expectedCharset:   "",
				match:             true,
				chainFunc:         chainFunc,
			})
		}
		runTest(testCase{
			respContentType:   "test-type",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "",
			match:             true,
			chainFunc:         chainFunc,
		})
		runTest(testCase{
			respContentType:   defaultType + "; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: defaultType,
			expectedCharset:   "test-charset",
			match:             true,
			chainFunc:         chainFunc,
		})
		runTest(testCase{
			respContentType:   defaultType + "; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: "",
			expectedCharset:   "test-charset",
			match:             true,
			chainFunc:         chainFunc,
		})
	}

	t.Run("text", func(t *testing.T) {
		check("text/plain",
			"utf-8",
			"test text",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.Text(opts).chain
			})
	})

	t.Run("form", func(t *testing.T) {
		check("application/x-www-form-urlencoded",
			"",
			"a=b",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.Form(opts).chain
			})
	})

	t.Run("json", func(t *testing.T) {
		check("application/json",
			"utf-8",
			"{}",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.JSON(opts).chain
			})
	})

	t.Run("jsonp", func(t *testing.T) {
		check("application/javascript",
			"utf-8",
			"cb({})",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.JSONP("cb", opts).chain
			})
	})
}

func TestResponse_Usage(t *testing.T) {
	t.Run("NewResponse multiple rtt arguments", func(t *testing.T) {
		reporter := newMockReporter(t)
		rtt := []time.Duration{time.Second, time.Second}
		resp := NewResponse(reporter, &http.Response{}, rtt...)
		resp.chain.assertFailed(t)
	})

	t.Run("ContentType multiple charset arguments", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain;charset=utf-8;charset=US-ASCII"},
		}
		resp := NewResponse(reporter, &http.Response{
			Header: headers,
		})
		resp.ContentType("text/plain", "utf-8", "US-ASCII")
		resp.chain.assertFailed(t)
	})

	t.Run("Text multiple arguments", func(t *testing.T) {
		reporter := newMockReporter(t)
		header := map[string][]string{
			"ContentType": {"text/plain"},
		}
		resp := NewResponse(reporter, &http.Response{
			Header: header,
		})
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.Text(contentOpts1, contentOpts2)
		resp.chain.assertFailed(t)
	})

	t.Run("Form multiple arguments", func(t *testing.T) {
		reporter := newMockReporter(t)
		headers := map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}

		body := `a=1&b=2`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.Form(contentOpts1, contentOpts2)
		resp.chain.assertFailed(t)

	})

	t.Run("JSON multiple arguments", func(t *testing.T) {
		reporter := newMockReporter(t)
		headers := map[string][]string{
			"Content-Type": {"application/json; charset=utf-8"},
		}

		body := `{"key": "value"}`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.JSON(contentOpts1, contentOpts2)
		resp.chain.assertFailed(t)
	})

	t.Run("JSONP multiple arguments", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"application/javascript; charset=utf-8"},
		}

		body1 := `foo({"key": "value"})`

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body1)),
		}

		resp := NewResponse(reporter, httpResp)
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.JSONP("foo", contentOpts1, contentOpts2)
		resp.chain.assertFailed(t)
	})
}
