package httpexpect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_FailedChain(t *testing.T) {
	check := func(resp *Response) {
		resp.chain.assert(t, failure)

		resp.Alias("foo")

		resp.RoundTripTime().chain.assert(t, failure)
		resp.Duration().chain.assert(t, failure)
		resp.Headers().chain.assert(t, failure)
		resp.Header("foo").chain.assert(t, failure)
		resp.Cookies().chain.assert(t, failure)
		resp.Cookie("foo").chain.assert(t, failure)
		resp.Body().chain.assert(t, failure)
		resp.Text().chain.assert(t, failure)
		resp.Form().chain.assert(t, failure)
		resp.JSON().chain.assert(t, failure)
		resp.JSONP("").chain.assert(t, failure)
		resp.Websocket().chain.assert(t, failure)

		resp.Status(123)
		resp.StatusRange(Status2xx)
		resp.StatusList(http.StatusOK, http.StatusBadGateway)
		resp.NoContent()
		resp.HasContentType("", "")
		resp.HasContentEncoding("")
		resp.HasTransferEncoding("")
	}

	t.Run("failed chain", func(t *testing.T) {
		reporter := newMockReporter(t)
		config := newMockConfig(reporter)
		chain := newChainWithDefaults("test", reporter, flagFailed)

		resp := newResponse(responseOpts{
			config:   config,
			chain:    chain,
			httpResp: &http.Response{},
		})

		check(resp)
	})

	t.Run("nil value", func(t *testing.T) {
		reporter := newMockReporter(t)
		config := newMockConfig(reporter)
		chain := newChainWithDefaults("test", reporter)

		resp := newResponse(responseOpts{
			config:   config,
			chain:    chain,
			httpResp: nil,
		})

		check(resp)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
		reporter := newMockReporter(t)
		config := newMockConfig(reporter)
		chain := newChainWithDefaults("test", reporter, flagFailed)

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
		resp.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		resp := NewResponseC(Config{
			Reporter: reporter,
		}, &http.Response{})
		resp.chain.assert(t, success)
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

func TestResponse_Raw(t *testing.T) {
	reporter := newMockReporter(t)

	httpResp := http.Response{}

	resp := NewResponse(reporter, &httpResp)

	assert.Same(t, &httpResp, resp.Raw())
	resp.chain.assert(t, success)
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
	t.Run("provided", func(t *testing.T) {
		duration := time.Second

		reporter := newMockReporter(t)
		resp := NewResponse(reporter, &http.Response{}, duration)
		resp.chain.assert(t, success)

		rt := resp.RoundTripTime()
		resp.chain.assert(t, success)
		rt.chain.assert(t, success)

		assert.Equal(t, time.Second, rt.Raw())
	})

	t.Run("omitted", func(t *testing.T) {
		reporter := newMockReporter(t)
		resp := NewResponse(reporter, &http.Response{})
		resp.chain.assert(t, success)

		rt := resp.RoundTripTime()
		resp.chain.assert(t, success)
		rt.chain.assert(t, success)

		assert.Equal(t, time.Duration(0), rt.Raw())
	})
}

func TestResponse_Status(t *testing.T) {
	reporter := newMockReporter(t)

	cases := []struct {
		status     int
		testStatus int
	}{
		{http.StatusOK, http.StatusOK},
		{http.StatusOK, http.StatusNotFound},
		{http.StatusNotFound, http.StatusNotFound},
		{http.StatusNotFound, http.StatusOK},
	}

	for _, tc := range cases {
		resp := NewResponse(reporter, &http.Response{
			StatusCode: tc.status,
		})

		resp.Status(tc.testStatus)

		if tc.status == tc.testStatus {
			resp.chain.assert(t, success)
		} else {
			resp.chain.assert(t, failure)
		}
	}
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
		status      int
		statusRange StatusRange
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

	for _, tc := range cases {
		for _, r := range ranges {
			resp := NewResponse(reporter, &http.Response{
				StatusCode: tc.status,
			})

			resp.StatusRange(r)

			if tc.statusRange == r {
				resp.chain.assert(t, success)
			} else {
				resp.chain.assert(t, failure)
			}
		}
	}
}

func TestResponse_StatusList(t *testing.T) {
	reporter := newMockReporter(t)

	cases := []struct {
		status     int
		statusList []int
		result     chainResult
	}{
		{
			http.StatusOK,
			[]int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError},
			success,
		},
		{
			http.StatusBadRequest,
			[]int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError},
			success,
		},
		{
			http.StatusOK,
			[]int{http.StatusInternalServerError, http.StatusBadRequest},
			failure,
		},
		{
			http.StatusBadGateway,
			[]int{},
			failure,
		},
	}

	for _, tc := range cases {
		resp := NewResponse(reporter, &http.Response{
			StatusCode: tc.status,
		})
		resp.StatusList(tc.statusList...)
		resp.chain.assert(t, tc.result)
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

	resp.Headers().IsEqual(headers).
		chain.assert(t, success)

	for k, v := range headers {
		for _, h := range []string{k, strings.ToLower(k), strings.ToUpper(k)} {
			resp.Header(h).IsEqual(v[0]).
				chain.assert(t, success)
		}
	}

	resp.Header("Bad-Header").IsEmpty().
		chain.assert(t, success)
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
		resp.chain.assert(t, success)

		assert.Equal(t, []interface{}{"foo", "bar"}, resp.Cookies().Raw())
		resp.chain.assert(t, success)

		c1 := resp.Cookie("foo")
		resp.chain.assert(t, success)
		assert.Equal(t, "foo", c1.Raw().Name)
		assert.Equal(t, "aaa", c1.Raw().Value)
		assert.Equal(t, "", c1.Raw().Domain)
		assert.Equal(t, "", c1.Raw().Path)

		c2 := resp.Cookie("bar")
		resp.chain.assert(t, success)
		assert.Equal(t, "bar", c2.Raw().Name)
		assert.Equal(t, "bbb", c2.Raw().Value)
		assert.Equal(t, "example.com", c2.Raw().Domain)
		assert.Equal(t, "/xxx", c2.Raw().Path)
		assert.True(t, time.Date(2010, 12, 31, 23, 59, 59, 0, time.UTC).
			Equal(c2.Raw().Expires))

		c3 := resp.Cookie("baz")
		resp.chain.assert(t, failure)
		c3.chain.assert(t, failure)
		assert.Nil(t, c3.Raw())
	})

	t.Run("no cookies", func(t *testing.T) {
		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     nil,
			Body:       nil,
		}

		resp := NewResponse(reporter, httpResp)
		resp.chain.assert(t, success)

		assert.Equal(t, []interface{}{}, resp.Cookies().Raw())
		resp.chain.assert(t, success)

		c := resp.Cookie("foo")
		resp.chain.assert(t, failure)
		c.chain.assert(t, failure)
		assert.Nil(t, c.Raw())
	})
}

func TestResponse_BodyOperations(t *testing.T) {
	t.Run("content", func(t *testing.T) {
		reporter := newMockReporter(t)

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("body")),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "body", resp.Body().Raw())
		resp.chain.assert(t, success)
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

		resp.chain.assert(t, success)
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

		respBody.chain.assert(t, failure)
		resp.chain.assert(t, failure)
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

		resp.chain.assert(t, failure)
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
		resp.chain.assert(t, success)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, []byte("body string"), resp.content)
		assert.Equal(t, contentRetreived, resp.contentState)

		// Second call should be no-op
		resp.Body()
		resp.chain.assert(t, success)

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
		resp.chain.assert(t, failure)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)

		// Second call should be no-op
		resp.Body()
		resp.chain.assert(t, failure)

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
		resp.chain.assert(t, failure)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)

		// Second call should be no-op
		resp.Body()
		resp.chain.assert(t, failure)

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
		resp.chain.assert(t, failure)

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)

		// Invoke getContent()
		chain := resp.chain.enter("Test()")
		content, ok := resp.getContent(chain, "Test()")

		chain.assert(t, failure)
		assert.Nil(t, content)
		assert.False(t, ok)

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentFailed, resp.contentState)
	})

	t.Run("hijacked state", func(t *testing.T) {
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

		// Hijack body
		reader := resp.Reader()
		assert.NotNil(t, reader)
		resp.chain.assert(t, success)

		// Invoke getContent()
		chain := resp.chain.enter("Test()")
		content, ok := resp.getContent(chain, "Test()")

		chain.assert(t, failure)
		assert.Nil(t, content)
		assert.False(t, ok)

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Nil(t, resp.content)
		assert.Equal(t, contentHijacked, resp.contentState)
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
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "", resp.Body().Raw())
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.NoContent()
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.Text()
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.Form()
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.JSON()
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.JSONP("")
		resp.chain.assert(t, failure)
		resp.chain.clear()
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
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.NoContent()
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.Text()
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.Form()
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.JSON()
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.JSONP("")
		resp.chain.assert(t, failure)
		resp.chain.clear()
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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.NoContent()
		resp.chain.assert(t, failure)
		resp.chain.clear()
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
		resp.chain.assert(t, failure)
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
		resp.chain.assert(t, failure)
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

		resp.HasContentType("text/plain")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain", "utf-8")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain", "UTF-8")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("bad")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.HasContentType("text/plain", "bad")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.HasContentType("")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.HasContentType("text/plain", "")
		resp.chain.assert(t, failure)
		resp.chain.clear()
	})

	t.Run("empty type", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"charset=utf-8"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.HasContentType("")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.HasContentType("", "")
		resp.chain.assert(t, failure)
		resp.chain.clear()
	})

	t.Run("empty charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.HasContentType("text/plain")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain", "")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain", "utf-8")
		resp.chain.assert(t, failure)
		resp.chain.clear()
	})

	t.Run("empty type and charset", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {";"},
		}

		resp := NewResponse(reporter, &http.Response{
			Header: http.Header(headers),
		})

		resp.HasContentType("")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.HasContentType("", "")
		resp.chain.assert(t, failure)
		resp.chain.clear()
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

	resp.HasContentEncoding("gzip", "deflate")
	resp.chain.assert(t, success)
	resp.chain.clear()

	resp.HasContentEncoding("deflate", "gzip")
	resp.chain.assert(t, failure)
	resp.chain.clear()

	resp.HasContentEncoding("gzip")
	resp.chain.assert(t, failure)
	resp.chain.clear()

	resp.HasContentEncoding()
	resp.chain.assert(t, failure)
	resp.chain.clear()
}

func TestResponse_TransferEncoding(t *testing.T) {
	reporter := newMockReporter(t)

	resp := NewResponse(reporter, &http.Response{
		TransferEncoding: []string{"foo", "bar"},
	})

	resp.HasTransferEncoding("foo", "bar")
	resp.chain.assert(t, success)
	resp.chain.clear()

	resp.HasTransferEncoding("foo")
	resp.chain.assert(t, failure)
	resp.chain.clear()

	resp.HasTransferEncoding()
	resp.chain.assert(t, failure)
	resp.chain.clear()
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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain", "utf-8")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("application/json")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.Text()
		resp.chain.assert(t, success)
		resp.chain.clear()

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
		respText.chain.assert(t, failure)
		resp.chain.assert(t, failure)
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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("application/x-www-form-urlencoded")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("application/x-www-form-urlencoded", "")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.Form()
		resp.chain.assert(t, success)
		resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.Form()
		resp.chain.assert(t, failure)
		resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.Form()
		resp.chain.assert(t, failure)
		resp.chain.clear()

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

		respForm.chain.assert(t, failure)
		resp.chain.assert(t, failure)
		resp.chain.clear()
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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("application/json")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("application/json", "utf-8")
		resp.chain.assert(t, success)
		resp.chain.clear()

		resp.HasContentType("text/plain")
		resp.chain.assert(t, failure)
		resp.chain.clear()

		resp.JSON()
		resp.chain.assert(t, success)
		resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSON()
		resp.chain.assert(t, failure)
		resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSON()
		resp.chain.assert(t, success)
		resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSON()
		resp.chain.assert(t, failure)
		resp.chain.clear()

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
		respJSON.chain.assert(t, failure)
		resp.chain.assert(t, failure)
		resp.chain.clear()
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
						Body:       io.NopCloser(bytes.NewBufferString(body)),
					}

					resp := NewResponse(reporter, httpResp)

					assert.Equal(t, body, resp.Body().Raw())
					resp.chain.assert(t, success)
					resp.chain.clear()

					resp.HasContentType("application/javascript")
					resp.chain.assert(t, success)
					resp.chain.clear()

					resp.HasContentType("application/javascript", "utf-8")
					resp.chain.assert(t, success)
					resp.chain.clear()

					resp.HasContentType("text/plain")
					resp.chain.assert(t, failure)
					resp.chain.clear()

					resp.JSONP("foo")
					resp.chain.assert(t, success)
					resp.chain.clear()

					assert.Equal(t,
						map[string]interface{}{"key": "value"},
						resp.JSONP("foo").Object().Raw())

					resp.JSONP("fo")
					resp.chain.assert(t, failure)
					resp.chain.clear()

					resp.JSONP("")
					resp.chain.assert(t, failure)
					resp.chain.clear()
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
						Body:       io.NopCloser(bytes.NewBufferString(body)),
					}

					resp := NewResponse(reporter, httpResp)

					resp.JSONP("foo")
					resp.chain.assert(t, failure)
					resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSONP("foo")
		resp.chain.assert(t, success)
		resp.chain.clear()

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSONP("foo")
		resp.chain.assert(t, failure)
		resp.chain.clear()

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
		respJSONP.chain.assert(t, failure)
		resp.chain.assert(t, failure)
		resp.chain.clear()
	})
}

func TestResponse_ContentOpts(t *testing.T) {
	type testCase struct {
		respContentType   string
		respBody          string
		expectedMediaType string
		expectedCharset   string
		match             bool
		chainFunc         func(*Response, ContentOpts) *chain
	}

	runTest := func(t *testing.T, tc testCase) {
		headers := map[string][]string{
			"Content-Type": {tc.respContentType},
		}

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       io.NopCloser(bytes.NewBufferString(tc.respBody)),
		}

		reporter := newMockReporter(t)
		resp := NewResponse(reporter, httpResp)

		c := tc.chainFunc(resp, ContentOpts{
			MediaType: tc.expectedMediaType,
			Charset:   tc.expectedCharset,
		})

		if tc.match {
			c.assert(t, success)
		} else {
			c.assert(t, failure)
		}
	}

	runAllTests := func(
		t *testing.T,
		defaultType, defaultCharset, respBody string,
		chainFunc func(*Response, ContentOpts) *chain,
	) {
		runTest(t, testCase{
			respContentType:   "test-type; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "test-charset",
			match:             true,
			chainFunc:         chainFunc,
		})
		runTest(t, testCase{
			respContentType:   "test-type; charset=BAD",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "test-charset",
			match:             false,
			chainFunc:         chainFunc,
		})
		runTest(t, testCase{
			respContentType:   "BAD; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "test-charset",
			match:             false,
			chainFunc:         chainFunc,
		})
		if defaultCharset != "" {
			runTest(t, testCase{
				respContentType:   "test-type; charset=" + defaultCharset,
				respBody:          respBody,
				expectedMediaType: "test-type",
				expectedCharset:   defaultCharset,
				match:             true,
				chainFunc:         chainFunc,
			})
			runTest(t, testCase{
				respContentType:   "test-type; charset=" + defaultCharset,
				respBody:          respBody,
				expectedMediaType: "test-type",
				expectedCharset:   "",
				match:             true,
				chainFunc:         chainFunc,
			})
		}
		runTest(t, testCase{
			respContentType:   "test-type",
			respBody:          respBody,
			expectedMediaType: "test-type",
			expectedCharset:   "",
			match:             true,
			chainFunc:         chainFunc,
		})
		runTest(t, testCase{
			respContentType:   defaultType + "; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: defaultType,
			expectedCharset:   "test-charset",
			match:             true,
			chainFunc:         chainFunc,
		})
		runTest(t, testCase{
			respContentType:   defaultType + "; charset=test-charset",
			respBody:          respBody,
			expectedMediaType: "",
			expectedCharset:   "test-charset",
			match:             true,
			chainFunc:         chainFunc,
		})
	}

	t.Run("text", func(t *testing.T) {
		runAllTests(t, "text/plain",
			"utf-8",
			"test text",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.Text(opts).chain
			})
	})

	t.Run("form", func(t *testing.T) {
		runAllTests(t, "application/x-www-form-urlencoded",
			"",
			"a=b",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.Form(opts).chain
			})
	})

	t.Run("json", func(t *testing.T) {
		runAllTests(t, "application/json",
			"utf-8",
			"{}",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.JSON(opts).chain
			})
	})

	t.Run("jsonp", func(t *testing.T) {
		runAllTests(t, "application/javascript",
			"utf-8",
			"cb({})",
			func(resp *Response, opts ContentOpts) *chain {
				return resp.JSONP("cb", opts).chain
			})
	})
}

func TestResponse_Reader(t *testing.T) {
	t.Run("read body", func(t *testing.T) {
		reporter := newMockReporter(t)
		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       newMockBody("test body"),
		}
		resp := NewResponse(reporter, httpResp)

		reader := resp.Reader()
		require.NotNil(t, reader)
		resp.chain.assert(t, success)

		b, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, "test body", string(b))

		err = reader.Close()
		assert.NoError(t, err)
	})

	t.Run("rewinds disabled", func(t *testing.T) {
		wrp := newBodyWrapper(newMockBody("test"), nil)

		reporter := newMockReporter(t)
		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       wrp,
		}
		resp := NewResponse(reporter, httpResp)

		assert.False(t, wrp.isRewindDisabled)

		reader := resp.Reader()
		require.NotNil(t, reader)
		resp.chain.assert(t, success)

		assert.True(t, wrp.isRewindDisabled)

		err := reader.Close()
		assert.NoError(t, err)
	})

	t.Run("conflicts", func(t *testing.T) {
		cases := []struct {
			name        string
			contentType string
			body        string
			method      func(resp *Response) *chain
		}{
			{
				name:        "Body",
				contentType: "text/plain; charset=utf-8",
				body:        `test`,
				method: func(resp *Response) *chain {
					return resp.Body().chain
				},
			},
			{
				name:        "Text",
				contentType: "text/plain; charset=utf-8",
				body:        `test`,
				method: func(resp *Response) *chain {
					return resp.Text().chain
				},
			},
			{
				name:        "Form",
				contentType: "application/x-www-form-urlencoded",
				body:        `x=1&y=0`,
				method: func(resp *Response) *chain {
					return resp.Form().chain
				},
			},
			{
				name:        "JSON",
				contentType: "application/json",
				body:        `{"x":"y"}`,
				method: func(resp *Response) *chain {
					return resp.JSON().chain
				},
			},
			{
				name:        "JSONP",
				contentType: "application/javascript",
				body:        `test({"x":"y"})`,
				method: func(resp *Response) *chain {
					return resp.JSONP("test").chain
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Run("before reader", func(t *testing.T) {
					reporter := newMockReporter(t)
					httpResp := &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header(map[string][]string{
							"Content-Type": {tc.contentType},
						}),
						Body: newMockBody(tc.body),
					}
					resp := NewResponse(reporter, httpResp)

					// other method reads body
					tc.method(resp).assert(t, success)

					// Reader will fail
					reader := resp.Reader()
					require.NotNil(t, reader)
					resp.chain.assert(t, failure)

					b, err := io.ReadAll(reader)
					assert.Error(t, err)
					assert.Empty(t, b)

					err = reader.Close()
					assert.Error(t, err)
				})
			})

			t.Run(tc.name, func(t *testing.T) {
				t.Run("after reader", func(t *testing.T) {
					reporter := newMockReporter(t)
					httpResp := &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header(map[string][]string{
							"Content-Type": {tc.contentType},
						}),
						Body: newMockBody(tc.body),
					}
					resp := NewResponse(reporter, httpResp)

					// Reader hijacks body
					reader := resp.Reader()
					require.NotNil(t, reader)
					resp.chain.assert(t, success)

					// other method will fail
					tc.method(resp).assert(t, failure)
					resp.chain.assert(t, failure)

					b, err := io.ReadAll(reader)
					assert.NoError(t, err)
					assert.Equal(t, tc.body, string(b))

					err = reader.Close()
					assert.NoError(t, err)
				})
			})
		}
	})
}

func TestResponse_Usage(t *testing.T) {
	t.Run("NewResponse multiple rtt arguments", func(t *testing.T) {
		reporter := newMockReporter(t)
		rtt := []time.Duration{time.Second, time.Second}
		resp := NewResponse(reporter, &http.Response{}, rtt...)
		resp.chain.assert(t, failure)
	})

	t.Run("ContentType multiple charset arguments", func(t *testing.T) {
		reporter := newMockReporter(t)

		headers := map[string][]string{
			"Content-Type": {"text/plain;charset=utf-8;charset=US-ASCII"},
		}
		resp := NewResponse(reporter, &http.Response{
			Header: headers,
		})
		resp.HasContentType("text/plain", "utf-8", "US-ASCII")
		resp.chain.assert(t, failure)
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
		resp.chain.assert(t, failure)
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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.Form(contentOpts1, contentOpts2)
		resp.chain.assert(t, failure)

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
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.JSON(contentOpts1, contentOpts2)
		resp.chain.assert(t, failure)
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
			Body:       io.NopCloser(bytes.NewBufferString(body1)),
		}

		resp := NewResponse(reporter, httpResp)
		contentOpts1 := ContentOpts{
			MediaType: "text/plain",
		}
		contentOpts2 := ContentOpts{
			MediaType: "application/json",
		}
		resp.JSONP("foo", contentOpts1, contentOpts2)
		resp.chain.assert(t, failure)
	})
}
