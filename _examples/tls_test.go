package examples

import (
	"crypto/tls"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

func TlsClient() *http.Client {
	cfg := tls.Config{RootCAs: NewRootCertPool()}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &cfg}}
}

func TestExampleTlsServer(t *testing.T) {

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

	e.PUT("/tls/car").WithText("1").
		Expect().
		Status(http.StatusNoContent).NoContent()

	e.PUT("/tls/car").WithText("10").
		Expect().
		Status(http.StatusNoContent).NoContent()

	e.GET("/tls/car").
		Expect().
		Status(http.StatusOK).Body().IsEqual("11")

	e.DELETE("/tls/car").WithText("1").
		Expect().
		Status(http.StatusNoContent).NoContent()

	e.GET("/tls/car").
		Expect().
		Status(http.StatusOK).Body().IsEqual("10")

	e.GET("/tls/not_there").
		Expect().
		Status(http.StatusNotFound).NoContent()

	e.DELETE("/tls/car").WithText("10").
		Expect().
		Status(http.StatusNoContent).NoContent()

	items := map[string]int{
		"car":    2,
		"house":  1,
		"cat":    6,
		"fridge": 3,
	}
	for item, amount := range items {
		e.PUT("/tls/" + item).WithText(strconv.Itoa(amount)).
			Expect().
			Status(http.StatusNoContent).NoContent()
	}

	e.DELETE("/tls/car").WithText("1").
		Expect().
		Status(http.StatusNoContent).NoContent()

	items["car"] -= 1

	var m map[string]int

	r := e.GET("/tls").Expect().Status(http.StatusOK)

	r.Header("Content-Type").IsEqual("application/json")

	r.JSON().Decode(&m)

	for item, amount := range m {
		assert.Equal(t, amount, items[item])
	}

}
