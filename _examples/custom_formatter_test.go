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
Test Failed!
	test name: {{ .TestName | underscore }}
	request name: {{ .RequestName }}
	error details: {{ .Errors }}
	actual value: {{ .Actual }}
	reference value: {{ .Reference }}
`

var templateFuncs = template.FuncMap{
	"underscore": func(s string) string {
		var sb strings.Builder

		elems := strings.Split(s, " ")
		sb.WriteString(strings.Join(elems, "_"))

		return sb.String()
	},
}

func TestDefaultFormatter(t *testing.T) {
	handler := CustomFormatterHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	// customize formatting template
	e := httpexpect.WithConfig(httpexpect.Config{
		TestName: "Check Fruits Name",
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Formatter: &httpexpect.DefaultFormatter{
			FailureTemplate: failureTemplate,
			TemplateFuncs:   templateFuncs,
		},
	})

	fruits := e.GET("/fruits").WithName("Get Fruits").
		Expect().
		Status(http.StatusOK).JSON().Array()
	fruits.Every(func(index int, value *httpexpect.Value) {
		value.String().NotEmpty()
	})
	fruits.ContainsAny("melon")
}
