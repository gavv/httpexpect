# httpexpect [![build status](https://travis-ci.org/gavv/httpexpect.svg?branch=master)](https://travis-ci.org/gavv/httpexpect) [![coverage status](https://coveralls.io/repos/github/gavv/httpexpect/badge.svg?branch=master)](https://coveralls.io/github/gavv/httpexpect?branch=master)

Go module that helps writing nice tests for HTTP API.

*Work in progress.*

## Example

```go
import (
    "github.com/gavv/httpexpect"
    "net/http"
    "testing"
)

func TestUsers(t *testing.T) {
    e := httpexpect.New(t, "http://127.0.0.1:8080")

    e.GET("/users").
        ExpectCode(http.StatusOK).ExpectList()

    e.GET("/users/john").
        ExpectCode(http.StatusNotFound)

    user1 := map[string]interface{}{
        "login": "john",
    }

    user2 := map[string]interface{}{
        "login": "bob",
    }

    e.POST("/users", user1).
        ExpectCode(http.StatusCreated).ExpectEmpty()

    e.POST("/users", user2).
        ExpectCode(http.StatusCreated).ExpectEmpty()

    e.GET("/users").
        ExpectCode(http.StatusOK).ExpectList(user1, user2)

    e.GET("/users/john").
        ExpectCode(http.StatusOK).ExpectMap(user1)
}
```

## Similar modules

* [`go-json-rest/rest/test`](https://godoc.org/github.com/ant0ine/go-json-rest/rest/test)
* [`http-test`](https://github.com/vsco/http-test)

## License

[MIT](LICENSE)
