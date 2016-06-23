package fasthttpexpect

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
)

// Binder implements networkless httpexpect.Client attached directly to
// fasthttp.RequestHandler.
type Binder struct {
	handler fasthttp.RequestHandler
}

// NewBinder returns a new Binder given fasthttp.RequestHandler.
func NewBinder(handler fasthttp.RequestHandler) *Binder {
	return &Binder{handler}
}

// Do implements httpexpect.Client.Do.
func (binder *Binder) Do(stdreq *http.Request) (*http.Response, error) {
	var fastreq fasthttp.Request

	convertRequest(stdreq, &fastreq)

	var ctx fasthttp.RequestCtx

	ctx.Init(&fastreq, nil, nil)

	if stdreq.Body != nil {
		ctx.Request.SetBodyStream(stdreq.Body, -1)
	}

	binder.handler(&ctx)

	return convertResponse(stdreq, &ctx.Response), nil
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
