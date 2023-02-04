package httpexpect

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrinter_Compact(t *testing.T) {
	printer := NewCompactPrinter(t)

	body1 := bytes.NewBufferString("body1")
	body2 := bytes.NewBufferString("body2")

	req1, _ := http.NewRequest("GET", "http://example.com", body1)
	req2, _ := http.NewRequest("GET", "http://example.com", nil)

	printer.Request(req1)
	printer.Request(req2)
	printer.Request(nil)

	printer.Response(&http.Response{Body: ioutil.NopCloser(body2)}, 0)
	printer.Response(&http.Response{}, 0)
	printer.Response(nil, 0)
}

func TestPrinter_Debug(t *testing.T) {
	printer := NewDebugPrinter(t, true)

	body1 := bytes.NewBufferString("body1")
	body2 := bytes.NewBufferString("body2")

	req1, _ := http.NewRequest("GET", "http://example.com", body1)
	req2, _ := http.NewRequest("GET", "http://example.com", nil)

	printer.Request(req1)
	printer.Request(req2)
	printer.Request(nil)

	printer.Response(&http.Response{Body: ioutil.NopCloser(body2)}, 0)
	printer.Response(&http.Response{}, 0)
	printer.Response(nil, 0)
}

type errorReader struct{}

func (errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("error")
}

func (errorReader) Close() error {
	return errors.New("error")
}

func TestPrinter_Panics(t *testing.T) {
	t.Run("CurlPrinter", func(t *testing.T) {
		curl := NewCurlPrinter(t)

		assert.Panics(t, func() {
			curl.Request(&http.Request{})
		})
	})

	t.Run("DebugPrinter", func(t *testing.T) {
		curl := NewDebugPrinter(t, true)

		assert.Panics(t, func() {
			curl.Request(&http.Request{
				Body: errorReader{},
			})
		})

		assert.Panics(t, func() {
			curl.Response(&http.Response{
				Body: errorReader{},
			}, 0)
		})
	})
}
