package examples

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/gavv/httpexpect/v2"
	"net/http"
	"testing"
)

func TlsClient() *http.Client {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}
	cfg := tls.Config{RootCAs: roots}

	return &http.Client{Transport: &http.Transport{TLSClientConfig: &cfg}}
}

func TestExampleTlsServer(t *testing.T) {
	create_c_k()
	server := ExampleTlsServer() // ExampleTlsServer()
	server.StartTLS()
	defer server.Close()

	config := httpexpect.Config{
		BaseURL: server.URL, Client: TlsClient(),
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		}}
	e := httpexpect.WithConfig(config)
	e.PUT("tls/car").WithText("1").Expect().Status(http.StatusNoContent).NoContent()
	e.GET("tls/car").Expect().Status(http.StatusOK).Body().IsEqual("1")
}
