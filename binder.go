package httpexpect

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// Binder implements networkless Client attached directly to http.Handler.
//
// Binder emulates network communication by invoking given http.Handler. It
// passes httptest.ResponseRecorder as http.ResponseWriter to the handler,
// and then constructs http.Response from recorded data.
type Binder struct {
	// Handler specifies the function to be invoked.
	Handler http.Handler

	// Jar specifies the cookie jar.
	// If Jar is nil, cookies are not sent in requests and ignored
	// in responses.
	Jar http.CookieJar
}

// NewBinder returns a new Binder given http.Handler.
// It uses DefaultJar() as cookie jar.
func NewBinder(handler http.Handler) *Binder {
	return &Binder{
		Handler: handler,
		Jar:     DefaultJar(),
	}
}

// Do implements Client.Do.
func (binder *Binder) Do(req *http.Request) (*http.Response, error) {
	if binder.Jar != nil {
		for _, cookie := range binder.Jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	recorder := httptest.NewRecorder()

	binder.Handler.ServeHTTP(recorder, req)

	resp := http.Response{
		Request:    req,
		StatusCode: recorder.Code,
		Status:     http.StatusText(recorder.Code),
		Header:     recorder.HeaderMap,
	}

	if recorder.Body != nil {
		resp.Body = ioutil.NopCloser(recorder.Body)
	}

	if binder.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			binder.Jar.SetCookies(req.URL, rc)
		}
	}

	return &resp, nil
}

// FastBinder implements networkless Client attached directly to
// fasthttp.RequestHandler.
type FastBinder struct {
	// Handler specifies the function to be invoked.
	Handler fasthttp.RequestHandler

	// Jar specifies the cookie jar.
	// If Jar is nil, cookies are not sent in requests and ignored
	// in responses.
	Jar http.CookieJar
}

// NewFastBinder returns a new FastBinder given a fasthttp.RequestHandler.
// It uses DefaultJar() as cookie jar.
func NewFastBinder(handler fasthttp.RequestHandler) *FastBinder {
	return &FastBinder{
		Handler: handler,
		Jar:     DefaultJar(),
	}
}

// Do implements Client.Do.
func (binder *FastBinder) Do(stdreq *http.Request) (*http.Response, error) {
	if binder.Jar != nil {
		for _, cookie := range binder.Jar.Cookies(stdreq.URL) {
			stdreq.AddCookie(cookie)
		}
	}

	var fastreq fasthttp.Request

	convertRequest(stdreq, &fastreq)

	var ctx fasthttp.RequestCtx

	ctx.Init(&fastreq, nil, nil)

	if stdreq.Body != nil {
		ctx.Request.SetBodyStream(stdreq.Body, -1)
	}

	binder.Handler(&ctx)

	stdresp := convertResponse(stdreq, &ctx.Response)

	if binder.Jar != nil {
		if rc := stdresp.Cookies(); len(rc) > 0 {
			binder.Jar.SetCookies(stdreq.URL, rc)
		}
	}

	return stdresp, nil
}

func convertRequest(stdreq *http.Request, fastreq *fasthttp.Request) {
	if stdreq.Body != nil {
		fastreq.SetBodyStream(stdreq.Body, -1)
	}

	fastreq.SetRequestURI(stdreq.URL.String())

	fastreq.Header.SetMethod(stdreq.Method)

	for k, a := range stdreq.Header {
		for _, v := range a {
			fastreq.Header.Add(k, v)
		}
	}
}

func convertResponse(stdreq *http.Request, fastresp *fasthttp.Response) *http.Response {
	status := fastresp.Header.StatusCode()
	body := fastresp.Body()

	stdresp := &http.Response{
		Request:    stdreq,
		StatusCode: status,
		Status:     http.StatusText(status),
	}

	fastresp.Header.VisitAll(func(k, v []byte) {
		if stdresp.Header == nil {
			stdresp.Header = make(http.Header)
		}
		stdresp.Header.Add(string(k), string(v))
	})

	if body != nil {
		stdresp.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	return stdresp
}
