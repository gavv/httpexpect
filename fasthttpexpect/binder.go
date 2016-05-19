package fasthttpexpect

import (
	"github.com/valyala/fasthttp"
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
