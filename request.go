package httpexpect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/ajg/form"
	"github.com/fatih/structs"
	"github.com/google/go-querystring/query"
	"github.com/gorilla/websocket"
	"github.com/imkira/go-interpol"
)

// Request provides methods to incrementally build http.Request object,
// send it, and receive response.
type Request struct {
	config Config
	chain  *chain

	redirectPolicy RedirectPolicy
	maxRedirects   int

	retryPolicy   RetryPolicy
	maxRetries    int
	minRetryDelay time.Duration
	maxRetryDelay time.Duration
	sleepFn       func(d time.Duration)

	timeout time.Duration

	httpReq *http.Request
	path    string
	query   url.Values

	form      url.Values
	formbuf   *bytes.Buffer
	multipart *multipart.Writer

	bodySetter string
	typeSetter string
	forceType  bool

	wsUpgrade bool

	transforms []func(*http.Request)
	matchers   []func(*Response)
}

// NewRequest returns a new Request instance.
//
// method defines the HTTP method (GET, POST, PUT, etc.). path defines url path.
//
// Simple interpolation is allowed for {named} parameters in path:
//   - if pathargs is given, it's used to substitute first len(pathargs) parameters,
//     regardless of their names
//   - if WithPath() or WithPathObject() is called, it's used to substitute given
//     parameters by name
//
// For example:
//
//	req := NewRequest(config, "POST", "/repos/{user}/{repo}", "gavv", "httpexpect")
//	// path will be "/repos/gavv/httpexpect"
//
// Or:
//
//	req := NewRequest(config, "POST", "/repos/{user}/{repo}")
//	req.WithPath("user", "gavv")
//	req.WithPath("repo", "httpexpect")
//	// path will be "/repos/gavv/httpexpect"
//
// After interpolation, path is urlencoded and appended to Config.BaseURL,
// separated by slash. If BaseURL ends with a slash and path (after interpolation)
// starts with a slash, only single slash is inserted.
func NewRequest(config Config, method, path string, pathargs ...interface{}) *Request {
	config.fillDefaults()

	return newRequest(
		newChainWithConfig("Request()", config),
		config,
		method,
		path,
		pathargs...,
	)
}

func newRequest(
	parent *chain, config Config, method, path string, pathargs ...interface{},
) *Request {
	if config.RequestFactory == nil {
		panic("Config.RequestFactory is nil")
	}

	if config.Client == nil {
		panic("Config.Client is nil")
	}

	if config.AssertionHandler == nil {
		panic("Config.AssertionHandler is nil")
	}

	r := &Request{
		config: config,
		chain:  parent.clone(),

		redirectPolicy: defaultRedirectPolicy,
		maxRedirects:   -1,

		retryPolicy:   RetryTemporaryNetworkAndServerErrors,
		maxRetries:    0,
		minRetryDelay: time.Millisecond * 50,
		maxRetryDelay: time.Second * 5,
		sleepFn:       time.Sleep,
	}

	r.initPath(path, pathargs...)
	r.initReq(method)

	r.chain.setRequest(r)

	return r
}

func (r *Request) initPath(path string, pathargs ...interface{}) {
	var n int

	path, err := interpol.WithFunc(path, func(k string, w io.Writer) error {
		if n < len(pathargs) {
			if pathargs[n] == nil {
				r.chain.fail(AssertionFailure{
					Type:   AssertValid,
					Actual: &AssertionValue{pathargs},
					Errors: []error{
						fmt.Errorf("unexpected nil argument at index %d", n),
					},
				})
			} else {
				mustWrite(w, fmt.Sprint(pathargs[n]))
			}
		} else {
			mustWrite(w, "{")
			mustWrite(w, k)
			mustWrite(w, "}")
		}
		n++
		return nil
	})

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{path},
			Errors: []error{
				errors.New("invalid interpol string"),
				err,
			},
		})
	}

	r.path = path
}

func (r *Request) initReq(method string) {
	httpReq, err := r.config.RequestFactory.NewRequest(method, r.config.BaseURL, nil)

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to create http request"),
				err,
			},
		})
	}

	r.httpReq = httpReq
}

// WithName sets convenient request name.
// This name will be included in assertion reports for this request.
//
// Example:
//
//	req := NewRequest(config, "POST", "/api/login")
//	req.WithName("Login Request")
func (r *Request) WithName(name string) *Request {
	r.chain.enter("WithName()")
	defer r.chain.leave()

	r.chain.setRequestName(name)

	return r
}

// WithMatcher attaches a matcher to the request.
// All attached matchers are invoked in the Expect method for a newly
// created Response.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithMatcher(func (resp *httpexpect.Response) {
//	    resp.Header("API-Version").NotEmpty()
//	})
func (r *Request) WithMatcher(matcher func(*Response)) *Request {
	r.chain.enter("WithMatcher()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if matcher == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	r.matchers = append(r.matchers, matcher)
	return r
}

// WithTransformer attaches a transform to the Request.
// All attachhed transforms are invoked in the Expect methods for
// http.Request struct, after it's encoded and before it's sent.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithTransformer(func(r *http.Request) { r.Header.Add("foo", "bar") })
func (r *Request) WithTransformer(transform func(*http.Request)) *Request {
	r.chain.enter("WithTransformer()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if transform == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	r.transforms = append(r.transforms, transform)

	return r
}

// WithClient sets client.
//
// The new client overwrites Config.Client. It will be used once to send the
// request and receive a response.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithClient(&http.Client{
//	  Transport: &http.Transport{
//	    DisableCompression: true,
//	  },
//	})
func (r *Request) WithClient(client Client) *Request {
	r.chain.enter("WithClient()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if client == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	r.config.Client = client

	return r
}

// WithHandler configures client to invoke the given handler directly.
//
// If Config.Client is http.Client, then only its Transport field is overwritten
// because the client may contain some state shared among requests like a cookie
// jar. Otherwise, the whole client is overwritten with a new client.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithHandler(myServer.someHandler)
func (r *Request) WithHandler(handler http.Handler) *Request {
	r.chain.enter("WithHandler()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if handler == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	if client, ok := r.config.Client.(*http.Client); ok {
		clientCopy := *client
		clientCopy.Transport = NewBinder(handler)
		r.config.Client = &clientCopy
	} else {
		r.config.Client = &http.Client{
			Transport: NewBinder(handler),
			Jar:       NewJar(),
		}
	}

	return r
}

// WithContext sets the context.
//
// Config.Context will be overwritten.
//
// Any retries will stop after one is cancelled.
// If the intended behavior is to continue any further retries, use WithTimeout.
//
// Example:
//
//	ctx, _ = context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
//	req := NewRequest(config, "GET", "/path")
//	req.WithContext(ctx)
//	req.Expect().Status(http.StatusOK)
func (r *Request) WithContext(ctx context.Context) *Request {
	r.chain.enter("WithContext()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if ctx == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	r.config.Context = ctx

	return r
}

// WithTimeout sets a timeout duration for the request.
//
// Will attach to the request a context.WithTimeout around the Config.Context
// or any context set WithContext. If these are nil, the new context will be
// created on top of a context.Background().
//
// Any retries will continue after one is cancelled.
// If the intended behavior is to stop any further retries, use WithContext or
// Config.Context.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithTimeout(time.Duration(3)*time.Second)
//	req.Expect().Status(http.StatusOK)
func (r *Request) WithTimeout(timeout time.Duration) *Request {
	r.chain.enter("WithTimeout()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.timeout = timeout

	return r
}

// RedirectPolicy defines how redirection responses are handled.
//
// Status codes 307, 308 require resending body. They are followed only if
// redirect policy is FollowAllRedirects.
//
// Status codes 301, 302, 303 don't require resending body. On such redirect,
// http.Client automatically switches HTTP method to GET, if it's not GET or
// HEAD already. These redirects are followed if redirect policy is either
// FollowAllRedirects or FollowRedirectsWithoutBody.
//
// Default redirect policy is FollowRedirectsWithoutBody.
type RedirectPolicy int

const (
	// indicates that WithRedirectPolicy was not called
	defaultRedirectPolicy RedirectPolicy = iota

	// DontFollowRedirects forbids following any redirects.
	// Redirection response is returned to the user and can be inspected.
	DontFollowRedirects

	// FollowAllRedirects allows following any redirects, including those
	// which require resending body.
	FollowAllRedirects

	// FollowRedirectsWithoutBody allows following only redirects which
	// don't require resending body.
	// If redirect requires resending body, it's not followed, and redirection
	// response is returned instead.
	FollowRedirectsWithoutBody
)

// WithRedirectPolicy sets policy for redirection response handling.
//
// How redirect is handled depends on both response status code and
// redirect policy. See comments for RedirectPolicy for details.
//
// Default redirect policy is defined by Client implementation.
// Default behavior of http.Client corresponds to FollowRedirectsWithoutBody.
//
// This method can be used only if Client interface points to
// *http.Client struct, since we rely on it in redirect handling.
//
// Example:
//
//	req1 := NewRequest(config, "POST", "/path")
//	req1.WithRedirectPolicy(FollowAllRedirects)
//	req1.Expect().Status(http.StatusOK)
//
//	req2 := NewRequest(config, "POST", "/path")
//	req2.WithRedirectPolicy(DontFollowRedirects)
//	req2.Expect().Status(http.StatusPermanentRedirect)
func (r *Request) WithRedirectPolicy(policy RedirectPolicy) *Request {
	r.chain.enter("WithRedirectPolicy()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.redirectPolicy = policy

	return r
}

// WithMaxRedirects sets maximum number of redirects to follow.
//
// If the number of redirects exceedes this limit, request is failed.
//
// Default limit is defined by Client implementation.
// Default behavior of http.Client corresponds to maximum of 10-1 redirects.
//
// This method can be used only if Client interface points to
// *http.Client struct, since we rely on it in redirect handling.
//
// Example:
//
//	req1 := NewRequest(config, "POST", "/path")
//	req1.WithMaxRedirects(1)
//	req1.Expect().Status(http.StatusOK)
func (r *Request) WithMaxRedirects(maxRedirects int) *Request {
	r.chain.enter("WithMaxRedirects()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if maxRedirects < 0 {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{maxRedirects},
			Errors: []error{
				errors.New("invalid negative argument"),
			},
		})
		return r
	}

	r.maxRedirects = maxRedirects

	return r
}

// RetryPolicy defines how failed requests are retried.
//
// Whether a request is retried depends on error type (if any), response
// status code (if any), and retry policy.
type RetryPolicy int

const (
	// DontRetry disables retrying at all.
	DontRetry RetryPolicy = iota

	// RetryTemporaryNetworkErrors enables retrying only temporary network errors.
	// Retry happens if Client returns net.Error and its Temporary() method
	// returns true.
	RetryTemporaryNetworkErrors

	// RetryTemporaryNetworkAndServerErrors enables retrying of temporary network
	// errors, as well as 5xx status codes.
	RetryTemporaryNetworkAndServerErrors

	// RetryAllErrors enables retrying of any error or 4xx/5xx status code.
	RetryAllErrors
)

// WithRetryPolicy sets policy for retries.
//
// Whether a request is retried depends on error type (if any), response
// status code (if any), and retry policy.
//
// How much retry attempts happens is defined by WithMaxRetries().
// How much to wait between attempts is defined by WithRetryDelay().
//
// Default retry policy is RetryTemporaryNetworkAndServerErrors, but
// default maximum number of retries is zero, so no retries happen
// unless WithMaxRetries() is called.
//
// Example:
//
//	req := NewRequest(config, "POST", "/path")
//	req.WithRetryPolicy(RetryAllErrors)
//	req.Expect().Status(http.StatusOK)
func (r *Request) WithRetryPolicy(policy RetryPolicy) *Request {
	r.chain.enter("WithRetryPolicy()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.retryPolicy = policy

	return r
}

// WithMaxRetries sets maximum number of retry attempts.
//
// After first request failure, additional retry attempts may happen,
// depending on the retry policy.
//
// Setting this to zero disables retries, i.e. only one request is sent.
// Setting this to N enables retries, and up to N+1 requests may be sent.
//
// Default number of retries is zero, i.e. retries are disabled.
//
// Example:
//
//	req := NewRequest(config, "POST", "/path")
//	req.WithMaxRetries(1)
//	req.Expect().Status(http.StatusOK)
func (r *Request) WithMaxRetries(maxRetries int) *Request {
	r.chain.enter("WithMaxRetries()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if maxRetries < 0 {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{maxRetries},
			Errors: []error{
				errors.New("invalid negative argument"),
			},
		})
		return r
	}

	r.maxRetries = maxRetries

	return r
}

// WithRetryDelay sets minimum and maximum delay between retries.
//
// If multiple retry attempts happen, delay between attempts starts from
// minDelay and then grows exponentionally until it reaches maxDelay.
//
// Default delay range is [50ms; 5s].
//
// Example:
//
//	req := NewRequest(config, "POST", "/path")
//	req.WithRetryDelay(time.Second, time.Minute)
//	req.Expect().Status(http.StatusOK)
func (r *Request) WithRetryDelay(minDelay, maxDelay time.Duration) *Request {
	r.chain.enter("WithRetryDelay()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if !(minDelay <= maxDelay) {
		r.chain.fail(AssertionFailure{
			Type: AssertValid,
			Actual: &AssertionValue{
				[2]time.Duration{minDelay, maxDelay},
			},
			Errors: []error{
				errors.New("invalid delay range"),
			},
		})
		return r
	}

	r.minRetryDelay = minDelay
	r.maxRetryDelay = maxDelay

	return r
}

// WithWebsocketUpgrade enables upgrades the connection to websocket.
//
// At least the following fields are added to the request header:
//
//	Upgrade: websocket
//	Connection: Upgrade
//
// The actual set of header fields is define by the protocol implementation
// in the gorilla/websocket package.
//
// The user should then call the Response.Websocket() method which returns
// the Websocket instance. This instance can be used to send messages to the
// server, to inspect the received messages, and to close the websocket.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithWebsocketUpgrade()
//	ws := req.Expect().Status(http.StatusSwitchingProtocols).Websocket()
//	defer ws.Disconnect()
func (r *Request) WithWebsocketUpgrade() *Request {
	r.chain.enter("WithWebsocketUpgrade()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.wsUpgrade = true

	return r
}

// WithWebsocketDialer sets the custom websocket dialer.
//
// The new dialer overwrites Config.WebsocketDialer. It will be used once to establish
// the WebSocket connection and receive a response of handshake result.
//
// Example:
//
//	req := NewRequest(config, "GET", "/path")
//	req.WithWebsocketUpgrade()
//	req.WithWebsocketDialer(&websocket.Dialer{
//	  EnableCompression: false,
//	})
//	ws := req.Expect().Status(http.StatusSwitchingProtocols).Websocket()
//	defer ws.Disconnect()
func (r *Request) WithWebsocketDialer(dialer WebsocketDialer) *Request {
	r.chain.enter("WithWebsocketDialer()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if dialer == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	r.config.WebsocketDialer = dialer

	return r
}

// WithPath substitutes named parameters in url path.
//
// value is converted to string using fmt.Sprint(). If there is no named
// parameter '{key}' in url path, failure is reported.
//
// Named parameters are case-insensitive.
//
// Example:
//
//	req := NewRequest(config, "POST", "/repos/{user}/{repo}")
//	req.WithPath("user", "gavv")
//	req.WithPath("repo", "httpexpect")
//	// path will be "/repos/gavv/httpexpect"
func (r *Request) WithPath(key string, value interface{}) *Request {
	r.chain.enter("WithPath()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if value == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	r.withPath(key, value)

	return r
}

// WithPathObject substitutes multiple named parameters in url path.
//
// object should be map or struct. If object is struct, it's converted
// to map using https://github.com/fatih/structs. Structs may contain
// "path" struct tag, similar to "json" struct tag for json.Marshal().
//
// Each map value is converted to string using fmt.Sprint(). If there
// is no named parameter for some map '{key}' in url path, failure is
// reported.
//
// Named parameters are case-insensitive.
//
// Example:
//
//	type MyPath struct {
//	    Login string `path:"user"`
//	    Repo  string
//	}
//
//	req := NewRequest(config, "POST", "/repos/{user}/{repo}")
//	req.WithPathObject(MyPath{"gavv", "httpexpect"})
//	// path will be "/repos/gavv/httpexpect"
//
//	req := NewRequest(config, "POST", "/repos/{user}/{repo}")
//	req.WithPathObject(map[string]string{"user": "gavv", "repo": "httpexpect"})
//	// path will be "/repos/gavv/httpexpect"
func (r *Request) WithPathObject(object interface{}) *Request {
	r.chain.enter("WithPathObject()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if object == nil {
		return r
	}

	var (
		m  map[string]interface{}
		ok bool
	)
	if reflect.Indirect(reflect.ValueOf(object)).Kind() == reflect.Struct {
		s := structs.New(object)
		s.TagName = "path"
		m = s.Map()
	} else {
		m, ok = canonMap(r.chain, object)
		if !ok {
			return r
		}
	}

	for key, value := range m {
		r.withPath(key, value)
	}

	return r
}

func (r *Request) withPath(key string, value interface{}) {
	found := false

	path, err := interpol.WithFunc(r.path, func(k string, w io.Writer) error {
		if strings.EqualFold(k, key) {
			if value == nil {
				r.chain.fail(AssertionFailure{
					Type: AssertUsage,
					Errors: []error{
						errors.New("unexpected nil interpol argument"),
					},
				})
			} else {
				mustWrite(w, fmt.Sprint(value))
				found = true
			}
		} else {
			mustWrite(w, "{")
			mustWrite(w, k)
			mustWrite(w, "}")
		}
		return nil
	})

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{path},
			Errors: []error{
				errors.New("invalid interpol string"),
				err,
			},
		})
		return
	}

	if !found {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf("key %q not found in interpol string", key),
			},
		})
		return
	}

	r.path = path
}

// WithQuery adds query parameter to request URL.
//
// value is converted to string using fmt.Sprint() and urlencoded.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithQuery("a", 123)
//	req.WithQuery("b", "foo")
//	// URL is now http://example.com/path?a=123&b=foo
func (r *Request) WithQuery(key string, value interface{}) *Request {
	r.chain.enter("WithQuery()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if value == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil argument"),
			},
		})
		return r
	}

	if r.query == nil {
		r.query = make(url.Values)
	}
	r.query.Add(key, fmt.Sprint(value))

	return r
}

// WithQueryObject adds multiple query parameters to request URL.
//
// object is converted to query string using github.com/google/go-querystring
// if it's a struct or pointer to struct, or github.com/ajg/form otherwise.
//
// Various object types are supported. Structs may contain "url" struct tag,
// similar to "json" struct tag for json.Marshal().
//
// Example:
//
//	type MyURL struct {
//	    A int    `url:"a"`
//	    B string `url:"b"`
//	}
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithQueryObject(MyURL{A: 123, B: "foo"})
//	// URL is now http://example.com/path?a=123&b=foo
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithQueryObject(map[string]interface{}{"a": 123, "b": "foo"})
//	// URL is now http://example.com/path?a=123&b=foo
func (r *Request) WithQueryObject(object interface{}) *Request {
	r.chain.enter("WithQueryObject()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if object == nil {
		return r
	}

	var (
		q   url.Values
		err error
	)
	if reflect.Indirect(reflect.ValueOf(object)).Kind() == reflect.Struct {
		q, err = query.Values(object)
		if err != nil {
			r.chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{object},
				Errors: []error{
					errors.New("invalid query object"),
					err,
				},
			})
			return r
		}
	} else {
		q, err = form.EncodeToValues(object)
		if err != nil {
			r.chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{object},
				Errors: []error{
					errors.New("invalid query object"),
					err,
				},
			})
			return r
		}
	}

	if r.query == nil {
		r.query = make(url.Values)
	}
	for k, v := range q {
		r.query[k] = append(r.query[k], v...)
	}

	return r
}

// WithQueryString parses given query string and adds it to request URL.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithQuery("a", 11)
//	req.WithQueryString("b=22&c=33")
//	// URL is now http://example.com/path?a=11&bb=22&c=33
func (r *Request) WithQueryString(query string) *Request {
	r.chain.enter("WithQueryString()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	v, err := url.ParseQuery(query)

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{query},
			Errors: []error{
				errors.New("invalid query string"),
				err,
			},
		})
		return r
	}

	if r.query == nil {
		r.query = make(url.Values)
	}
	for k, v := range v {
		r.query[k] = append(r.query[k], v...)
	}

	return r
}

// WithURL sets request URL.
//
// This URL overwrites Config.BaseURL. Request path passed to NewRequest()
// is appended to this URL, separated by slash if necessary.
//
// Example:
//
//	req := NewRequest(config, "PUT", "/path")
//	req.WithURL("http://example.com")
//	// URL is now http://example.com/path
func (r *Request) WithURL(urlStr string) *Request {
	r.chain.enter("WithURL()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{urlStr},
			Errors: []error{
				errors.New("invalid url string"),
				err,
			},
		})
		return r
	}

	r.httpReq.URL = u

	return r
}

// WithHeaders adds given headers to request.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithHeaders(map[string]string{
//	    "Content-Type": "application/json",
//	})
func (r *Request) WithHeaders(headers map[string]string) *Request {
	r.chain.enter("WithHeaders()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	for k, v := range headers {
		r.withHeader(k, v)
	}

	return r
}

// WithHeader adds given single header to request.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithHeader("Content-Type", "application/json")
func (r *Request) WithHeader(k, v string) *Request {
	r.chain.enter("WithHeader()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.withHeader(k, v)

	return r
}

func (r *Request) withHeader(k, v string) {
	switch http.CanonicalHeaderKey(k) {
	case "Host":
		r.httpReq.Host = v

	case "Content-Type":
		if !r.forceType {
			delete(r.httpReq.Header, "Content-Type")
		}
		r.forceType = true
		r.typeSetter = "WithHeader()"
		r.httpReq.Header.Add(k, v)

	default:
		r.httpReq.Header.Add(k, v)
	}
}

// WithCookies adds given cookies to request.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithCookies(map[string]string{
//	    "foo": "aa",
//	    "bar": "bb",
//	})
func (r *Request) WithCookies(cookies map[string]string) *Request {
	r.chain.enter("WithCookies()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	for k, v := range cookies {
		r.httpReq.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	return r
}

// WithCookie adds given single cookie to request.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithCookie("name", "value")
func (r *Request) WithCookie(k, v string) *Request {
	r.chain.enter("WithCookie()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.httpReq.AddCookie(&http.Cookie{
		Name:  k,
		Value: v,
	})

	return r
}

// WithBasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password
// are not encrypted.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithBasicAuth("john", "secret")
func (r *Request) WithBasicAuth(username, password string) *Request {
	r.chain.enter("WithBasicAuth()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.httpReq.SetBasicAuth(username, password)

	return r
}

// WithHost sets request host to given string.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithHost("example.com")
func (r *Request) WithHost(host string) *Request {
	r.chain.enter("WithHost()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.httpReq.Host = host

	return r
}

// WithProto sets HTTP protocol version.
//
// proto should have form of "HTTP/{major}.{minor}", e.g. "HTTP/1.1".
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithProto("HTTP/2.0")
func (r *Request) WithProto(proto string) *Request {
	r.chain.enter("WithProto()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	major, minor, ok := http.ParseHTTPVersion(proto)

	if !ok {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf(
					`unexpected protocol version %q, expected "HTTP/{major}.{minor}"`,
					proto),
			},
		})
		return r
	}

	r.httpReq.ProtoMajor = major
	r.httpReq.ProtoMinor = minor

	return r
}

// WithChunked enables chunked encoding and sets request body reader.
//
// Expect() will read all available data from given reader. Content-Length
// is not set, and "chunked" Transfer-Encoding is used.
//
// If protocol version is not at least HTTP/1.1 (required for chunked
// encoding), failure is reported.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/upload")
//	fh, _ := os.Open("data")
//	defer fh.Close()
//	req.WithHeader("Content-Type", "application/octet-stream")
//	req.WithChunked(fh)
func (r *Request) WithChunked(reader io.Reader) *Request {
	r.chain.enter("WithChunked()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if !r.httpReq.ProtoAtLeast(1, 1) {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf(
					`chunked Transfer-Encoding requires at least "HTTP/1.1",`+
						` but "HTTP/%d.%d" is used`,
					r.httpReq.ProtoMajor, r.httpReq.ProtoMinor),
			},
		})
		return r
	}

	r.setBody("WithChunked()", reader, -1, false)

	return r
}

// WithBytes sets request body to given slice of bytes.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithHeader("Content-Type", "application/json")
//	req.WithBytes([]byte(`{"foo": 123}`))
func (r *Request) WithBytes(b []byte) *Request {
	r.chain.enter("WithBytes()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if b == nil {
		r.setBody("WithBytes()", nil, 0, false)
	} else {
		r.setBody("WithBytes()", bytes.NewReader(b), len(b), false)
	}

	return r
}

// WithText sets Content-Type header to "text/plain; charset=utf-8" and
// sets body to given string.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithText("hello, world!")
func (r *Request) WithText(s string) *Request {
	r.chain.enter("WithText()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.setType("WithText()", "text/plain; charset=utf-8", false)
	r.setBody("WithText()", strings.NewReader(s), len(s), false)

	return r
}

// WithJSON sets Content-Type header to "application/json; charset=utf-8"
// and sets body to object, marshaled using json.Marshal().
//
// Example:
//
//	type MyJSON struct {
//	    Foo int `json:"foo"`
//	}
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithJSON(MyJSON{Foo: 123})
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithJSON(map[string]interface{}{"foo": 123})
func (r *Request) WithJSON(object interface{}) *Request {
	r.chain.enter("WithJSON()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	b, err := json.Marshal(object)

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{object},
			Errors: []error{
				errors.New("invalid json object"),
				err,
			},
		})
		return r
	}

	r.setType("WithJSON()", "application/json; charset=utf-8", false)
	r.setBody("WithJSON()", bytes.NewReader(b), len(b), false)

	return r
}

// WithForm sets Content-Type header to "application/x-www-form-urlencoded"
// or (if WithMultipart() was called) "multipart/form-data", converts given
// object to url.Values using github.com/ajg/form, and adds it to request body.
//
// Various object types are supported, including maps and structs. Structs may
// contain "form" struct tag, similar to "json" struct tag for json.Marshal().
// See https://github.com/ajg/form for details.
//
// Multiple WithForm(), WithFormField(), and WithFile() calls may be combined.
// If WithMultipart() is called, it should be called first.
//
// Example:
//
//	type MyForm struct {
//	    Foo int `form:"foo"`
//	}
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithForm(MyForm{Foo: 123})
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithForm(map[string]interface{}{"foo": 123})
func (r *Request) WithForm(object interface{}) *Request {
	r.chain.enter("WithForm()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	f, err := form.EncodeToValues(object)

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{object},
			Errors: []error{
				errors.New("invalid form object"),
				err,
			},
		})
		return r
	}

	if r.multipart != nil {
		r.setType("WithForm()", "multipart/form-data", false)

		var keys []string
		for k := range f {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if err := r.multipart.WriteField(k, f[k][0]); err != nil {
				r.chain.fail(AssertionFailure{
					Type: AssertOperation,
					Errors: []error{
						fmt.Errorf("failed to write multipart form field %q", k),
						err,
					},
				})
				return r
			}
		}
	} else {
		r.setType("WithForm()", "application/x-www-form-urlencoded", false)

		if r.form == nil {
			r.form = make(url.Values)
		}
		for k, v := range f {
			r.form[k] = append(r.form[k], v...)
		}
	}

	return r
}

// WithFormField sets Content-Type header to "application/x-www-form-urlencoded"
// or (if WithMultipart() was called) "multipart/form-data", converts given
// value to string using fmt.Sprint(), and adds it to request body.
//
// Multiple WithForm(), WithFormField(), and WithFile() calls may be combined.
// If WithMultipart() is called, it should be called first.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithFormField("foo", 123).
//	    WithFormField("bar", 456)
func (r *Request) WithFormField(key string, value interface{}) *Request {
	r.chain.enter("WithFormField()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if r.multipart != nil {
		r.setType("WithFormField()", "multipart/form-data", false)

		err := r.multipart.WriteField(key, fmt.Sprint(value))
		if err != nil {
			r.chain.fail(AssertionFailure{
				Type: AssertOperation,
				Errors: []error{
					fmt.Errorf("failed to write multipart form field %q", key),
					err,
				},
			})
			return r
		}
	} else {
		r.setType("WithFormField()", "application/x-www-form-urlencoded", false)

		if r.form == nil {
			r.form = make(url.Values)
		}
		r.form[key] = append(r.form[key], fmt.Sprint(value))
	}

	return r
}

// WithFile sets Content-Type header to "multipart/form-data", reads given
// file and adds its contents to request body.
//
// If reader is given, it's used to read file contents. Otherwise, os.Open()
// is used to read a file with given path.
//
// Multiple WithForm(), WithFormField(), and WithFile() calls may be combined.
// WithMultipart() should be called before WithFile(), otherwise WithFile()
// fails.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithFile("avatar", "./john.png")
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	fh, _ := os.Open("./john.png")
//	req.WithMultipart().
//	    WithFile("avatar", "john.png", fh)
//	fh.Close()
func (r *Request) WithFile(key, path string, reader ...io.Reader) *Request {
	r.chain.enter("WithFile()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	if len(reader) > 1 {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple reader arguments"),
			},
		})
		return r
	}

	r.withFile("WithFile()", key, path, reader...)

	return r
}

// WithFileBytes is like WithFile, but uses given slice of bytes as the
// file contents.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	fh, _ := os.Open("./john.png")
//	b, _ := ioutil.ReadAll(fh)
//	req.WithMultipart().
//	    WithFileBytes("avatar", "john.png", b)
//	fh.Close()
func (r *Request) WithFileBytes(key, path string, data []byte) *Request {
	r.chain.enter("WithFileBytes()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.withFile("WithFileBytes()", key, path, bytes.NewReader(data))

	return r
}

func (r *Request) withFile(method, key, path string, reader ...io.Reader) {
	r.setType(method, "multipart/form-data", false)

	if r.multipart == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf("%s requires WithMultipart() to be called first", method),
			},
		})
		return
	}

	wr, err := r.multipart.CreateFormFile(key, path)
	if err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				fmt.Errorf(
					"failed to create form file with key %q and path %q",
					key, path),
				err,
			},
		})
		return
	}

	var rd io.Reader
	if len(reader) != 0 && reader[0] != nil {
		rd = reader[0]
	} else {
		f, err := os.Open(path)
		if err != nil {
			r.chain.fail(AssertionFailure{
				Type: AssertOperation,
				Errors: []error{
					fmt.Errorf("failed to open file %q", path),
					err,
				},
			})
			return
		}
		rd = f
		defer f.Close()
	}

	if _, err := io.Copy(wr, rd); err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				fmt.Errorf("failed to read file %q", path),
				err,
			},
		})
		return
	}
}

// WithMultipart sets Content-Type header to "multipart/form-data".
//
// After this call, WithForm() and WithFormField() switch to multipart
// form instead of urlencoded form.
//
// If WithMultipart() is called, it should be called before WithForm(),
// WithFormField(), and WithFile().
//
// WithFile() always requires WithMultipart() to be called first.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithMultipart().
//	    WithForm(map[string]interface{}{"foo": 123})
func (r *Request) WithMultipart() *Request {
	r.chain.enter("WithMultipart()")
	defer r.chain.leave()

	if r.chain.failed() {
		return r
	}

	r.setType("WithMultipart()", "multipart/form-data", false)

	if r.multipart == nil {
		r.formbuf = &bytes.Buffer{}
		r.multipart = multipart.NewWriter(r.formbuf)
		r.setBody("WithMultipart()", r.formbuf, 0, false)
	}

	return r
}

// Expect constructs http.Request, sends it, receives http.Response, and
// returns a new Response instance.
//
// Request is sent using Client interface, or WebsocketDialer in case of
// WebSocket request.
//
// Example:
//
//	req := NewRequest(config, "PUT", "http://example.com/path")
//	req.WithJSON(map[string]interface{}{"foo": 123})
//	resp := req.Expect()
//	resp.Status(http.StatusOK)
func (r *Request) Expect() *Response {
	r.chain.enter("Expect()")
	defer r.chain.leave()

	resp := r.roundTrip()

	if resp == nil {
		return newResponse(responseOpts{
			config: r.config,
			chain:  r.chain,
		})
	}

	for _, matcher := range r.matchers {
		matcher(resp)
	}

	return resp
}

func (r *Request) roundTrip() *Response {
	if !r.encodeRequest() {
		return nil
	}

	if r.wsUpgrade {
		if !r.encodeWebsocketRequest() {
			return nil
		}
	}

	for _, transform := range r.transforms {
		transform(r.httpReq)
	}

	var (
		httpResp *http.Response
		websock  *websocket.Conn
		elapsed  time.Duration
	)
	if r.wsUpgrade {
		httpResp, websock, elapsed = r.sendWebsocketRequest()
	} else {
		httpResp, elapsed = r.sendRequest()
	}

	if httpResp == nil {
		return nil
	}

	return newResponse(responseOpts{
		config:    r.config,
		chain:     r.chain,
		httpResp:  httpResp,
		websocket: websock,
		rtt:       []time.Duration{elapsed},
	})
}

func (r *Request) encodeRequest() bool {
	if r.chain.failed() {
		return false
	}

	r.httpReq.URL.Path = concatPaths(r.httpReq.URL.Path, r.path)

	if r.query != nil {
		r.httpReq.URL.RawQuery = r.query.Encode()
	}

	if r.multipart != nil {
		if err := r.multipart.Close(); err != nil {
			r.chain.fail(AssertionFailure{
				Type: AssertOperation,
				Errors: []error{
					errors.New("failed to close multipart form"),
					err,
				},
			})
			return false
		}

		r.setType("Expect()", r.multipart.FormDataContentType(), true)
		r.setBody("Expect()", r.formbuf, r.formbuf.Len(), true)
	} else if r.form != nil {
		s := r.form.Encode()
		r.setBody("WithForm() or WithFormField()", strings.NewReader(s), len(s), false)
	}

	if r.httpReq.Body == nil {
		r.httpReq.Body = http.NoBody
	}

	if r.config.Context != nil {
		r.httpReq = r.httpReq.WithContext(r.config.Context)
	}

	r.setupRedirects()

	return true
}

var websocketErr = `webocket request can not have body:
  body was set by %s
  webocket was enabled by WithWebsocketUpgrade()`

func (r *Request) encodeWebsocketRequest() bool {
	if r.chain.failed() {
		return false
	}

	if r.bodySetter != "" {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf(websocketErr, r.bodySetter),
			},
		})
		return false
	}

	switch r.httpReq.URL.Scheme {
	case "https":
		r.httpReq.URL.Scheme = "wss"
	default:
		r.httpReq.URL.Scheme = "ws"
	}

	return true
}

func (r *Request) sendRequest() (*http.Response, time.Duration) {
	if r.chain.failed() {
		return nil, 0
	}

	resp, elapsed, err := r.retryRequest(func() (*http.Response, error) {
		return r.config.Client.Do(r.httpReq)
	})

	if err != nil {
		r.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to send http request"),
				err,
			},
		})
		return nil, 0
	}

	return resp, elapsed
}

func (r *Request) sendWebsocketRequest() (
	*http.Response, *websocket.Conn, time.Duration,
) {
	if r.chain.failed() {
		return nil, nil, 0
	}

	var conn *websocket.Conn
	resp, elapsed, err := r.retryRequest(func() (resp *http.Response, err error) {
		conn, resp, err = r.config.WebsocketDialer.Dial(
			r.httpReq.URL.String(), r.httpReq.Header)
		return resp, err
	})

	if err != nil && err != websocket.ErrBadHandshake {
		r.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to send websocket request"),
				err,
			},
		})
		return nil, nil, 0
	}

	if conn == nil {
		r.chain.fail(AssertionFailure{
			Type: AssertOperation,
			Errors: []error{
				errors.New("failed to upgrade connection to websocket"),
			},
		})
		return nil, nil, 0
	}

	return resp, conn, elapsed
}

func (r *Request) retryRequest(reqFunc func() (*http.Response, error)) (
	*http.Response, time.Duration, error,
) {
	if r.httpReq.Body != nil && r.httpReq.Body != http.NoBody {
		if _, ok := r.httpReq.Body.(*bodyWrapper); !ok {
			r.httpReq.Body = newBodyWrapper(r.httpReq.Body, nil)
		}
	}

	reqBody, _ := r.httpReq.Body.(*bodyWrapper)

	delay := r.minRetryDelay
	i := 0

	for {
		for _, printer := range r.config.Printers {
			if reqBody != nil {
				reqBody.Rewind()
			}
			printer.Request(r.httpReq)
		}

		if reqBody != nil {
			reqBody.Rewind()
		}

		var cancelFn context.CancelFunc

		if r.timeout > 0 {
			var ctx context.Context
			if r.config.Context != nil {
				ctx, cancelFn = context.WithTimeout(r.config.Context, r.timeout)
			} else {
				ctx, cancelFn = context.WithTimeout(context.Background(), r.timeout)
			}

			r.httpReq = r.httpReq.WithContext(ctx)
		}

		start := time.Now()
		resp, err := reqFunc()
		elapsed := time.Since(start)

		if resp != nil && resp.Body != nil {
			resp.Body = newBodyWrapper(resp.Body, cancelFn)
		} else if cancelFn != nil {
			cancelFn()
		}

		if resp != nil {
			for _, printer := range r.config.Printers {
				if resp.Body != nil {
					resp.Body.(*bodyWrapper).Rewind()
				}
				printer.Response(resp, elapsed)
			}
		}

		i++
		if i == r.maxRetries+1 {
			return resp, elapsed, err
		}

		if !r.shouldRetry(resp, err) {
			return resp, elapsed, err
		}

		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		r.sleepFn(delay)

		delay *= 2
		if delay > r.maxRetryDelay {
			delay = r.maxRetryDelay
		}
	}
}

func (r *Request) shouldRetry(resp *http.Response, err error) bool {
	var (
		isTemporaryNetworkError bool
		isTemporaryServerError  bool
		isHTTPError             bool
	)

	if netErr, ok := err.(net.Error); ok {
		//nolint
		isTemporaryNetworkError = netErr.Temporary()
	}

	if resp != nil {
		isTemporaryServerError = resp.StatusCode >= 500 && resp.StatusCode <= 599
		isHTTPError = resp.StatusCode >= 400 && resp.StatusCode <= 599
	}

	switch r.retryPolicy {
	case DontRetry:
		break

	case RetryTemporaryNetworkErrors:
		return isTemporaryNetworkError

	case RetryTemporaryNetworkAndServerErrors:
		return isTemporaryNetworkError || isTemporaryServerError

	case RetryAllErrors:
		return err != nil || isHTTPError
	}

	return false
}

func (r *Request) setupRedirects() {
	httpClient, _ := r.config.Client.(*http.Client)

	if httpClient == nil {
		if r.redirectPolicy != defaultRedirectPolicy {
			r.chain.fail(AssertionFailure{
				Type: AssertUsage,
				Errors: []error{
					errors.New(
						"WithRedirectPolicy() can be used only if Client is *http.Client"),
				},
			})
			return
		}

		if r.maxRedirects != -1 {
			r.chain.fail(AssertionFailure{
				Type: AssertUsage,
				Errors: []error{
					errors.New(
						"WithMaxRedirects() can be used only if Client is *http.Client"),
				},
			})
			return
		}
	} else {
		if r.redirectPolicy != defaultRedirectPolicy || r.maxRedirects != -1 {
			clientCopy := *httpClient
			httpClient = &clientCopy
			r.config.Client = &clientCopy
		}
	}

	if r.redirectPolicy == DontFollowRedirects {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if r.maxRedirects >= 0 {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) > r.maxRedirects {
				return fmt.Errorf("stopped after %d redirects", r.maxRedirects)
			}
			return nil
		}
	} else if r.redirectPolicy != defaultRedirectPolicy {
		httpClient.CheckRedirect = nil
	}

	if r.redirectPolicy == FollowAllRedirects {
		if r.httpReq.Body != nil && r.httpReq.Body != http.NoBody {
			if _, ok := r.httpReq.Body.(*bodyWrapper); !ok {
				r.httpReq.Body = newBodyWrapper(r.httpReq.Body, nil)
			}
			r.httpReq.GetBody = r.httpReq.Body.(*bodyWrapper).GetBody
		} else {
			r.httpReq.GetBody = func() (io.ReadCloser, error) {
				return http.NoBody, nil
			}
		}
	} else if r.redirectPolicy != defaultRedirectPolicy {
		r.httpReq.GetBody = nil
	}
}

var typeErr = `ambiguous request "Content-Type" header values:
  first set by %s:
    %q
  then replaced by %s:
    %q`

func (r *Request) setType(newSetter, newType string, overwrite bool) {
	if r.forceType {
		return
	}

	if !overwrite {
		previousType := r.httpReq.Header.Get("Content-Type")

		if previousType != "" && previousType != newType {
			r.chain.fail(AssertionFailure{
				Type: AssertUsage,
				Errors: []error{
					fmt.Errorf(typeErr,
						r.typeSetter, previousType, newSetter, newType),
				},
			})
			return
		}
	}

	r.typeSetter = newSetter
	r.httpReq.Header["Content-Type"] = []string{newType}
}

var bodyErr = `ambiguous request body contents:
  first set by %s
  then replaced by %s`

func (r *Request) setBody(setter string, reader io.Reader, len int, overwrite bool) {
	if !overwrite && r.bodySetter != "" {
		r.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				fmt.Errorf(bodyErr, r.bodySetter, setter),
			},
		})
		return
	}

	if len > 0 && reader == nil {
		panic("invalid length")
	}

	if reader == nil {
		r.httpReq.Body = http.NoBody
		r.httpReq.ContentLength = 0
	} else {
		r.httpReq.Body = ioutil.NopCloser(reader)
		r.httpReq.ContentLength = int64(len)
	}

	r.bodySetter = setter
}

func concatPaths(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	a = strings.TrimSuffix(a, "/")
	b = strings.TrimPrefix(b, "/")
	return a + "/" + b
}

func mustWrite(w io.Writer, s string) {
	_, err := w.Write([]byte(s))
	if err != nil {
		panic(err)
	}
}
