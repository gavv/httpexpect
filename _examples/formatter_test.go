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

	// customize error message with `FailureTemplate`
	e := httpexpect.WithConfig(httpexpect.Config{
		TestName: "Check Fruits List",
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Formatter: &httpexpect.DefaultFormatter{
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

	// Customize success message with `SuccessTemplate`.
	// To make it prints on success we need to set `AssertionHandler`
	// with `t` as the `Logger`
	e := httpexpect.WithConfig(httpexpect.Config{
		TestName: "Check Fruits List",
		BaseURL:  server.URL,
		AssertionHandler: &httpexpect.DefaultAssertionHandler{
			Formatter: &httpexpect.DefaultFormatter{
				SuccessTemplate: successTemplate,
				TemplateFuncs:   templateFuncs,
			},
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
