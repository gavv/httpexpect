package examples

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func irisTester(t *testing.T) *httpexpect.Expect {
	handler := IrisHandler()

	return httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewCookieJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestIrisThings(t *testing.T) {
	e := irisTester(t)

	schema := `{
		"type": "array",
		"items": {
			"type": "object",
			"properties": {
				"name":        {"type": "string"},
				"description": {"type": "string"}
			},
			"required": ["name", "description"]
		}
	}`

	things := e.GET("/things").
		Expect().
		Status(http.StatusOK).JSON()

	things.Schema(schema)

	names := things.Path("$[*].name").Array()

	names.ConsistsOf("foo", "bar")

	for n, desc := range things.Path("$..description").Array().Iter() {
		m := desc.String().Match("(.+) (.+)")

		m.Index(1).IsEqual(names.Element(n).String().Raw())
		m.Index(2).IsEqual("thing")
	}
}

func TestIrisRedirect(t *testing.T) {
	e := irisTester(t)

	things := e.POST("/redirect").
		Expect().
		Status(http.StatusOK).JSON().Array()

	things.Length().IsEqual(2)

	things.Element(0).Object().ValueEqual("name", "foo")
	things.Element(1).Object().ValueEqual("name", "bar")
}

func TestIrisParams(t *testing.T) {
	e := irisTester(t)

	type Form struct {
		P1 string `form:"p1"`
		P2 string `form:"p2"`
	}

	// POST /params/xxx/yyy?q=qqq
	// Form: p1=P1&p2=P2

	r := e.POST("/params/{x}/{y}", "xxx", "yyy").
		WithQuery("q", "qqq").
		WithForm(Form{P1: "P1", P2: "P2"}).
		Expect().
		Status(http.StatusOK).JSON().Object()

	r.Value("x").IsEqual("xxx")
	r.Value("y").IsEqual("yyy")
	r.Value("q").IsEqual("qqq")

	r.ValueEqual("p1", "P1")
	r.ValueEqual("p2", "P2")
}

func TestIrisAuth(t *testing.T) {
	e := irisTester(t)

	e.GET("/auth").
		Expect().
		Status(http.StatusUnauthorized)

	e.GET("/auth").WithBasicAuth("ford", "<bad password>").
		Expect().
		Status(http.StatusUnauthorized)

	e.GET("/auth").WithBasicAuth("ford", "betelgeuse7").
		Expect().
		Status(http.StatusOK).Body().IsEqual("authenticated!")
}

func TestIrisSession(t *testing.T) {
	e := irisTester(t)

	sub := e.Builder(func(req *httpexpect.Request) {
		req.WithURL("http://example.com")
	})

	sub.POST("/session/set").WithJSON(map[string]string{"name": "test"}).
		Expect().
		Status(http.StatusOK).Cookies().NotEmpty()

	r := sub.GET("/session/get").
		Expect().
		Status(http.StatusOK).JSON().Object()

	r.IsEqual(map[string]string{
		"name": "test",
	})
}

func TestIrisStream(t *testing.T) {
	e := irisTester(t)

	e.GET("/stream").
		Expect().
		Status(http.StatusOK).
		TransferEncoding("chunked"). // ensure server sent chunks
		Body().IsEqual("0123456789")

	// send chunks to server
	e.POST("/stream").WithChunked(strings.NewReader("<long text>")).
		Expect().
		Status(http.StatusOK).Body().IsEqual("<long text>")
}

func TestIrisSubdomain(t *testing.T) {
	e := irisTester(t)

	sub := e.Builder(func(req *httpexpect.Request) {
		req.WithURL("http://subdomain.example.com")
	})

	sub.POST("/set").
		Expect().
		Status(http.StatusOK)

	sub.GET("/get").
		Expect().
		Status(http.StatusOK).
		Body().IsEqual("hello from subdomain")
}
