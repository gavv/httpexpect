package httpexpect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ajg/form"
	"github.com/gorilla/websocket"
)

// Response provides methods to inspect attached http.Response object.
type Response struct {
	config Config
	chain  *chain

	httpResp  *http.Response
	websocket *websocket.Conn
	rtt       *time.Duration

	content []byte
	cookies []*http.Cookie
}

// NewResponse returns a new Response instance.
//
// Both reporter and response should not be nil. If response is nil,
// failure is reported.
//
// If rtt is given, it defines response round-trip time to be reported
// by response.RoundTripTime().
func NewResponse(
	reporter Reporter, response *http.Response, rtt ...time.Duration,
) *Response {
	config := Config{
		Reporter: reporter,
	}
	config.fillDefaults()

	return newResponse(responseOpts{
		config:   config,
		chain:    newChainWithConfig("Response()", config),
		httpResp: response,
		rtt:      rtt,
	})
}

type responseOpts struct {
	config    Config
	chain     *chain
	httpResp  *http.Response
	websocket *websocket.Conn
	rtt       []time.Duration
}

func newResponse(opts responseOpts) *Response {
	r := &Response{
		config: opts.config,
		chain:  opts.chain.clone(),
	}

	if opts.httpResp == nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{opts.httpResp},
			Errors: []error{
				errors.New("expected: non-nil response"),
			},
		})
		return r
	}

	if len(opts.rtt) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple rtt arguments"),
			},
		})
		return r
	}

	r.httpResp = opts.httpResp
	r.websocket = opts.websocket

	r.content = getContent(r.chain, r.httpResp)
	r.cookies = r.httpResp.Cookies()

	if len(opts.rtt) > 0 {
		rtt := opts.rtt[0]
		r.rtt = &rtt
	}

	r.chain.setResponse(r)

	return r
}

func getContent(chain *chain, resp *http.Response) []byte {
	if resp.Body == nil {
		return []byte{}
	}

	if bw, ok := resp.Body.(*bodyWrapper); ok {
		bw.Rewind()
	}

	content, err := ioutil.ReadAll(resp.Body)

	closeErr := resp.Body.Close()
	if err == nil {
		err = closeErr
	}

	if err != nil {
		chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to read response body"),
				err,
			},
		})
		return nil
	}

	return content
}

// Raw returns underlying http.Response object.
// This is the value originally passed to NewResponse.
func (r *Response) Raw() *http.Response {
	return r.httpResp
}

// RoundTripTime returns a new Duration instance with response round-trip time.
//
// The returned duration is the time interval starting just before request is
// sent and ending right after response is received (handshake finished for
// WebSocket request), retrieved from a monotonic clock source.
//
// Example:
//
//	resp := NewResponse(t, response, time.Duration(10000000))
//	resp.RoundTripTime().Lt(10 * time.Millisecond)
func (r *Response) RoundTripTime() *Duration {
	r.chain.enter("RoundTripTime()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newDuration(r.chain, nil)
	}

	return newDuration(r.chain, r.rtt)
}

// Deprecated: use RoundTripTime instead.
func (r *Response) Duration() *Number {
	r.chain.enter("Duration()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newNumber(r.chain, 0)
	}

	if r.rtt == nil {
		return newNumber(r.chain, 0)
	}

	return newNumber(r.chain, float64(*r.rtt))
}

// Status succeeds if response contains given status code.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Status(http.StatusOK)
func (r *Response) Status(status int) *Response {
	r.chain.enter("Status()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.checkEqual("http status",
		statusCodeText(status), statusCodeText(r.httpResp.StatusCode))

	return r
}

// StatusRange is enum for response status ranges.
type StatusRange int

const (
	// Status1xx defines "Informational" status codes.
	Status1xx StatusRange = 100

	// Status2xx defines "Success" status codes.
	Status2xx StatusRange = 200

	// Status3xx defines "Redirection" status codes.
	Status3xx StatusRange = 300

	// Status4xx defines "Client Error" status codes.
	Status4xx StatusRange = 400

	// Status5xx defines "Server Error" status codes.
	Status5xx StatusRange = 500
)

// StatusRange succeeds if response status belongs to given range.
//
// Supported ranges:
//   - Status1xx - Informational
//   - Status2xx - Success
//   - Status3xx - Redirection
//   - Status4xx - Client Error
//   - Status5xx - Server Error
//
// See https://en.wikipedia.org/wiki/List_of_HTTP_status_codes.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.StatusRange(Status2xx)
func (r *Response) StatusRange(rn StatusRange) *Response {
	r.chain.enter("StatusRange()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	status := statusCodeText(r.httpResp.StatusCode)

	actual := statusRangeText(r.httpResp.StatusCode)
	expected := statusRangeText(int(rn))

	if actual == "" || actual != expected {
		r.chain.fail(AssertionFailure{
			Type:   AssertBelongs,
			Actual: &AssertionValue{status},
			Expected: &AssertionValue{AssertionList{
				statusRangeText(int(rn)),
			}},
			Errors: []error{
				errors.New("expected: http status belongs to given range"),
			},
		})
	}

	return r
}

func statusCodeText(code int) string {
	if s := http.StatusText(code); s != "" {
		return strconv.Itoa(code) + " " + s
	}
	return strconv.Itoa(code)
}

func statusRangeText(code int) string {
	switch {
	case code >= 100 && code < 200:
		return "1xx Informational"
	case code >= 200 && code < 300:
		return "2xx Success"
	case code >= 300 && code < 400:
		return "3xx Redirection"
	case code >= 400 && code < 500:
		return "4xx Client Error"
	case code >= 500 && code < 600:
		return "5xx Server Error"
	default:
		return ""
	}
}

// Headers returns a new Object instance with response header map.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Headers().Value("Content-Type").String().Equal("application-json")
func (r *Response) Headers() *Object {
	r.chain.enter("Headers()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newObject(r.chain, nil)
	}

	var value map[string]interface{}
	value, _ = canonMap(r.chain, r.httpResp.Header)

	return newObject(r.chain, value)
}

// Header returns a new String instance with given header field.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Header("Content-Type").Equal("application-json")
//	resp.Header("Date").AsDateTime().Le(time.Now())
func (r *Response) Header(header string) *String {
	r.chain.enter("Header(%q)", header)
	defer r.chain.leave()

	if r.chain.failed() {
		return newString(r.chain, "")
	}

	value := r.httpResp.Header.Get(header)

	return newString(r.chain, value)
}

// Cookies returns a new Array instance with all cookie names set by this response.
// Returned Array contains a String value for every cookie name.
//
// Note that this returns only cookies set by Set-Cookie headers of this response.
// It doesn't return session cookies from previous responses, which may be stored
// in a cookie jar.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Cookies().Contains("session")
func (r *Response) Cookies() *Array {
	r.chain.enter("Cookies()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newArray(r.chain, nil)
	}

	names := []interface{}{}
	for _, c := range r.cookies {
		names = append(names, c.Name)
	}

	return newArray(r.chain, names)
}

// Cookie returns a new Cookie instance with specified cookie from response.
//
// Note that this returns only cookies set by Set-Cookie headers of this response.
// It doesn't return session cookies from previous responses, which may be stored
// in a cookie jar.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Cookie("session").Domain().Equal("example.com")
func (r *Response) Cookie(name string) *Cookie {
	r.chain.enter("Cookie(%q)", name)
	defer r.chain.leave()

	if r.chain.failed() {
		return newCookie(r.chain, nil)
	}

	var cookie *Cookie

	names := []string{}
	for _, c := range r.cookies {
		if c.Name == name {
			cookie = newCookie(r.chain, c)
			break
		}
		names = append(names, c.Name)
	}

	if cookie == nil {
		r.chain.fail(AssertionFailure{
			Type:     AssertContainsElement,
			Actual:   &AssertionValue{names},
			Expected: &AssertionValue{name},
			Errors: []error{
				errors.New("expected: response contains cookie with given name"),
			},
		})
		return newCookie(r.chain, nil)
	}

	return cookie
}

// Websocket returns Websocket instance for interaction with WebSocket server.
//
// May be called only if the WithWebsocketUpgrade was called on the request.
// That is responsibility of the caller to explicitly disconnect websocket after use.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithWebsocketUpgrade()
//	ws := req.Expect().Websocket()
//	defer ws.Disconnect()
func (r *Response) Websocket() *Websocket {
	r.chain.enter("Websocket()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newWebsocket(r.chain, r.config, nil)
	}

	if r.websocket == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New(
					"Websocket() requires WithWebsocketUpgrade() to be called on request"),
			},
		})
		return newWebsocket(r.chain, r.config, nil)
	}

	return newWebsocket(r.chain, r.config, r.websocket)
}

// Body returns a new String instance with response body.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Body().NotEmpty()
//	resp.Body().Length().Equal(100)
func (r *Response) Body() *String {
	return newString(r.chain, string(r.content))
}

// NoContent succeeds if response contains empty Content-Type header and
// empty body.
func (r *Response) NoContent() *Response {
	r.chain.enter("NoContent()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	contentType := r.httpResp.Header.Get("Content-Type")

	r.checkEqual(`"Content-Type" header`, "", contentType)
	r.checkEqual("body", "", string(r.content))

	return r
}

// ContentType succeeds if response contains Content-Type header with given
// media type and charset.
//
// If charset is omitted, and mediaType is non-empty, Content-Type header
// should contain empty or utf-8 charset.
//
// If charset is omitted, and mediaType is also empty, Content-Type header
// should contain no charset.
func (r *Response) ContentType(mediaType string, charset ...string) *Response {
	r.chain.enter("ContentType()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if len(charset) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple charset arguments"),
			},
		})
		return r
	}

	r.checkContentType(mediaType, charset...)

	return r
}

// ContentEncoding succeeds if response has exactly given Content-Encoding list.
// Common values are empty, "gzip", "compress", "deflate", "identity" and "br".
func (r *Response) ContentEncoding(encoding ...string) *Response {
	r.chain.enter("ContentEncoding()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.checkEqual(`"Content-Encoding" header`,
		encoding,
		r.httpResp.Header["Content-Encoding"])

	return r
}

// TransferEncoding succeeds if response contains given Transfer-Encoding list.
// Common values are empty, "chunked" and "identity".
func (r *Response) TransferEncoding(encoding ...string) *Response {
	r.chain.enter("TransferEncoding()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.checkEqual(`"Transfer-Encoding" header`,
		encoding,
		r.httpResp.TransferEncoding)

	return r
}

// ContentOpts define parameters for matching the response content parameters.
type ContentOpts struct {
	// The media type Content-Type part, e.g. "application/json"
	MediaType string
	// The character set Content-Type part, e.g. "utf-8"
	Charset string
}

// Text returns a new String instance with response body.
//
// Text succeeds if response contains "text/plain" Content-Type header
// with empty or "utf-8" charset.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Text().Equal("hello, world!")
//	resp.Text(ContentOpts{
//	  MediaType: "text/plain",
//	}).Equal("hello, world!")
func (r *Response) Text(options ...ContentOpts) *String {
	r.chain.enter("Text()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newString(r.chain, "")
	}

	if len(options) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple options arguments"),
			},
		})
		return newString(r.chain, "")
	}

	if !r.checkContentOptions(options, "text/plain") {
		return newString(r.chain, "")
	}

	content := string(r.content)

	return newString(r.chain, content)
}

// Form returns a new Object instance with form decoded from response body.
//
// Form succeeds if response contains "application/x-www-form-urlencoded"
// Content-Type header and if form may be decoded from response body.
// Decoding is performed using https://github.com/ajg/form.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.Form().Value("foo").Equal("bar")
//	resp.Form(ContentOpts{
//	  MediaType: "application/x-www-form-urlencoded",
//	}).Value("foo").Equal("bar")
func (r *Response) Form(options ...ContentOpts) *Object {
	r.chain.enter("Form()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newObject(r.chain, nil)
	}

	if len(options) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple options arguments"),
			},
		})
		return newObject(r.chain, nil)
	}

	object := r.getForm(options...)

	return newObject(r.chain, object)
}

func (r *Response) getForm(options ...ContentOpts) map[string]interface{} {
	if !r.checkContentOptions(options, "application/x-www-form-urlencoded", "") {
		return nil
	}

	decoder := form.NewDecoder(bytes.NewReader(r.content))

	var object map[string]interface{}

	if err := decoder.Decode(&object); err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertValid,
			Actual: &AssertionValue{
				string(r.content),
			},
			Errors: []error{
				errors.New("failed to decode form"),
				err,
			},
		})
		return nil
	}

	return object
}

// JSON returns a new Value instance with JSON decoded from response body.
//
// JSON succeeds if response contains "application/json" Content-Type header
// with empty or "utf-8" charset and if JSON may be decoded from response body.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.JSON().Array().Elements("foo", "bar")
//	resp.JSON(ContentOpts{
//	  MediaType: "application/json",
//	}).Array.Elements("foo", "bar")
func (r *Response) JSON(options ...ContentOpts) *Value {
	r.chain.enter("JSON()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newValue(r.chain, nil)
	}

	if len(options) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple options arguments"),
			},
		})
		return newValue(r.chain, nil)
	}

	value := r.getJSON(options...)

	return newValue(r.chain, value)
}

func (r *Response) getJSON(options ...ContentOpts) interface{} {
	if !r.checkContentOptions(options, "application/json") {
		return nil
	}

	var value interface{}

	if err := json.Unmarshal(r.content, &value); err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertValid,
			Actual: &AssertionValue{
				string(r.content),
			},
			Errors: []error{
				errors.New("failed to decode json"),
				err,
			},
		})
		return nil
	}

	return value
}

// JSON returns a new Value instance with JSONP decoded from response body.
//
// JSONP succeeds if response contains "application/javascript" Content-Type
// header with empty or "utf-8" charset and response body of the following form:
//
//	callback(<valid json>);
//
// or:
//
//	callback(<valid json>)
//
// Whitespaces are allowed.
//
// Example:
//
//	resp := NewResponse(t, response)
//	resp.JSONP("myCallback").Array().Elements("foo", "bar")
//	resp.JSONP("myCallback", ContentOpts{
//	  MediaType: "application/javascript",
//	}).Array.Elements("foo", "bar")
func (r *Response) JSONP(callback string, options ...ContentOpts) *Value {
	r.chain.enter("JSONP()")
	defer r.chain.leave()

	if r.chain.failed() {
		return newValue(r.chain, nil)
	}

	if len(options) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple options arguments"),
			},
		})
		return newValue(r.chain, nil)
	}

	value := r.getJSONP(callback, options...)

	return newValue(r.chain, value)
}

var (
	jsonp = regexp.MustCompile(`^\s*([^\s(]+)\s*\((.*)\)\s*;*\s*$`)
)

func (r *Response) getJSONP(callback string, options ...ContentOpts) interface{} {
	if !r.checkContentOptions(options, "application/javascript") {
		return nil
	}

	m := jsonp.FindSubmatch(r.content)

	if len(m) != 3 || string(m[1]) != callback {
		r.chain.fail(AssertionFailure{
			Type: AssertValid,
			Actual: &AssertionValue{
				string(r.content),
			},
			Errors: []error{
				fmt.Errorf(`expected: JSONP body in form of "%s(<valid json>)"`,
					callback),
			},
		})
		return nil
	}

	var value interface{}

	if err := json.Unmarshal(m[2], &value); err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertValid,
			Actual: &AssertionValue{
				string(r.content),
			},
			Errors: []error{
				errors.New("failed to decode json"),
				err,
			},
		})
		return nil
	}

	return value
}

func (r *Response) checkContentOptions(
	options []ContentOpts, expectedType string, expectedCharset ...string,
) bool {
	if len(options) != 0 {
		if options[0].MediaType != "" {
			expectedType = options[0].MediaType
		}
		if options[0].Charset != "" {
			expectedCharset = []string{options[0].Charset}
		}
	}
	return r.checkContentType(expectedType, expectedCharset...)
}

func (r *Response) checkContentType(expectedType string, expectedCharset ...string) bool {
	contentType := r.httpResp.Header.Get("Content-Type")

	if expectedType == "" && len(expectedCharset) == 0 {
		if contentType == "" {
			return true
		}
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{contentType},
			Errors: []error{
				errors.New(`invalid "Content-Type" response header`),
				err,
			},
		})
		return false
	}

	if mediaType != expectedType {
		r.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{mediaType},
			Expected: &AssertionValue{expectedType},
			Errors: []error{
				errors.New(`unexpected media type in "Content-Type" response header`),
			},
		})
		return false
	}

	charset := params["charset"]

	if len(expectedCharset) == 0 {
		if charset != "" && !strings.EqualFold(charset, "utf-8") {
			r.chain.fail(AssertionFailure{
				Type:     AssertBelongs,
				Actual:   &AssertionValue{charset},
				Expected: &AssertionValue{AssertionList{"", "utf-8"}},
				Errors: []error{
					errors.New(`unexpected charset in "Content-Type" response header`),
				},
			})
			return false
		}
	} else {
		if !strings.EqualFold(charset, expectedCharset[0]) {
			r.chain.fail(AssertionFailure{
				Type:     AssertEqual,
				Actual:   &AssertionValue{charset},
				Expected: &AssertionValue{expectedCharset[0]},
				Errors: []error{
					errors.New(`unexpected charset in "Content-Type" response header`),
				},
			})
			return false
		}
	}

	return true
}

func (r *Response) checkEqual(what string, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		r.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{actual},
			Expected: &AssertionValue{expected},
			Errors: []error{
				fmt.Errorf("unexpected %s value", what),
			},
		})
	}
}
