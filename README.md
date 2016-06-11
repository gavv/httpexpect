# httpexpect [![GoDoc](https://godoc.org/github.com/gavv/httpexpect?status.svg)](https://godoc.org/github.com/gavv/httpexpect) [![Gitter](https://badges.gitter.im/gavv/httpexpect.svg)](https://gitter.im/gavv/httpexpect?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Wercker](https://app.wercker.com/status/47ee8e7bcc23183a0f6fc26ff9f2c47d/s "wercker status")](https://app.wercker.com/project/bykey/47ee8e7bcc23183a0f6fc26ff9f2c47d) [![Coveralls](https://coveralls.io/repos/github/gavv/httpexpect/badge.svg?branch=master)](https://coveralls.io/github/gavv/httpexpect?branch=master)

*Go module that helps to write nice tests for your HTTP API.*

## Features

**Essential:**

* Incrementally build HTTP requests.
* Inspect HTTP responses.
* Inspect JSON or form payload, recursively (supported types: object, array, string, number, boolean, null).

**Tuning:**
* Can communicate with server via HTTP client or invoke HTTP handler directly.
* Configurable (accepts custom implementations of failure reporter, HTTP client, and logger).

**Pretty printing:**
* By default, uses [`testify`](https://github.com/stretchr/testify/) to report failures (can be configured to use `assert` or `require` package).
* May dump requests and responses in various formats, using [`httputil`](https://golang.org/pkg/net/http/httputil/), [`http2curl`](https://github.com/moul/http2curl), or simple compact logger.
* Produces nice diff on failure, using [`gojsondiff`](https://github.com/yudai/gojsondiff/).

**Integrations:**
* Uses [`form`](https://github.com/ajg/form) and [`go-querystring`](https://github.com/google/go-querystring) packages to encode and decode forms and URL parameters.
* Provides integration with [`fasthttp`](https://github.com/valyala/fasthttp/) client and HTTP handler via `fasthttpexpect` module.

## Status

First stable release (v1) is planned soon. Prior to this, API may be changing slightly.

## Documentation

Documentation is available on [GoDoc](https://godoc.org/github.com/gavv/httpexpect). It contains an overview and reference.

## Installation

```
$ go get github.com/gavv/httpexpect
```

## Examples

See [`example`](example) directory for various usage examples.

* [`fruits_test.go`](example/fruits_test.go)

  Using httpexpect with default and custom config. Communicating with server via HTTP client or invoking `http.Handler` directly.

* [`echo_test.go`](example/echo_test.go)

  Using httpexpect with two http handlers created with [`echo`](https://github.com/labstack/echo/) framework: `http.Handler` and `fasthttp.RequestHandler`.

* [`iris_test.go`](example/iris_test.go)

  Using httpexpect with `fasthttp.RequestHandler` created with [`iris`](https://github.com/kataras/iris) framework.

## Quick start

Here is a complete example of end-to-end test for [`FruitServer`](example/fruits.go).

```go
import (
	"github.com/gavv/httpexpect"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFruits(t *testing.T) {
	server := httptest.NewServer(FruitServer())
	defer server.Close()

	e := httpexpect.New(t, server.URL)

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
```

## Similar modules

* [`gorequest`](https://github.com/parnurzeal/gorequest)
* [`gabs`](https://github.com/Jeffail/gabs)
* [`go-json-rest/rest/test`](https://godoc.org/github.com/ant0ine/go-json-rest/rest/test)
* [`http-test`](https://github.com/vsco/http-test)

## Contributing

Feel free to report bugs, suggest improvements, and send pull requests! Don't forget to add documentation and tests for new features and run all tests before submitting pull requests:

```
$ go test -bench . github.com/gavv/httpexpect/...
```

## License

[MIT](LICENSE)
