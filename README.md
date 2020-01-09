# httpexpect [![GoDoc](https://godoc.org/github.com/gavv/httpexpect?status.svg)](https://godoc.org/github.com/gavv/httpexpect) [![Travis](https://img.shields.io/travis/gavv/httpexpect.svg)](https://travis-ci.org/gavv/httpexpect) [![Coveralls](https://coveralls.io/repos/github/gavv/httpexpect/badge.svg?branch=master)](https://coveralls.io/github/gavv/httpexpect?branch=master)

Concise, declarative, and easy to use end-to-end HTTP and REST API testing for Go (golang).

Basically, httpexpect is a set of chainable *builders* for HTTP requests and *assertions* for HTTP responses and payload, on top of net/http and several utility packages.

Workflow:

* Incrementally build HTTP requests.
* Inspect HTTP responses.
* Inspect response payload recursively.

## Features

##### Request builder

* URL path construction, with simple string interpolation provided by [`go-interpol`](https://github.com/imkira/go-interpol) package.
* URL query parameters (encoding using [`go-querystring`](https://github.com/google/go-querystring) package).
* Headers, cookies, payload: JSON,  urlencoded or multipart forms (encoding using [`form`](https://github.com/ajg/form) package), plain text.
* Custom reusable [request builders](#reusable-builders).

##### Response assertions

* Response status, predefined status ranges.
* Headers, cookies, payload: JSON, JSONP, forms, text.
* Round-trip time.
* Custom reusable [response matchers](#reusable-matchers).

##### Payload assertions

* Type-specific assertions, supported types: object, array, string, number, boolean, null, datetime.
* Regular expressions.
* Simple JSON queries (using subset of [JSONPath](http://goessner.net/articles/JsonPath/)), provided by [`jsonpath`](https://github.com/yalp/jsonpath) package.
* [JSON Schema](http://json-schema.org/) validation, provided by [`gojsonschema`](https://github.com/xeipuuv/gojsonschema) package.

##### WebSocket support (thanks to [@tyranron](https://github.com/tyranron))

* Upgrade an HTTP connection to a WebSocket connection (we use [`gorilla/websocket`](https://github.com/gorilla/websocket) internally).
* Interact with the WebSocket server.
* Inspect WebSocket connection parameters and WebSocket messages.

##### Pretty printing

* Verbose error messages.
* JSON diff is produced on failure using [`gojsondiff`](https://github.com/yudai/gojsondiff/) package.
* Failures are reported using [`testify`](https://github.com/stretchr/testify/) (`assert` or `require` package) or standard `testing` package.
* Dumping requests and responses in various formats, using [`httputil`](https://golang.org/pkg/net/http/httputil/), [`http2curl`](https://github.com/moul/http2curl), or simple compact logger.

##### Tuning

* Tests can communicate with server via real HTTP client or invoke `net/http` or [`fasthttp`](https://github.com/valyala/fasthttp/) handler directly.
* Custom HTTP client, logger, printer, and failure reporter may be provided by user.
* Custom HTTP request factory may be provided, e.g. from the Google App Engine testing.

## Versions

The versions are selected according to the [semantic versioning](https://semver.org/) scheme. Every new major version gets its own stable branch with a backwards compatibility promise. Releases are tagged from stable branches.

The current stable branch is `v2`. Previous branches are still maintained, but no new features are added.

If you're using go.mod, use a versioned import path:

```go
import "github.com/gavv/httpexpect/v2"
```

Otherwise, use gopkg.in import path:

```go
import "gopkg.in/gavv/httpexpect.v2"
```

## Documentation

Documentation is available on [GoDoc](https://godoc.org/github.com/gavv/httpexpect). It contains an overview and reference.

## Examples

See [`_examples`](_examples) directory for complete standalone examples.

* [`fruits_test.go`](_examples/fruits_test.go)

    Testing a simple CRUD server made with bare `net/http`.

* [`iris_test.go`](_examples/iris_test.go)

    Testing a server made with [`iris`](https://github.com/kataras/iris/) framework. Example includes JSON queries and validation, URL and form parameters, basic auth, sessions, and streaming. Tests invoke the `http.Handler` directly.

* [`echo_test.go`](_examples/echo_test.go)

    Testing a server with JWT authentication made with [`echo`](https://github.com/labstack/echo/) framework. Tests use either HTTP client or invoke the `http.Handler` directly.

* [`gin_test.go`](_examples/gin_test.go)

    Testing a server utilizing the [`gin`](https://github.com/gin-gonic/gin) web framework. Tests invoke the `http.Handler` directly.

* [`fasthttp_test.go`](_examples/fasthttp_test.go)

    Testing a server made with [`fasthttp`](https://github.com/valyala/fasthttp) package. Tests invoke the `fasthttp.RequestHandler` directly.

* [`websocket_test.go`](_examples/websocket_test.go)

    Testing a WebSocket server based on [`gorilla/websocket`](https://github.com/gorilla/websocket). Tests invoke the `http.Handler` or `fasthttp.RequestHandler` directly.

* [`gae_test.go`](_examples/gae_test.go)

    Testing a server running under the [Google App Engine](https://en.wikipedia.org/wiki/Google_App_Engine).

## Quick start

##### Hello, world!

```go
package example

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestFruits(t *testing.T) {
	// create http.Handler
	handler := FruitsHandler()

	// run server using httptest
	server := httptest.NewServer(handler)
	defer server.Close()

	// create httpexpect instance
	e := httpexpect.New(t, server.URL)

	// is it working?
	e.GET("/fruits").
		Expect().
		Status(http.StatusOK).JSON().Array().Empty()
}
```

##### JSON

```go
orange := map[string]interface{}{
	"weight": 100,
}

e.PUT("/fruits/orange").WithJSON(orange).
	Expect().
	Status(http.StatusNoContent).NoContent()

e.GET("/fruits/orange").
	Expect().
	Status(http.StatusOK).
	JSON().Object().ContainsKey("weight").ValueEqual("weight", 100)

apple := map[string]interface{}{
	"colors": []interface{}{"green", "red"},
	"weight": 200,
}

e.PUT("/fruits/apple").WithJSON(apple).
	Expect().
	Status(http.StatusNoContent).NoContent()

obj := e.GET("/fruits/apple").
	Expect().
	Status(http.StatusOK).JSON().Object()

obj.Keys().ContainsOnly("colors", "weight")

obj.Value("colors").Array().Elements("green", "red")
obj.Value("colors").Array().Element(0).String().Equal("green")
obj.Value("colors").Array().Element(1).String().Equal("red")
obj.Value("colors").Array().First().String().Equal("green")
obj.Value("colors").Array().Last().String().Equal("red")
```

##### JSON Schema and JSON Path

```go
schema := `{
	"type": "array",
	"items": {
		"type": "object",
		"properties": {
			...
			"private": {
				"type": "boolean"
			}
		}
	}
}`

repos := e.GET("/repos/octocat").
	Expect().
	Status(http.StatusOK).JSON()

// validate JSON schema
repos.Schema(schema)

// run JSONPath query and iterate results
for _, private := range repos.Path("$..private").Array().Iter() {
	private.Boolean().False()
}
```

##### Forms

```go
// post form encoded from struct or map
e.POST("/form").WithForm(structOrMap).
	Expect().
	Status(http.StatusOK)

// set individual fields
e.POST("/form").WithFormField("foo", "hello").WithFormField("bar", 123).
	Expect().
	Status(http.StatusOK)

// multipart form
e.POST("/form").WithMultipart().
	WithFile("avatar", "./john.png").WithFormField("username", "john").
	Expect().
	Status(http.StatusOK)
```

##### URL construction

```go
// construct path using ordered parameters
e.GET("/repos/{user}/{repo}", "octocat", "hello-world").
	Expect().
	Status(http.StatusOK)

// construct path using named parameters
e.GET("/repos/{user}/{repo}").
	WithPath("user", "octocat").WithPath("repo", "hello-world").
	Expect().
	Status(http.StatusOK)

// set query parameters
e.GET("/repos/{user}", "octocat").WithQuery("sort", "asc").
	Expect().
	Status(http.StatusOK)    // "/repos/octocat?sort=asc"
```

##### Headers

```go
// set If-Match
e.POST("/users/john").WithHeader("If-Match", etag).WithJSON(john).
	Expect().
	Status(http.StatusOK)

// check ETag
e.GET("/users/john").
	Expect().
	Status(http.StatusOK).Header("ETag").NotEmpty()

// check Date
t := time.Now()

e.GET("/users/john").
	Expect().
	Status(http.StatusOK).Header("Date").DateTime().InRange(t, time.Now())
```

##### Cookies

```go
// set cookie
t := time.Now()

e.POST("/users/john").WithCookie("session", sessionID).WithJSON(john).
	Expect().
	Status(http.StatusOK)

// check cookies
c := e.GET("/users/john").
	Expect().
	Status(http.StatusOK).Cookie("session")

c.Value().Equal(sessionID)
c.Domain().Equal("example.com")
c.Path().Equal("/")
c.Expires().InRange(t, t.Add(time.Hour * 24))
```

##### Regular expressions

```go
// simple match
e.GET("/users/john").
	Expect().
	Header("Location").
	Match("http://(.+)/users/(.+)").Values("example.com", "john")

// check capture groups by index or name
m := e.GET("/users/john").
	Expect().
	Header("Location").Match("http://(?P<host>.+)/users/(?P<user>.+)")

m.Index(0).Equal("http://example.com/users/john")
m.Index(1).Equal("example.com")
m.Index(2).Equal("john")

m.Name("host").Equal("example.com")
m.Name("user").Equal("john")
```

##### Subdomains and per-request URL

```go
e.GET("/path").WithURL("http://example.com").
	Expect().
	Status(http.StatusOK)

e.GET("/path").WithURL("http://subdomain.example.com").
	Expect().
	Status(http.StatusOK)
```

##### WebSocket support

```go
ws := e.GET("/mysocket").WithWebsocketUpgrade().
	Expect().
	Status(http.StatusSwitchingProtocols).
	Websocket()
defer ws.Disconnect()

ws.WriteText("some request").
	Expect().
	TextMessage().Body().Equal("some response")

ws.CloseWithText("bye").
	Expect().
	CloseMessage().NoContent()
```

##### Reusable builders

```go
e := httpexpect.New(t, "http://example.com")

r := e.POST("/login").WithForm(Login{"ford", "betelgeuse7"}).
	Expect().
	Status(http.StatusOK).JSON().Object()

token := r.Value("token").String().Raw()

auth := e.Builder(func (req *httpexpect.Request) {
	req.WithHeader("Authorization", "Bearer "+token)
})

auth.GET("/restricted").
	Expect().
	Status(http.StatusOK)

e.GET("/restricted").
	Expect().
	Status(http.StatusUnauthorized)
```

##### Reusable matchers

```go
e := httpexpect.New(t, "http://example.com")

// every response should have this header
m := e.Matcher(func (resp *httpexpect.Response) {
	resp.Header("API-Version").NotEmpty()
})

m.GET("/some-path").
	Expect().
	Status(http.StatusOK)

m.GET("/bad-path").
	Expect().
	Status(http.StatusNotFound)
```

##### Custom config

```go
e := httpexpect.WithConfig(httpexpect.Config{
	// prepend this url to all requests
	BaseURL: "http://example.com",

	// use http.Client with a cookie jar and timeout
	Client: &http.Client{
		Jar:     httpexpect.NewJar(),
		Timeout: time.Second * 30,
	},

	// use fatal failures
	Reporter: httpexpect.NewRequireReporter(t),

	// use verbose logging
	Printers: []httpexpect.Printer{
		httpexpect.NewCurlPrinter(t),
		httpexpect.NewDebugPrinter(t, true),
	},
})
```

##### Session support

```go
// cookie jar is used to store cookies from server
e := httpexpect.WithConfig(httpexpect.Config{
	Reporter: httpexpect.NewAssertReporter(t),
	Client: &http.Client{
		Jar: httpexpect.NewJar(), // used by default if Client is nil
	},
})

// cookies are disabled
e := httpexpect.WithConfig(httpexpect.Config{
	Reporter: httpexpect.NewAssertReporter(t),
	Client: &http.Client{
		Jar: nil,
	},
})
```

##### Use HTTP handler directly

```go
// invoke http.Handler directly using httpexpect.Binder
var handler http.Handler = MyHandler()

e := httpexpect.WithConfig(httpexpect.Config{
	Reporter: httpexpect.NewAssertReporter(t),
	Client: &http.Client{
		Transport: httpexpect.NewBinder(handler),
		Jar:       httpexpect.NewJar(),
	},
})

// invoke fasthttp.RequestHandler directly using httpexpect.FastBinder
var handler fasthttp.RequestHandler = myHandler()

e := httpexpect.WithConfig(httpexpect.Config{
	Reporter: httpexpect.NewAssertReporter(t),
	Client: &http.Client{
		Transport: httpexpect.NewFastBinder(handler),
		Jar:       httpexpect.NewJar(),
	},
})
```

##### Per-request client or handler

```go
e := httpexpect.New(t, server.URL)

client := &http.Client{
	Transport: &http.Transport{
		DisableCompression: true,
	},
}

// overwrite client
e.GET("/path").WithClient(client).
	Expect().
	Status(http.StatusOK)

// construct client that invokes a handler directly and overwrite client
e.GET("/path").WithHandler(handler).
	Expect().
	Status(http.StatusOK)
```

## Similar packages

* [`gorequest`](https://github.com/parnurzeal/gorequest)
* [`baloo`](https://github.com/h2non/baloo)
* [`gofight`](https://github.com/appleboy/gofight)
* [`frisby`](https://github.com/verdverm/frisby)
* [`forest`](https://github.com/emicklei/forest)
* [`restit`](https://github.com/go-restit/restit)
* [`httptesting`](https://github.com/dolab/httptesting)
* [`http-test`](https://github.com/vsco/http-test)
* [`go-json-rest`](https://github.com/ant0ine/go-json-rest)

## Contributing

Feel free to report bugs, suggest improvements, and send pull requests! Please add documentation and tests for new features.

Update dependencies, build code, and run tests and linters:

```
$ make
```

Format code:

```
$ make fmt
```

## License

[MIT](LICENSE)
