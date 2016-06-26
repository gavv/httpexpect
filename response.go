package httpexpect

import (
	"bytes"
	"encoding/json"
	"github.com/ajg/form"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Response provides methods to inspect attached http.Response object.
type Response struct {
	chain   chain
	resp    *http.Response
	content []byte
	time    time.Duration
}

// NewResponse returns a new Response given a reporter used to report failures
// and http.Response to be inspected.
//
// Both reporter and response should not be nil. If response is nil, failure
// is reported.
//
// If duration, it defines response time to be reported by response.Time().
func NewResponse(
	reporter Reporter, response *http.Response, duration ...time.Duration) *Response {
	var dr time.Duration
	if len(duration) > 0 {
		dr = duration[0]
	}
	return makeResponse(makeChain(reporter), response, dr)
}

func makeResponse(chain chain, response *http.Response, duration time.Duration) *Response {
	if response == nil {
		chain.fail("expected non-nil response")
	}
	content := getContent(&chain, response)
	return &Response{
		chain:   chain,
		resp:    response,
		content: content,
		time:    duration,
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

// Time returns a new Number object that may be used to inspect response time,
// in nanoseconds.
//
// Example:
//  resp := NewResponse(t, response, time.Duration(10000000))
//  resp.Time().Equal(10 * time.Millisecond)
func (r *Response) Time() *Number {
	return &Number{r.chain, float64(r.time)}
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

// Headers returns a new Object that may be used to inspect header map.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Headers().Value("Content-Type").String().Equal("application-json")
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
//  resp.Header("Content-Type").Equal("application-json")
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
//  resp.Body().Length().Equal(100)
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

// Form returns a new Object that may be used to inspect form contents
// of response.
//
// Form succeedes if response contains "application/x-www-form-urlencoded"
// Content-Type header and if form may be decoded from response body.
// Decoding is performed using https://github.com/ajg/form.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Form().Value("foo").Equal("bar")
func (r *Response) Form() *Object {
	object := r.getForm()
	return &Object{r.chain, object}
}

func (r *Response) getForm() map[string]interface{} {
	if r.chain.failed() {
		return nil
	}

	if !r.checkContentType("application/x-www-form-urlencoded", "") {
		return nil
	}

	decoder := form.NewDecoder(bytes.NewReader(r.content))

	var object map[string]interface{}
	if err := decoder.Decode(&object); err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	return object
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

	if expectedType == "" && len(expectedCharset) == 0 {
		if contentType == "" {
			return true
		}
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		r.chain.fail("\ngot invalid \"Content-Type\" header %s",
			strconv.Quote(contentType))
		return false
	}

	if mediaType != expectedType {
		r.chain.fail(
			"\nexpected \"Content-Type\" header with %s media type,"+
				"\nbut got %s",
			strconv.Quote(expectedType), strconv.Quote(mediaType))
		return false
	}

	charset := params["charset"]

	if len(expectedCharset) == 0 {
		if charset != "" && !strings.EqualFold(charset, "utf-8") {
			r.chain.fail(
				"\nexpected \"Content-Type\" header with \"utf-8\" or empty charset,"+
					"\nbut got %s",
				strconv.Quote(charset))
			return false
		}
	} else {
		if !strings.EqualFold(charset, expectedCharset[0]) {
			r.chain.fail(
				"\nexpected \"Content-Type\" header with %s charset,"+
					"\nbut got %s",
				strconv.Quote(expectedCharset[0]), strconv.Quote(charset))
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
