package httpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestResponseFailed(t *testing.T) {
	checker := newMockChecker(t)

	checker.Fail("fail")

	resp := NewResponse(checker, nil)

	resp.Status(123)
	resp.Headers(nil)
	resp.Header("foo", "bar")
	resp.NoContent()
	resp.JSON()
}

func TestResponseHeaders(t *testing.T) {
	checker := newMockChecker(t)

	headers := map[string][]string{
		"First-Header":  []string{"foo"},
		"Second-Header": []string{"bar"},
	}

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       nil,
	}

	resp := NewResponse(checker, httpResp)
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t, httpResp, resp.Raw())

	resp.Status(http.StatusOK)
	checker.AssertSuccess(t)
	checker.Reset()

	resp.Status(http.StatusNotFound)
	checker.AssertFailed(t)
	checker.Reset()

	resp.Headers(headers)
	checker.AssertSuccess(t)
	checker.Reset()

	partialHeaders := make(map[string][]string)
	partialHeaders["Content-Type"] = headers["Content-Type"]

	resp.Headers(partialHeaders)
	checker.AssertFailed(t)
	checker.Reset()

	for k, v := range headers {
		resp.Header(k, v[0])
		checker.AssertSuccess(t)
		checker.Reset()
	}

	resp.Header("Bad-Header", "noValue")
	checker.AssertFailed(t)
	checker.Reset()
}

func TestResponseNoContentEmpty(t *testing.T) {
	checker := newMockChecker(t)

	headers := map[string][]string{
		"Content-Type": []string{""},
	}

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       closingBuffer{bytes.NewBufferString("")},
	}

	resp := NewResponse(checker, httpResp)

	resp.NoContent()
	checker.AssertSuccess(t)
	checker.Reset()

	resp.JSON()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestResponseNoContentNil(t *testing.T) {
	checker := newMockChecker(t)

	headers := map[string][]string{
		"Content-Type": []string{""},
	}

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       nil,
	}

	resp := NewResponse(checker, httpResp)

	resp.NoContent()
	checker.AssertSuccess(t)
	checker.Reset()

	resp.JSON()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestResponseJson(t *testing.T) {
	checker := newMockChecker(t)

	headers := map[string][]string{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}

	body := `{"key": "value"}`

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       closingBuffer{bytes.NewBufferString(body)},
	}

	resp := NewResponse(checker, httpResp)

	resp.NoContent()
	checker.AssertFailed(t)
	checker.Reset()

	resp.JSON()
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())

	assert.False(t, resp.checker == resp.JSON().checker)
}

func TestResponseJsonEncodingEmpty(t *testing.T) {
	checker := newMockChecker(t)

	headers := map[string][]string{
		"Content-Type": []string{"application/json"},
	}

	body := `{"key": "value"}`

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       closingBuffer{bytes.NewBufferString(body)},
	}

	resp := NewResponse(checker, httpResp)

	resp.NoContent()
	checker.AssertFailed(t)
	checker.Reset()

	resp.JSON()
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t,
		map[string]interface{}{"key": "value"}, resp.JSON().Object().Raw())
}

func TestResponseJsonEncodingBad(t *testing.T) {
	checker := newMockChecker(t)

	headers := map[string][]string{
		"Content-Type": []string{"application/json; charset=bad"},
	}

	body := `{"key": "value"}`

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header(headers),
		Body:       closingBuffer{bytes.NewBufferString(body)},
	}

	resp := NewResponse(checker, httpResp)

	resp.NoContent()
	checker.AssertFailed(t)
	checker.Reset()

	resp.JSON()
	checker.AssertFailed(t)
	checker.Reset()

	assert.Equal(t, nil, resp.JSON().Raw())
}
