# httpexpect [![build status](https://travis-ci.org/gavv/httpexpect.svg?branch=master)](https://travis-ci.org/gavv/httpexpect) [![coverage status](https://coveralls.io/repos/github/gavv/httpexpect/badge.svg?branch=master)](https://coveralls.io/github/gavv/httpexpect?branch=master)

Go module that helps writing nice tests for HTTP API.

## Example

See [`example`](example) directory for complete sources of server and test.

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
		Status(http.StatusOK).JSON().Array().ElementsAnyOrder("orange", "apple")

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

	obj.Keys().ElementsAnyOrder("colors", "weight")

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

* [`go-json-rest/rest/test`](https://godoc.org/github.com/ant0ine/go-json-rest/rest/test)
* [`http-test`](https://github.com/vsco/http-test)

## License

[MIT](LICENSE)
