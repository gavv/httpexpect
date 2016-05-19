package fasthttpexpect

import (
	"github.com/valyala/fasthttp"
	"net/http"
)

// ClientBackend defines interface compatible with various fasthttp clients.
// fasthttp.Client, fasthttp.HostClient, fasthttp.PipelineClient implement
// this interface.
type ClientBackend interface {
	// Do sends request and returns response.
	Do(*fasthttp.Request, *fasthttp.Response) error
}

// ClientAdapter wraps ClientBackend and implements httpexpect.Client.
type ClientAdapter struct {
	backend ClientBackend
}

var (
	defaultBackend fasthttp.Client
)

// NewClient returns a new adapater for default fasthttp.Client.
func NewClient() ClientAdapter {
	return WithClient(&defaultBackend)
}

// WithClient returns a new adapter for custom fasthttp.Client.
func WithClient(backend ClientBackend) ClientAdapter {
	return ClientAdapter{backend}
}

// Do implements httpexpect.Client.Do.
func (adapter ClientAdapter) Do(stdreq *http.Request) (stdresp *http.Response, err error) {
	var fastreq fasthttp.Request

	convertRequest(stdreq, &fastreq)

	var fastresp fasthttp.Response

	if err = adapter.backend.Do(&fastreq, &fastresp); err == nil {
		stdresp = convertResponse(stdreq, &fastresp)
	}

	return
}
