package httpexpect

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

// Response provides methods to inspect attached http.Response object.
type Response struct {
	checker Checker
	resp    *http.Response
	content []byte
}

// NewResponse returns a new Response given a checker used to report failures
// and http.Response to be inspected.
//
// Both checker and response should not be nil. If response is nil, failure is reported.
func NewResponse(checker Checker, response *http.Response) *Response {
	if response == nil {
		checker.Fail("expected non-nil response")
	}
	return &Response{checker, response, nil}
}

// Raw returns underlying http.Response object.
// This is the value originally passed to NewResponse.
func (r *Response) Raw() *http.Response {
	return r.resp
}

// Status succeedes if response contains given status code.
//
// Example:
//  resp := NewResponse(checker, response)
//  resp.Status(http.StatusOK)
func (r *Response) Status(status int) *Response {
	if r.checker.Failed() {
		return r
	}
	r.checker.Equal(status, r.resp.StatusCode)
	return r
}

// Headers succeedes if response has exactly given headers map.
//
// Example:
//  resp := NewResponse(checker, response)
//  resp.Headers(map[string][]string{
//      "Content-Type": []string{"application-json"},
//  })
func (r *Response) Headers(headers map[string][]string) *Response {
	if r.checker.Failed() {
		return r
	}
	r.checker.Equal(headers, map[string][]string(r.resp.Header))
	return r
}

// Header succeedes if response contains given single header.
//
// Example:
//  resp := NewResponse(checker, response)
//  resp.Header("Content-Type", "application-json")
func (r *Response) Header(k, v string) *Response {
	if r.checker.Failed() {
		return r
	}
	r.checker.Equal(v, r.resp.Header.Get(k))
	return r
}

// Body returns a new String object that may be used to inspect response body.
//
// Example:
//  resp := NewResponse(checker, response)
//  resp.Body().NotEmpty()
func (r *Response) Body() *String {
	value := r.getContent()
	return NewString(r.checker.Clone(), string(value))
}

// NoContent succeedes if response contains empty Content-Type header and
// empty body.
func (r *Response) NoContent() *Response {
	if r.checker.Failed() {
		return r
	}

	contentType := r.resp.Header.Get("Content-Type")

	content := string(r.getContent())

	r.checker.Equal("", contentType)
	r.checker.Equal("", content)

	return r
}

// ContentTypeJSON succeedes if response contains "application/json" Content-Type
// header with empty or "utf-8" charset
func (r *Response) ContentTypeJSON() *Response {
	r.checkJSON()
	return r
}

// JSON returns a new Value object that may be used to inspect JSON contents
// of response.
//
// JSON succeedes if response contains "application/json" Content-Type header
// with empty or "utf-8" charset and if JSON may be decoded from response body.
//
// Example:
//  resp := NewResponse(checker, response)
//  resp.JSON().Array().Elements("foo", "bar")
func (r *Response) JSON() *Value {
	value := r.getJSON()
	return NewValue(r.checker.Clone(), value)
}

func (r *Response) getContent() []byte {
	if r.content != nil {
		return r.content
	}

	if r.resp.Body == nil {
		return []byte{}
	}

	content, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		r.checker.Fail(err.Error())
		return nil
	}

	r.content = content
	return r.content
}

func (r *Response) getJSON() interface{} {
	if r.checker.Failed() {
		return nil
	}

	if !r.checkJSON() {
		return nil
	}

	content := r.getContent()

	var value interface{}
	if err := json.Unmarshal(content, &value); err != nil {
		r.checker.Fail(err.Error())
		return nil
	}

	return value
}

func (r *Response) checkJSON() bool {
	if r.checker.Failed() {
		return false
	}

	contentType := r.resp.Header.Get("Content-Type")

	mediaType, params, _ := mime.ParseMediaType(contentType)
	charset := params["charset"]

	r.checker.Equal("application/json", mediaType)
	if r.checker.Failed() {
		return false
	}

	if charset != "" && strings.ToLower(charset) != "utf-8" {
		r.checker.Fail("bad charset: expected empty or 'utf-8', got '" + charset + "'")
		return false
	}

	return true
}
