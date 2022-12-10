package httpexpect

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponseFailed(t *testing.T) {
	check := func(resp *Response) {
		resp.chain.assertFailed(t)

		assert.NotNil(t, resp.RoundTripTime())
		assert.NotNil(t, resp.Duration())
		assert.NotNil(t, resp.Headers())
		assert.NotNil(t, resp.Header("foo"))
		assert.NotNil(t, resp.Cookies())
		assert.NotNil(t, resp.Cookie("foo"))
		assert.NotNil(t, resp.Body())
		assert.NotNil(t, resp.Text())
		assert.NotNil(t, resp.Form())
		assert.NotNil(t, resp.JSON())
		assert.NotNil(t, resp.JSONP(""))
		assert.NotNil(t, resp.Websocket())

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
		resp.NoContent()
		resp.ContentType("", "")
		resp.ContentEncoding("")
		resp.TransferEncoding("")
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(mockFailure())

		resp := newResponse(responseOpts{
			chain:    chain,
			httpResp: &http.Response{},
		})

		check(resp)
	})

	t.Run("nil_value", func(t *testing.T) {
		chain := newMockChain(t)

		resp := newResponse(responseOpts{
			chain:    chain,
			httpResp: nil,
		})

		check(resp)
	})

	t.Run("failed_chain_nil_value", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(mockFailure())

		resp := newResponse(responseOpts{
			chain:    chain,
			httpResp: nil,
		})

		check(resp)
	})
}

func TestResponseRoundTripTime(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("set", func(t *testing.T) {
		duration := time.Second

		resp := NewResponse(reporter, &http.Response{}, duration)
		resp.chain.assertOK(t)
		resp.chain.reset()

		rt := resp.RoundTripTime()

		assert.Equal(t, time.Second, rt.Raw())

		rt.IsSet()
		rt.Equal(time.Second)
		rt.chain.assertOK(t)
	})

	t.Run("unset", func(t *testing.T) {
		resp := NewResponse(reporter, &http.Response{})
		resp.chain.assertOK(t)
		resp.chain.reset()

		rt := resp.RoundTripTime()

		assert.Equal(t, time.Duration(0), rt.Raw())

		rt.NotSet()
		rt.chain.assertOK(t)

		rt.IsSet()
		rt.chain.assertFailed(t)
	})
}

func TestResponseDuration(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("set", func(t *testing.T) {
		duration := time.Second

		resp := NewResponse(reporter, &http.Response{}, duration)
		resp.chain.assertOK(t)
		resp.chain.reset()

		d := resp.Duration()

		assert.Equal(t, float64(time.Second), d.Raw())

		d.chain.assertOK(t)
	})

	t.Run("unset", func(t *testing.T) {
		resp := NewResponse(reporter, &http.Response{})
		resp.chain.assertOK(t)
		resp.chain.reset()

		d := resp.Duration()

		assert.Equal(t, float64(0), d.Raw())

		d.chain.assertOK(t)
	})
}

func TestResponseStatusRange(t *testing.T) {
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
				resp.chain.assertOK(t)
			} else {
				resp.chain.assertFailed(t)
			}
		}
	}
}

func TestResponseHeaders(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Same(t, httpResp, resp.Raw())

	resp.Status(http.StatusOK)
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.Status(http.StatusNotFound)
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Headers().Equal(headers).chain.assertOK(t)

	for k, v := range headers {
		for _, h := range []string{k, strings.ToLower(k), strings.ToUpper(k)} {
			resp.Header(h).Equal(v[0]).chain.assertOK(t)
		}
	}

	resp.Header("Bad-Header").Empty().chain.assertOK(t)
}

func TestResponseCookies(t *testing.T) {
	reporter := newMockReporter(t)

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
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t, []interface{}{"foo", "bar"}, resp.Cookies().Raw())
	resp.chain.assertOK(t)

	c1 := resp.Cookie("foo")
	resp.chain.assertOK(t)
	assert.Equal(t, "foo", c1.Raw().Name)
	assert.Equal(t, "aaa", c1.Raw().Value)
	assert.Equal(t, "", c1.Raw().Domain)
	assert.Equal(t, "", c1.Raw().Path)

	c2 := resp.Cookie("bar")
	resp.chain.assertOK(t)
	assert.Equal(t, "bar", c2.Raw().Name)
	assert.Equal(t, "bbb", c2.Raw().Value)
	assert.Equal(t, "example.com", c2.Raw().Domain)
	assert.Equal(t, "/xxx", c2.Raw().Path)
	assert.True(t, time.Date(2010, 12, 31, 23, 59, 59, 0, time.UTC).
		Equal(c2.Raw().Expires))

	c3 := resp.Cookie("baz")
	resp.chain.assertFailed(t)
	c3.chain.assertFailed(t)
	assert.True(t, c3.Raw() == nil)
}

func TestResponseNoCookies(t *testing.T) {
	reporter := newMockReporter(t)

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     nil,
		Body:       nil,
	}

	resp := NewResponse(reporter, httpResp)
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t, []interface{}{}, resp.Cookies().Raw())
	resp.chain.assertOK(t)

	c := resp.Cookie("foo")
	resp.chain.assertFailed(t)
	c.chain.assertFailed(t)
	assert.True(t, c.Raw() == nil)
}

func TestResponseBody(t *testing.T) {
	reporter := newMockReporter(t)

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString("body")),
	}

	resp := NewResponse(reporter, httpResp)

	assert.Equal(t, "body", resp.Body().Raw())
	resp.chain.assertOK(t)
	resp.chain.reset()
}

func TestResponseBodyClose(t *testing.T) {
	reporter := newMockReporter(t)

	body := newMockBody("test_body")

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       body,
	}

	resp := NewResponse(reporter, httpResp)

	assert.Equal(t, "test_body", resp.Body().Raw())
	assert.True(t, body.closed)

	resp.chain.assertOK(t)
}

func TestResponseBodyError(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("read_err", func(t *testing.T) {
		body := newMockBody("test_body")
		body.readErr = errors.New("test_error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "", resp.Body().Raw())
		assert.True(t, body.closed)

		resp.chain.assertFailed(t)
	})

	t.Run("close_err", func(t *testing.T) {
		body := newMockBody("test_body")
		body.closeErr = errors.New("test_error")

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, "", resp.Body().Raw())
		assert.True(t, body.closed)

		resp.chain.assertFailed(t)
	})
}

func TestResponseNoContentEmpty(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.NoContent()
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.Text()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Form()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSONP("")
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseNoContentNil(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.NoContent()
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.Text()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Form()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSONP("")
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseNoContentFailed(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.NoContent()
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseContentType(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"text/plain; charset=utf-8"},
	}

	resp := NewResponse(reporter, &http.Response{
		Header: http.Header(headers),
	})

	resp.ContentType("text/plain")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "utf-8")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "UTF-8")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("bad")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "bad")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentType("")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "")
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseContentTypeEmptyCharset(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	resp := NewResponse(reporter, &http.Response{
		Header: http.Header(headers),
	})

	resp.ContentType("text/plain")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "utf-8")
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseContentTypeInvalid(t *testing.T) {
	reporter := newMockReporter(t)

	headers1 := map[string][]string{
		"Content-Type": {";"},
	}

	headers2 := map[string][]string{
		"Content-Type": {"charset=utf-8"},
	}

	resp1 := NewResponse(reporter, &http.Response{
		Header: http.Header(headers1),
	})

	resp2 := NewResponse(reporter, &http.Response{
		Header: http.Header(headers2),
	})

	resp1.ContentType("")
	resp1.chain.assertFailed(t)
	resp1.chain.reset()

	resp1.ContentType("", "")
	resp1.chain.assertFailed(t)
	resp1.chain.reset()

	resp2.ContentType("")
	resp2.chain.assertFailed(t)
	resp2.chain.reset()

	resp2.ContentType("", "")
	resp2.chain.assertFailed(t)
	resp2.chain.reset()
}

func TestResponseContentEncoding(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Encoding": {"gzip", "deflate"},
	}

	resp := NewResponse(reporter, &http.Response{
		Header: http.Header(headers),
	})

	resp.ContentEncoding("gzip", "deflate")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentEncoding("deflate", "gzip")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentEncoding("gzip")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentEncoding()
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseTransferEncoding(t *testing.T) {
	reporter := newMockReporter(t)

	resp := NewResponse(reporter, &http.Response{
		TransferEncoding: []string{"foo", "bar"},
	})

	resp.TransferEncoding("foo", "bar")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.TransferEncoding("foo")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.TransferEncoding()
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseText(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain", "utf-8")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("application/json")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Text()
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t, "hello, world!", resp.Text().Raw())
}

func TestResponseForm(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("application/x-www-form-urlencoded")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("application/x-www-form-urlencoded", "")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Form()
	resp.chain.assertOK(t)
	resp.chain.reset()

	expected := map[string]interface{}{
		"a": "1",
		"b": "2",
	}

	assert.Equal(t, expected, resp.Form().Raw())
}

func TestResponseFormBadBody(t *testing.T) {
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
	resp.chain.reset()

	assert.True(t, resp.Form().Raw() == nil)
}

func TestResponseFormBadType(t *testing.T) {
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
	resp.chain.reset()

	assert.True(t, resp.Form().Raw() == nil)
}

func TestResponseJSON(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("application/json")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("application/json", "utf-8")
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentType("text/plain")
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
}

func TestResponseJSONBadBody(t *testing.T) {
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
	resp.chain.reset()

	assert.True(t, resp.JSON().Raw() == nil)
}

func TestResponseJSONCharsetEmpty(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
}

func TestResponseJSONCharsetBad(t *testing.T) {
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
	resp.chain.reset()

	assert.Equal(t, nil, resp.JSON().Raw())
}

func TestResponseJSONP(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"application/javascript; charset=utf-8"},
	}

	body1 := `foo({"key": "value"})`
	body2 := `foo({"key": "value"});`
	body3 := ` foo ( {"key": "value"} ) ; `

	for _, body := range []string{body1, body2, body3} {
		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		assert.Equal(t, body, resp.Body().Raw())
		resp.chain.assertOK(t)
		resp.chain.reset()

		resp.ContentType("application/javascript")
		resp.chain.assertOK(t)
		resp.chain.reset()

		resp.ContentType("application/javascript", "utf-8")
		resp.chain.assertOK(t)
		resp.chain.reset()

		resp.ContentType("text/plain")
		resp.chain.assertFailed(t)
		resp.chain.reset()

		resp.JSONP("foo")
		resp.chain.assertOK(t)
		resp.chain.reset()

		assert.Equal(t,
			map[string]interface{}{"key": "value"}, resp.JSONP("foo").Object().Raw())

		resp.JSONP("fo")
		resp.chain.assertFailed(t)
		resp.chain.reset()

		resp.JSONP("")
		resp.chain.assertFailed(t)
		resp.chain.reset()
	}
}

func TestResponseJSONPBadBody(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"application/javascript; charset=utf-8"},
	}

	body1 := `foo`
	body2 := `foo();`
	body3 := `foo(`
	body4 := `foo({);`

	for _, body := range []string{body1, body2, body3, body4} {
		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header(headers),
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}

		resp := NewResponse(reporter, httpResp)

		resp.JSONP("foo")
		resp.chain.assertFailed(t)
		resp.chain.reset()

		assert.True(t, resp.JSONP("foo").Raw() == nil)
	}
}

func TestResponseJSONPCharsetEmpty(t *testing.T) {
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
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSONP("foo").Object().Raw())
}

func TestResponseJSONPCharsetBad(t *testing.T) {
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
	resp.chain.reset()

	assert.Nil(t, resp.JSONP("foo").Raw())
}

func TestResponseContentOpts(t *testing.T) {
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
			c.assertOK(t)
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
