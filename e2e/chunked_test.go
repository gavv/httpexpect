package e2e

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
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

func testChunkedHandler(e *httpexpect.Expect) {
	e.PUT("/").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithChunked(strings.NewReader("key=value")).
		Expect().
		Status(http.StatusOK).
		HasContentType("application/json").
		HasTransferEncoding("chunked").
		JSON().Array().ConsistsOf(1, 2)
}

func TestE2EChunked_Live(t *testing.T) {
	handler := createChunkedHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testChunkedHandler(httpexpect.Default(t, server.URL))
}

func TestE2EChunked_BinderStandard(t *testing.T) {
	handler := createChunkedHandler()

	testChunkedHandler(httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
		},
	}))
}

func TestE2EChunked_BinderFast(t *testing.T) {
	handler := createChunkedFastHandler(t)

	testChunkedHandler(httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
		},
	}))
}

func TestE2EChunked_ResponseReader(t *testing.T) {
	const chars = "abcdefghijklmnopqrstuvwxyz"

	doneCh := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		for {
			wb := make([]byte, len(chars)*10)
			for i := range wb {
				wb[i] = chars[i%26]
			}
			_, err := w.Write(wb)
			if err != nil {
				break
			}
			w.(http.Flusher).Flush()
		}
		close(doneCh)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	resp := e.GET("/test").Expect()

	resp.Status(http.StatusOK).
		HasContentType("text/plain").
		HasTransferEncoding("chunked")

	reader := resp.Reader()

	rb := make([]byte, 1000000)
	n, err := io.ReadFull(reader, rb)
	assert.NoError(t, err)
	assert.Equal(t, 1000000, n)

	diff := 0
	for i := range rb {
		if rb[i] != chars[i%26] {
			diff++
		}
	}
	assert.Equal(t, 0, diff)

	err = reader.Close()
	assert.NoError(t, err)

	<-doneCh
}
