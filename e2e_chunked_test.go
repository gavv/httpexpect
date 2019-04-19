package httpexpect

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createChunkedHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Proto != "HTTP/1.1" {
			w.WriteHeader(http.StatusBadRequest)
		} else if len(r.TransferEncoding) != 1 || r.TransferEncoding[0] != "chunked" {
			w.WriteHeader(http.StatusBadRequest)
		} else if r.PostFormValue("key") != "value" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[1, `))
			w.(http.Flusher).Flush()
			_, _ = w.Write([]byte(`2]`))
		}
	})

	return mux
}

func createChunkedFastHandler(t *testing.T) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		headers := map[string][]string{}

		ctx.Request.Header.VisitAll(func(k, v []byte) {
			headers[string(k)] = append(headers[string(k)], string(v))
		})

		assert.Equal(t, []string{"chunked"}, headers["Transfer-Encoding"])
		assert.Equal(t, "value", string(ctx.FormValue("key")))
		assert.Equal(t, "key=value", string(ctx.Request.Body()))

		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBodyStreamWriter(func(w *bufio.Writer) {
			_, _ = w.WriteString(`[1, `)
			_ = w.Flush()
			_, _ = w.WriteString(`2]`)
		})
	}
}

func testChunkedHandler(e *Expect) {
	e.PUT("/").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithChunked(strings.NewReader("key=value")).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		TransferEncoding("chunked").
		JSON().Array().Elements(1, 2)
}

func TestE2EChunkedLive(t *testing.T) {
	handler := createChunkedHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testChunkedHandler(New(t, server.URL))
}

func TestE2EChunkedBinderStandard(t *testing.T) {
	handler := createChunkedHandler()

	testChunkedHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewBinder(handler),
		},
	}))
}

func TestE2EChunkedBinderFast(t *testing.T) {
	handler := createChunkedFastHandler(t)

	testChunkedHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewFastBinder(handler),
		},
	}))
}
