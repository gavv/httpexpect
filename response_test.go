package httpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestResponseFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	resp := &Response{chain, nil, nil}

	resp.chain.assertFailed(t)

	assert.False(t, resp.Body() == nil)
	assert.False(t, resp.JSON() == nil)

	resp.Body().chain.assertFailed(t)
	resp.JSON().chain.assertFailed(t)

	resp.Status(123)
	resp.Headers(nil)
	resp.Header("foo", "bar")
	resp.NoContent()
	resp.ContentTypeJSON()
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

	resp.Headers(headers)
	resp.chain.assertOK(t)
	resp.chain.reset()

	partialHeaders := make(map[string][]string)
	partialHeaders["Content-Type"] = headers["Content-Type"]

	resp.Headers(partialHeaders)
	resp.chain.assertFailed(t)
	resp.chain.reset()

	for k, v := range headers {
		resp.Header(k, v[0])
		resp.chain.assertOK(t)
		resp.chain.reset()
	}

	resp.Header("Bad-Header", "noValue")
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseBody(t *testing.T) {
	reporter := newMockReporter(t)

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       readCloserAdapter{bytes.NewBufferString("body")},
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
		Body:       readCloserAdapter{bytes.NewBufferString("")},
	}

	resp := NewResponse(reporter, httpResp)

	assert.Equal(t, "", resp.Body().Raw())
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.NoContent()
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.ContentTypeJSON()
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

	resp.ContentTypeJSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()
}

func TestResponseJson(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"application/json; charset=utf-8"},
	}

	body := `{"key": "value"}`

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       readCloserAdapter{bytes.NewBufferString(body)},
	}

	resp := NewResponse(reporter, httpResp)

	assert.Equal(t, body, resp.Body().Raw())
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.NoContent()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentTypeJSON()
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
}

func TestResponseJsonEncodingEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}

	body := `{"key": "value"}`

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       readCloserAdapter{bytes.NewBufferString(body)},
	}

	resp := NewResponse(reporter, httpResp)

	resp.NoContent()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentTypeJSON()
	resp.chain.assertOK(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertOK(t)
	resp.chain.reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
}

func TestResponseJsonEncodingBad(t *testing.T) {
	reporter := newMockReporter(t)

	headers := map[string][]string{
		"Content-Type": {"application/json; charset=bad"},
	}

	body := `{"key": "value"}`

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       readCloserAdapter{bytes.NewBufferString(body)},
	}

	resp := NewResponse(reporter, httpResp)

	resp.NoContent()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.ContentTypeJSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	resp.JSON()
	resp.chain.assertFailed(t)
	resp.chain.reset()

	assert.Equal(t, nil, resp.JSON().Raw())
}
