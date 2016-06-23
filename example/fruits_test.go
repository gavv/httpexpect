package example

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
)

func runFruitsTests(e *httpexpect.Expect) {
	e.GET("/fruits").
		Expect().
		Status(http.StatusOK).JSON().Array().Empty()

	orange := map[string]interface{}{
		"weight": 100,
	}

	e.PUT("/fruits/orange").WithJSON(orange).
		Expect().
		Status(http.StatusNoContent).NoContent()

	apple := map[string]interface{}{
		"colors": []interface{}{"green", "red"},
		"weight": 200,
	}

	e.PUT("/fruits/apple").WithJSON(apple).
		Expect().
		Status(http.StatusNoContent).NoContent()

	e.GET("/fruits").
		Expect().
		Status(http.StatusOK).JSON().Array().ContainsOnly("orange", "apple")

	e.GET("/fruits/orange").
		Expect().
		Status(http.StatusOK).JSON().Object().Equal(orange).NotEqual(apple)

	e.GET("/fruits/orange").
		Expect().
		Status(http.StatusOK).
		JSON().Object().ContainsKey("weight").ValueEqual("weight", 100)

	obj := e.GET("/fruits/apple").
		Expect().
		Status(http.StatusOK).JSON().Object()

	obj.Keys().ContainsOnly("colors", "weight")

	obj.Value("colors").Array().Elements("green", "red")
	obj.Value("colors").Array().Element(0).String().Equal("green")
	obj.Value("colors").Array().Element(1).String().Equal("red")

	obj.Value("weight").Number().Equal(200)

	e.GET("/fruits/melon").
		Expect().
		Status(http.StatusNotFound)
}

func TestFruits_DefaultClient(t *testing.T) {
	// create http.Handler
	handler := FruitServer()

	// start server using httptest
	server := httptest.NewServer(handler)
	defer server.Close()

	// create httpexpect instance using http.DefaultClient
	e := httpexpect.New(t, server.URL)

	// run tests
	runFruitsTests(e)
}

func TestFruits_CustomClientAndConfig(t *testing.T) {
	// create http.Handler
	handler := FruitServer()

	// start server using httptest
	server := httptest.NewServer(handler)
	defer server.Close()

	// create httpexpect instance using custom config
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: server.URL,
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCurlPrinter(t),
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	// run tests
	runFruitsTests(e)
}

func TestFruits_UseHandlerDirectly(t *testing.T) {
	// create http.Handler
	handler := FruitServer()

	// create httpexpect instance that will call htpp.Handler directly
	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   httpexpect.NewBinder(handler),
	})

	// run tests
	runFruitsTests(e)
}
