# httpexpect [![GoDoc](https://godoc.org/github.com/gavv/httpexpect?status.svg)](https://godoc.org/github.com/gavv/httpexpect) [![Travis-CI](https://travis-ci.org/gavv/httpexpect.svg?branch=master)](https://travis-ci.org/gavv/httpexpect) [![Coveralls](https://coveralls.io/repos/github/gavv/httpexpect/badge.svg?branch=master)](https://coveralls.io/github/gavv/httpexpect?branch=master)

*Go module that helps to write nice tests for your HTTP API.*

## Features

* Incrementally build HTTP requests.
* Inspect HTTP responses.
* Inspect JSON payload, recursively (supported types: object, array, string, number, boolean, null).
* By default, uses [`testify`](https://github.com/stretchr/testify/) to report failures (can be configured to use `assert` or `require` package).
* May use [`httputil`](https://golang.org/pkg/net/http/httputil/) to dump requests and responses, or more compact logger.
* Produces nice diff on failure, using [`gojsondiff`](https://github.com/yudai/gojsondiff/).
* Configurable (accepts custom implementations of failure reporter, HTTP client, and logger).

## Documentation

Documentation is available on [GoDoc](https://godoc.org/github.com/gavv/httpexpect).

## Installation

```
$ go get github.com/gavv/httpexpect
```

## Example

See [`example`](example) directory for complete sources of fruits server and test.

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
$ go test github.com/gavv/httpexpect/...
```

## License

[MIT](LICENSE)
