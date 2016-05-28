package httpexpect

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Response provides methods to inspect attached http.Response object.
type Response struct {
	chain   chain
	resp    *http.Response
	content []byte
}

// NewResponse returns a new Response given a reporter used to report failures
// and http.Response to be inspected.
//
// Both reporter and response should not be nil. If response is nil, failure
// is reported.
func NewResponse(reporter Reporter, response *http.Response) *Response {
	return makeResponse(makeChain(reporter), response)
}

func makeResponse(chain chain, response *http.Response) *Response {
	if response == nil {
		chain.fail("expected non-nil response")
	}
	content := getContent(&chain, response)
	return &Response{
		chain:   chain,
		resp:    response,
		content: content,
	}
}

func getContent(chain *chain, resp *http.Response) []byte {
	if chain.failed() {
		return nil
	}

	if resp.Body == nil {
		return []byte{}
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		chain.fail(err.Error())
		return nil
	}

	return content
}

// Raw returns underlying http.Response object.
// This is the value originally passed to NewResponse.
func (r *Response) Raw() *http.Response {
	return r.resp
}

// Status succeedes if response contains given status code.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Status(http.StatusOK)
func (r *Response) Status(status int) *Response {
	if r.chain.failed() {
		return r
	}
	r.checkEqual("status", statusText(status), statusText(r.resp.StatusCode))
	return r
}

func statusText(code int) string {
	if s := http.StatusText(code); s != "" {
		return strconv.Itoa(code) + " " + s
	}
	return strconv.Itoa(code)
}

// Header returns a new Object that may be used to inspect header map.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Headers().Value("Content-Type").Contains("application-json")
func (r *Response) Headers() *Object {
	var value map[string]interface{}
	if !r.chain.failed() {
		value, _ = canonMap(&r.chain, r.resp.Header)
	}
	return &Object{r.chain, value}
}

// Header returns a new String object that may be used to inspect given header.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Header("Content-Type").Contains("application-json")
func (r *Response) Header(header string) *String {
	value := ""
	if !r.chain.failed() {
		value = r.resp.Header.Get(header)
	}
	return &String{r.chain, value}
}

// Body returns a new String object that may be used to inspect response body.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Body().NotEmpty()
func (r *Response) Body() *String {
	return &String{r.chain, string(r.content)}
}

// NoContent succeedes if response contains empty Content-Type header and
// empty body.
func (r *Response) NoContent() *Response {
	if r.chain.failed() {
		return r
	}

	contentType := r.resp.Header.Get("Content-Type")

	r.checkEqual("\"Content-Type\" header", "", contentType)
	r.checkEqual("body", "", string(r.content))

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
//  resp := NewResponse(t, response)
//  resp.JSON().Array().Elements("foo", "bar")
func (r *Response) JSON() *Value {
	value := r.getJSON()
	return &Value{r.chain, value}
}

func (r *Response) getJSON() interface{} {
	if r.chain.failed() {
		return nil
	}

	if !r.checkJSON() {
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(r.content, &value); err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	return value
}

func (r *Response) checkJSON() bool {
	if r.chain.failed() {
		return false
	}

	contentType := r.resp.Header.Get("Content-Type")

	mediaType, params, _ := mime.ParseMediaType(contentType)
	charset := params["charset"]

	if mediaType != "application/json" {
		r.chain.fail(
			"\nexpected \"Content-Type\" header with \"application/json\" media type,\n" +
				"but got \"" + mediaType + "\"")
		return false
	}

	if charset != "" && strings.ToLower(charset) != "utf-8" {
		r.chain.fail(
			"\nexpected \"Content-Type\" header with \"utf-8\" or empty charset,\n" +
				"but got \"" + charset + "\"")
		return false
	}

	return true
}

func (r *Response) checkEqual(what string, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		r.chain.fail("\nexpected %s equal to:\n%s\n\nbut got:\n%s", what,
			dumpValue(expected), dumpValue(actual))
	}
}
