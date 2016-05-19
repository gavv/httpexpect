// Package fasthttpexpect provides fasthttp adapter for httpexpect.
package fasthttpexpect

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"io"
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
	fastreq := fasthttp.AcquireRequest()

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

	var fastresp fasthttp.Response

	if err = adapter.backend.Do(fastreq, &fastresp); err == nil {
		status := fastresp.Header.StatusCode()
		body := fastresp.Body()

		stdresp = &http.Response{
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
			stdresp.Body = readCloserAdapter{bytes.NewReader(body)}
		}
	}

	fasthttp.ReleaseRequest(fastreq)

	return
}

type readCloserAdapter struct {
	io.Reader
}

func (b readCloserAdapter) Close() error {
	return nil
}
