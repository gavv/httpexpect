package httpexpect

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

type recordingReporter struct {
	recorded string
}

func (r *recordingReporter) Errorf(msg string, args ...interface{}) {
	r.recorded += fmt.Sprintf(msg, args...) + "\n"
}

type recordingLogger struct {
	recorded string
}

func (r *recordingLogger) Logf(msg string, args ...interface{}) {
	r.recorded += fmt.Sprintf(msg, args...) + "\n"
}

func TestE2EReport_Names(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	reporter := &recordingReporter{}

	e := WithConfig(Config{
		TestName: "TestExample",
		BaseURL:  server.URL,
		Reporter: reporter,
	})

	e.GET("/test").
		WithName("RequestExample").
		Expect().
		JSON() // will fail

	t.Logf("%s", reporter.recorded)

	assert.Contains(t, reporter.recorded, "TestExample")
	assert.Contains(t, reporter.recorded, "RequestExample")
}

func TestE2EReport_Aliases(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	reporter := &recordingReporter{}

	e := WithConfig(Config{
		TestName: "TestExample",
		BaseURL:  server.URL,
		Reporter: reporter,
	})

	foo := e.GET("/test").
		WithName("RequestExample").
		Expect().
		JSON().Alias("foo")

	foo.Object().ContainsKey("bar") // will fail

	t.Logf("%s", reporter.recorded)

	assert.Contains(t, reporter.recorded, "foo.Object().ContainsKey()")
}

func TestE2EReport_LineWidth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"foo":{"bar":{"baz":[1,2,3]}}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	type widthRange struct {
		above int
		below int
	}

	tests := []struct {
		name        string
		formatter   *DefaultFormatter
		longestLine widthRange
	}{
		{
			name: "no limit",
			formatter: &DefaultFormatter{
				LineWidth: -1, // no limit
			},
			longestLine: widthRange{
				above: 100,
			},
		},
		{
			name: "large limit",
			formatter: &DefaultFormatter{
				LineWidth: 1000, // explicit limit - 1000 chars
			},
			longestLine: widthRange{
				above: 100,
			},
		},
		{
			name: "default limit",
			formatter: &DefaultFormatter{
				LineWidth: 0, // default limit - 60 chars
			},
			longestLine: widthRange{
				above: 40,
				below: 60,
			},
		},
		{
			name: "explicit limit",
			formatter: &DefaultFormatter{
				LineWidth: 30, // explicit limit - 30 chars
			},
			longestLine: widthRange{
				below: 30,
			},
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
				Value("foo").
				Object().
				Value("bar").
				Object().
				Value("baz").
				Array().
				NotContains(1)

			t.Logf("%s", rep.recorded)

			actualLongestLine := ""

			for _, s := range strings.Split(rep.recorded, "\n") {
				if len(actualLongestLine) < len(s) {
					actualLongestLine = s
				}
			}

			if tt.longestLine.above != 0 {
				assert.GreaterOrEqual(t, len(actualLongestLine), tt.longestLine.above)
			}
			if tt.longestLine.below != 0 {
				assert.LessOrEqual(t, len(actualLongestLine), tt.longestLine.below)
			}
		})
	}
}

func TestE2EReport_CustomTemplate(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	formatter := &DefaultFormatter{
		SuccessTemplate: "{{ .TestName | underscore }} succeeded",
		FailureTemplate: "{{ .TestName | underscore }} failed: " +
			"want {{ index .Expected 0 }}, got {{ .Actual }}",
		TemplateFuncs: template.FuncMap{
			"underscore": func(s string) string {
				var sb strings.Builder

				elems := strings.Split(s, " ")
				sb.WriteString(strings.Join(elems, "_"))

				return sb.String()
			},
		},
	}

	t.Run("success template", func(t *testing.T) {
		reporter := &recordingReporter{}
		logger := &recordingLogger{}

		e := WithConfig(Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			AssertionHandler: &DefaultAssertionHandler{
				Formatter: formatter,
				Reporter:  reporter,
				Logger:    logger,
			},
		})

		e.GET("/test").
			Expect()

		assert.Contains(t,
			logger.recorded,
			"formatter_test succeeded")
	})

	t.Run("failure template", func(t *testing.T) {
		reporter := &recordingReporter{}
		logger := &recordingLogger{}

		e := WithConfig(Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			AssertionHandler: &DefaultAssertionHandler{
				Formatter: formatter,
				Reporter:  reporter,
				Logger:    logger,
			},
		})

		e.GET("/test").
			Expect().
			JSON().
			Object().
			HasValue("foo", 456)

		assert.Contains(t,
			reporter.recorded,
			"formatter_test failed: want 456, got 123")
	})

	t.Run("invalid template", func(t *testing.T) {
		reporter := &recordingReporter{}
		logger := &recordingLogger{}

		e := WithConfig(Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			Reporter: reporter,
			AssertionHandler: &DefaultAssertionHandler{
				Formatter: &DefaultFormatter{
					SuccessTemplate: "{{ Invalid }}",
				},
				Reporter: reporter,
				Logger:   logger,
			},
		})

		assert.Panics(t, func() {
			e.GET("/test").
				Expect()
		})
	})

	t.Run("invalid field", func(t *testing.T) {
		reporter := &recordingReporter{}
		logger := &recordingLogger{}

		e := WithConfig(Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			Reporter: reporter,
			AssertionHandler: &DefaultAssertionHandler{
				Formatter: &DefaultFormatter{
					SuccessTemplate: "{{ .Invalid }}",
				},
				Reporter: reporter,
				Logger:   logger,
			},
		})

		assert.Panics(t, func() {
			e.GET("/test").
				Expect()
		})
	})
}

func TestE2EReport_RequestAndResponse(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"foo":{"bar":{"baz":[1,2,3]}}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name      string
		formatter *DefaultFormatter
	}{
		{
			name: "request and response enabled",
			formatter: &DefaultFormatter{
				EnableRequests:  true,
				EnableResponses: true,
			},
		},
		{
			name: "request enabled, response disabled",
			formatter: &DefaultFormatter{
				EnableRequests:  true,
				EnableResponses: false,
			},
		},
		{
			name: "request disabled, response enabled",
			formatter: &DefaultFormatter{
				EnableRequests:  false,
				EnableResponses: true,
			},
		},
		{
			name: "request and response disabled",
			formatter: &DefaultFormatter{
				EnableRequests:  false,
				EnableResponses: false,
			},
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
				Value("foo").
				Object().
				Value("bar").
				Object().
				Value("baz").
				Array().
				NotContains(1)

			logs := rep.recorded

			if tt.formatter.EnableRequests {
				assert.Contains(t,
					logs,
					"GET /test HTTP/1.1",
					"expected log output to contain request information")
			} else {
				assert.NotContains(t,
					logs,
					"GET /test HTTP/1.1",
					"expected log output not to contain request information")
			}

			if tt.formatter.EnableResponses {
				assert.Contains(t,
					logs,
					"HTTP/1.1 200 OK", "expected log output to contain response information")
			} else {
				assert.NotContains(t,
					logs,
					"HTTP/1.1 200 OK",
					"expected log output not to contain response information")
			}
		})
	}
}
