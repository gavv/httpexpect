package httpexpect

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createPrinterHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != "test_request" {
			panic("unexpected request body " + string(body))
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`test_response`))
	})

	return mux
}

func TestE2EPrinter(t *testing.T) {
	handler := createPrinterHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	p := &mockPrinter{}

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
		Printers: []Printer{
			p,
		},
	})

	e.POST("/test").
		WithText("test_request").
		Expect().
		Text().
		Equal("test_response")

	assert.Equal(t, "test_request", string(p.reqBody))
	assert.Equal(t, "test_response", string(p.respBody))
}

func TestE2EPrinterMultiple(t *testing.T) {
	handler := createPrinterHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	p1 := &mockPrinter{}
	p2 := &mockPrinter{}

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
		Printers: []Printer{
			p1,
			p2,
		},
	})

	e.POST("/test").
		WithText("test_request").
		Expect().
		Text().
		Equal("test_response")

	assert.Equal(t, "test_request", string(p1.reqBody))
	assert.Equal(t, "test_response", string(p1.respBody))

	assert.Equal(t, "test_request", string(p2.reqBody))
	assert.Equal(t, "test_response", string(p2.respBody))
}
