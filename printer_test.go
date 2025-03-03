package httpexpect

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type failingReader struct{}

func (failingReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("error")
}

func (failingReader) Close() error {
	return errors.New("error")
}

func httpRequest(body io.Reader) *http.Request {
	req, _ := http.NewRequest("GET", "http://example.com", body)
	return req
}

func httpResponse(body io.Reader) *http.Response {
	resp := &http.Response{}
	if body != nil {
		resp.Body = io.NopCloser(body)
	}
	return resp
}

func TestPrinter_Compact(t *testing.T) {
	cases := []struct {
		name     string
		request  *http.Request
		response *http.Response
	}{
		{
			name:     "nil request and response",
			request:  nil,
			response: nil,
		},
		{
			name:     "has body",
			request:  httpRequest(strings.NewReader("test")),
			response: httpResponse(strings.NewReader("test")),
		},
		{
			name:     "no body",
			request:  httpRequest(nil),
			response: httpResponse(nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			printer := NewCompactPrinter(t)

			printer.Request(tc.request)
			printer.Response(tc.response, 0)
		})
	}
}

func TestPrinter_Curl(t *testing.T) {
	cases := []struct {
		name     string
		request  *http.Request
		response *http.Response
	}{
		{
			name:     "nil request and response",
			request:  nil,
			response: nil,
		},
		{
			name:     "has body",
			request:  httpRequest(strings.NewReader("test")),
			response: httpResponse(strings.NewReader("test")),
		},
		{
			name:     "no body",
			request:  httpRequest(nil),
			response: httpResponse(nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			printer := NewCurlPrinter(t)

			printer.Request(tc.request)
			printer.Response(tc.response, 0)
		})
	}
}

func TestPrinter_Debug(t *testing.T) {
	cases := []struct {
		name      string
		printBody bool
		request   *http.Request
		response  *http.Response
	}{
		{
			name:      "nil request and response",
			printBody: true,
			request:   nil,
			response:  nil,
		},
		{
			name:      "has body, print body",
			printBody: true,
			request:   httpRequest(strings.NewReader("test")),
			response:  httpResponse(strings.NewReader("test")),
		},
		{
			name:      "has body, don't print body",
			printBody: false,
			request:   httpRequest(strings.NewReader("test")),
			response:  httpResponse(strings.NewReader("test")),
		},
		{
			name:      "no body, print body",
			printBody: true,
			request:   httpRequest(nil),
			response:  httpResponse(nil),
		},
		{
			name:      "no body, don't print body",
			printBody: false,
			request:   httpRequest(nil),
			response:  httpResponse(nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			printer := NewDebugPrinter(t, tc.printBody)

			printer.Request(tc.request)
			printer.Response(tc.response, 0)
		})
	}
}

func TestPrinter_Panics(t *testing.T) {
	t.Run("CompactPrinter", func(t *testing.T) {
		printer := NewCompactPrinter(t)

		assert.NotPanics(t, func() {
			printer.Request(&http.Request{})
		})
		assert.NotPanics(t, func() {
			printer.Response(&http.Response{
				Body: failingReader{},
			}, 0)
		})
	})

	t.Run("CurlPrinter", func(t *testing.T) {
		printer := NewCurlPrinter(t)

		assert.Panics(t, func() {
			printer.Request(&http.Request{})
		})
		assert.NotPanics(t, func() {
			printer.Response(&http.Response{
				Body: failingReader{},
			}, 0)
		})
	})

	t.Run("DebugPrinter", func(t *testing.T) {
		printer := NewDebugPrinter(t, true)

		assert.Panics(t, func() {
			printer.Request(&http.Request{
				Body: failingReader{},
			})
		})
		assert.Panics(t, func() {
			printer.Response(&http.Response{
				Body: failingReader{},
			}, 0)
		})
	})
}
