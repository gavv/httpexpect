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

// ContentType succeedes if response contains Content-Type header with given
// media type and charset.
//
// If charset is omitted, and mediaType is non-empty, Content-Type header
// should contain empty or utf-8 charset.
//
// If charset is omitted, and mediaType is also empty, Content-Type header
// should contain no charset.
func (r *Response) ContentType(mediaType string, charset ...string) *Response {
	r.checkContentType(mediaType, charset...)
	return r
}

// Text returns a new String object that may be used to inspect response body.
//
// Text succeedes if response contains "text/plain" Content-Type header
// with empty or "utf-8" charset.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Text().Equal("hello, world!")
func (r *Response) Text() *String {
	var content string

	if !r.chain.failed() && r.checkContentType("text/plain") {
		content = string(r.content)
	}

	return &String{r.chain, content}
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

	if !r.checkContentType("application/json") {
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(r.content, &value); err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	return value
}

func (r *Response) checkContentType(expectedType string, expectedCharset ...string) bool {
	if r.chain.failed() {
		return false
	}

	contentType := r.resp.Header.Get("Content-Type")

	mediaType, params, _ := mime.ParseMediaType(contentType)
	charset := params["charset"]

	if mediaType != expectedType {
		r.chain.fail(
			"\nexpected \"Content-Type\" header with \"" + expectedType +
				"\" media type,\nbut got \"" + mediaType + "\"")
		return false
	}

	if len(expectedCharset) == 0 {
		if expectedType == "" {
			if charset != "" {
				r.chain.fail(
					"\nexpected \"Content-Type\" header with empty charset," +
						"\nbut got \"" + charset + "\"")
				return false
			}
		} else {
			if charset != "" && !strings.EqualFold(charset, "utf-8") {
				r.chain.fail(
					"\nexpected \"Content-Type\" header with \"utf-8\" or empty charset," +
						"\nbut got \"" + charset + "\"")
				return false
			}
		}
	} else {
		if !strings.EqualFold(charset, expectedCharset[0]) {
			r.chain.fail(
				"\nexpected \"Content-Type\" header with \"" + expectedCharset[0] +
					"\" charset,\nbut got \"" + charset + "\"")
			return false
		}
	}

	return true
}

func (r *Response) checkEqual(what string, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		r.chain.fail("\nexpected %s equal to:\n%s\n\nbut got:\n%s", what,
			dumpValue(expected), dumpValue(actual))
	}
}
