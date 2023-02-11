package examples

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

var failureTemplate = `
Test Failure!
	test name: {{ .TestName | underscore }}
	request name: {{ .RequestName }}
	error details: {{ .Errors }}
	actual value: {{ .Actual }}
	reference value: {{ .Reference }}
`

var successTemplate = `
Test Success!
	test name: {{ .TestName | underscore }}
	request name: {{ .RequestName }}
`

var templateFuncs = template.FuncMap{
	"underscore": func(s string) string {
		var sb strings.Builder

		elems := strings.Split(s, " ")
		sb.WriteString(strings.Join(elems, "_"))

		return sb.String()
	},
}

// TestFailureTemplate tests with custom `FailureTemplate`
func TestFailureTemplate(t *testing.T) {
	t.Skip() // remove this line to see test failure

	handler := FruitsHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		TestName: "Check Fruits List",
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Formatter: &httpexpect.DefaultFormatter{
			// customize error message with `FailureTemplate`
			FailureTemplate: failureTemplate,
			TemplateFuncs:   templateFuncs,
		},
	})

	orange := map[string]interface{}{
		"weight": 100,
	}

	e.PUT("/fruits/orange").WithJSON(orange).
		Expect().
		Status(http.StatusNoContent).NoContent()

	fruits := e.GET("/fruits").WithName("Get Fruits").
		Expect().
		Status(http.StatusOK).JSON().Array()

	fruits.ContainsAny("melon")
}

// TestSuccessTemplate tests with custom `SuccessTemplate`
func TestSuccessTemplate(t *testing.T) {
	handler := FruitsHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		TestName: "Check Fruits List",
		BaseURL:  server.URL,
		AssertionHandler: &httpexpect.DefaultAssertionHandler{
			Formatter: &httpexpect.DefaultFormatter{
				// customize success message with `SuccessTemplate`
				SuccessTemplate: successTemplate,
				TemplateFuncs:   templateFuncs,
			},
			Reporter: t,
			// to enable printing of success messages, we need to set `Logger`
			Logger: t,
		},
	})

	orange := map[string]interface{}{
		"weight": 100,
	}

	e.PUT("/fruits/orange").WithJSON(orange).
		Expect().
		Status(http.StatusNoContent).NoContent()

	fruits := e.GET("/fruits").WithName("Get Fruits").
		Expect().
		Status(http.StatusOK).JSON().Array()

	fruits.ContainsAny("orange")
}
