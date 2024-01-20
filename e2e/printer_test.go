package e2e

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

type mockPrinter struct {
	reqBody  []byte
	respBody []byte
	rtt      time.Duration
}

func (p *mockPrinter) Request(req *http.Request) {
	if req.Body != nil {
		p.reqBody, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}
}

func (p *mockPrinter) Response(resp *http.Response, rtt time.Duration) {
	if resp.Body != nil {
		p.respBody, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}
	p.rtt = rtt
}

func createPrinterHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test_request" {
			panic("unexpected request body " + string(body))
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`test_response`))
	})

	return mux
}

func TestE2EPrinter_Single(t *testing.T) {
	handler := createPrinterHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	p := &mockPrinter{}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			p,
		},
	})

	e.POST("/test").
		WithText("test_request").
		Expect().
		Text().
		IsEqual("test_response")

	assert.Equal(t, "test_request", string(p.reqBody))
	assert.Equal(t, "test_response", string(p.respBody))
}

func TestE2EPrinter_Multiple(t *testing.T) {
	handler := createPrinterHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	p1 := &mockPrinter{}
	p2 := &mockPrinter{}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			p1,
			p2,
		},
	})

	e.POST("/test").
		WithText("test_request").
		Expect().
		Text().
		IsEqual("test_response")

	assert.Equal(t, "test_request", string(p1.reqBody))
	assert.Equal(t, "test_response", string(p1.respBody))

	assert.Equal(t, "test_request", string(p2.reqBody))
	assert.Equal(t, "test_response", string(p2.respBody))
}
