package httpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestResponseFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	resp := &Response{chain, nil, nil}

	resp.chain.assertFailed(t)

	assert.False(t, resp.Headers() == nil)
	assert.False(t, resp.Header("foo") == nil)
	assert.False(t, resp.Body() == nil)
	assert.False(t, resp.JSON() == nil)

	resp.Headers().chain.assertFailed(t)
	resp.Header("foo").chain.assertFailed(t)
	resp.Body().chain.assertFailed(t)
	resp.Text().chain.assertFailed(t)
	resp.JSON().chain.assertFailed(t)

	resp.Status(123)
	resp.NoContent()
	resp.ContentType("", "")
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

	assert.Equal(t, httpResp, resp.Raw())

	resp.Status(http.StatusOK)
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.Status(http.StatusNotFound)
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Headers().Equal(headers).chain.assertOK(t)

	for k, v := range headers {
		resp.Header(k).Equal(v[0]).chain.assertOK(t)
	}

	resp.Header("Bad-Header").Empty().chain.assertOK(t)
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

	resp.JSON()
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

	resp.JSON()
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

	resp.NoContent()
	resp.chain.assertFailed(t)
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

	resp.Form()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertFailed(t)
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

	resp.NoContent()
	resp.chain.assertFailed(t)
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

	resp.Text()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
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

	resp.NoContent()
	resp.chain.assertFailed(t)
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

	resp.Text()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.Form()
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
		"Content-Type": {"application/json"},
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

	resp.NoContent()
	resp.chain.assertFailed(t)
	resp.chain.reset()

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

	resp.NoContent()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	assert.Equal(t, nil, resp.JSON().Raw())
}
