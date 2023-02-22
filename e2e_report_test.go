package httpexpect

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type recordingReporter struct {
	reported string
}

func (r *recordingReporter) Errorf(msg string, args ...interface{}) {
	r.reported += fmt.Sprintf(msg, args...)
}

func TestE2EReport_Names(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	rep := &recordingReporter{}

	e := WithConfig(Config{
		TestName: "TestExample",
		BaseURL:  server.URL,
		Reporter: rep,
	})

	e.GET("/test").
		WithName("RequestExample").
		Expect().
		JSON() // will fail

	t.Logf("%s", rep.reported)

	assert.Contains(t, rep.reported, "TestExample")
	assert.Contains(t, rep.reported, "RequestExample")
}

func TestE2EReport_Aliases(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	rep := &recordingReporter{}

	e := WithConfig(Config{
		TestName: "TestExample",
		BaseURL:  server.URL,
		Reporter: rep,
	})

	foo := e.GET("/test").
		WithName("RequestExample").
		Expect().
		JSON().Alias("foo")

	foo.Object().ContainsKey("bar") // will fail

	t.Logf("%s", rep.reported)

	assert.Contains(t, rep.reported, "foo.Object().ContainsKey()")
}

const (
	failureTemplateAssertPath = `{{ join .AssertPath .LineWidth | indent }}`
	failureTemplateErrors     = `
	{{- range $n, $err := .Errors }}
	{{ if eq $n 0 -}}
	{{ wrap $err $.LineWidth }}
	{{- else -}}
	{{ wrap $err $.LineWidth | indent }}
	{{- end -}}
	{{- end -}}`
)

func TestE2EReport_LineWidth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name             string
		formatter        *DefaultFormatter
		wantMaxLineWidth int
	}{
		{
			name: "assertPath zero value",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateAssertPath,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       0,
			},
			wantMaxLineWidth: defaultLineWidth,
		},
		{
			name: "assertPath small positive value",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateAssertPath,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       15,
			},
			wantMaxLineWidth: 15,
		},
		{
			name: "assertPath very large positive value",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateAssertPath,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       1000,
			},
			wantMaxLineWidth: 1000,
		},
		{
			name: "assertPath negative value (no limit)",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateAssertPath,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       -1,
			},
			wantMaxLineWidth: math.MaxInt,
		},
		{
			name: "errors zero value",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateErrors,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       0,
			},
			wantMaxLineWidth: defaultLineWidth,
		},
		{
			name: "errors small positive value",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateErrors,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       15,
			},
			wantMaxLineWidth: 15,
		},
		{
			name: "errors very large positive value",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateErrors,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       1000,
			},
			wantMaxLineWidth: 1000,
		},
		{
			name: "errors negative value (no limit)",
			formatter: &DefaultFormatter{
				FailureTemplate: failureTemplateErrors,
				TemplateFuncs:   defaultTemplateFuncs,
				LineWidth:       -1,
			},
			wantMaxLineWidth: math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := &recordingReporter{}

			e := WithConfig(Config{
				TestName: "TestExample",
				BaseURL:  server.URL,
				AssertionHandler: &DefaultAssertionHandler{
					Formatter: tt.formatter,
					Reporter:  rep,
				},
			})

			e.GET("/test").
				Expect().
				JSON().
				Object().
				IsValueEqual("foo", 1234)

			scanner := bufio.NewScanner(strings.NewReader(rep.reported))

			for scanner.Scan() {
				s := strings.Trim(strings.TrimSpace(scanner.Text()), ".")

				var ss []string
				switch tt.formatter.FailureTemplate {
				case failureTemplateAssertPath:
					ss = strings.Split(s, ".")
				case failureTemplateErrors:
					ss = strings.Fields(s)
				}

				if len(ss) <= 1 {
					continue
				}

				lenBeforeWrapped := len(s) - len(ss[len(ss)-1]) - 1

				t.Logf("%s", s)

				assert.LessOrEqual(t, lenBeforeWrapped, tt.wantMaxLineWidth)
			}
		})
	}
}
