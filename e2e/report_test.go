package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/gavv/httpexpect/v2"
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

	e := httpexpect.WithConfig(httpexpect.Config{
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

func TestE2EReport_Values(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ACTUAL TEXT"))
	})
	mux.HandleFunc("/number", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("88888888"))
	})
	mux.HandleFunc("/array", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[111, 222, 444, 333]"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	formatter := &httpexpect.DefaultFormatter{
		DigitSeparator: httpexpect.DigitSeparatorNone,
	}

	t.Run("actual vs expected", func(t *testing.T) {
		reporter := &recordingReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:   server.URL,
			Reporter:  reporter,
			Formatter: formatter,
		})

		e.GET("/text").
			Expect().
			Body().
			IsEqual("EXPECTED TEXT")
		t.Logf("%s", reporter.recorded)

		assert.Contains(
			t, reporter.recorded, "expected: strings are equal",
			"missing Errors",
		)
		assert.Contains(
			t, reporter.recorded, "ACTUAL TEXT", "missing Actual",
		)
		assert.Contains(
			t, reporter.recorded, "EXPECTED TEXT", "missing Expected",
		)
	})

	t.Run("reference", func(t *testing.T) {
		reporter := &recordingReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:   server.URL,
			Reporter:  reporter,
			Formatter: formatter,
		})

		e.GET("/array").
			Expect().
			JSON().
			Array().
			IsOrdered()
		t.Logf("%s", reporter.recorded)

		trimmed := strings.Join(strings.Fields(reporter.recorded), "")

		assert.Contains(
			t, reporter.recorded, "expected: reference array is ordered",
			"missing Errors",
		)
		assert.Contains(
			t, trimmed, "[\"111\",\"222\",\"444\",\"333\"]", "missing Reference",
		)
	})

	t.Run("delta", func(t *testing.T) {
		reporter := &recordingReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:   server.URL,
			Reporter:  reporter,
			Formatter: formatter,
		})

		e.GET("/number").
			Expect().
			Body().
			AsNumber().
			InDelta(9999999, 0.55555555)
		t.Logf("%s", reporter.recorded)

		assert.Contains(
			t, reporter.recorded, "expected: numbers lie within delta",
			"missing Errors",
		)
		assert.Contains(
			t, reporter.recorded, "8888888", "missing Actual",
		)
		assert.Contains(
			t, reporter.recorded, "9999999", "missing Expected",
		)
		assert.Contains(
			t, reporter.recorded, "0.55555555", "missing Delta",
		)
	})

	t.Run("range", func(t *testing.T) {
		reporter := &recordingReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:   server.URL,
			Reporter:  reporter,
			Formatter: formatter,
		})

		e.GET("/number").
			Expect().
			Body().
			AsNumber().
			InRange(333333, 444444)
		t.Logf("%s", reporter.recorded)

		assert.Contains(
			t, reporter.recorded, "expected: number is within given range",
			"missing Errors",
		)
		assert.Contains(
			t, reporter.recorded, "8888888", "missing Actual",
		)
		assert.Contains(
			t, reporter.recorded, "[333333; 444444]", "missing Expected",
		)
	})

	t.Run("list", func(t *testing.T) {
		reporter := &recordingReporter{}

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:   server.URL,
			Reporter:  reporter,
			Formatter: formatter,
		})

		e.GET("/number").
			Expect().
			Body().
			AsNumber().
			InList(333333, 444444)
		t.Logf("%s", reporter.recorded)

		trimmed := strings.Join(strings.Fields(reporter.recorded), " ")

		assert.Contains(
			t, reporter.recorded, "expected: number is equal to one of the values",
			"missing Errors",
		)
		assert.Contains(
			t, reporter.recorded, "8888888", "missing Actual",
		)
		assert.Contains(
			t, trimmed, "333333 444444", "missing Expected",
		)
	})
}

func TestE2EReport_Path(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	reporter := &recordingReporter{}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: reporter,
	})

	e.GET("/test").
		Expect().
		JSON().
		Object().
		ContainsKey("bar") // will fail

	t.Logf("%s", reporter.recorded)

	assert.Contains(
		t,
		reporter.recorded,
		`Request("GET").Expect().JSON().Object()`,
		"cannot find Path value in report",
	)
}

func TestE2EReport_Alias(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	reporter := &recordingReporter{}

	e := httpexpect.WithConfig(httpexpect.Config{
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

	assert.Contains(
		t,
		reporter.recorded,
		"foo.Object().ContainsKey()",
		"cannot find AliasedPath value in report")
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

	cases := []struct {
		name        string
		formatter   *httpexpect.DefaultFormatter
		longestLine widthRange
	}{
		{
			name: "no limit",
			formatter: &httpexpect.DefaultFormatter{
				LineWidth: -1, // no limit
			},
			longestLine: widthRange{
				above: 100,
			},
		},
		{
			name: "large limit",
			formatter: &httpexpect.DefaultFormatter{
				LineWidth: 1000, // explicit limit - 1000 chars
			},
			longestLine: widthRange{
				above: 100,
			},
		},
		{
			name: "default limit",
			formatter: &httpexpect.DefaultFormatter{
				LineWidth: 0, // default limit - 60 chars
			},
			longestLine: widthRange{
				above: 40,
				below: 60,
			},
		},
		{
			name: "explicit limit",
			formatter: &httpexpect.DefaultFormatter{
				LineWidth: 30, // explicit limit - 30 chars
			},
			longestLine: widthRange{
				below: 30,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := &recordingReporter{}

			fmt := tc.formatter
			fmt.DisableRequests = true
			fmt.DisableResponses = true

			e := httpexpect.WithConfig(httpexpect.Config{
				TestName: "TestExample",
				BaseURL:  server.URL,
				AssertionHandler: &httpexpect.DefaultAssertionHandler{
					Formatter: fmt,
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
				NotContainsAll(1)

			t.Logf("%s", rep.recorded)

			actualLongestLine := ""

			for _, s := range strings.Split(rep.recorded, "\n") {
				if len(actualLongestLine) < len(s) {
					actualLongestLine = s
				}
			}

			if tc.longestLine.above != 0 {
				assert.GreaterOrEqual(t, len(actualLongestLine), tc.longestLine.above)
			}
			if tc.longestLine.below != 0 {
				assert.LessOrEqual(t, len(actualLongestLine), tc.longestLine.below)
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

	formatter := &httpexpect.DefaultFormatter{
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

		e := httpexpect.WithConfig(httpexpect.Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			AssertionHandler: &httpexpect.DefaultAssertionHandler{
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

		e := httpexpect.WithConfig(httpexpect.Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			AssertionHandler: &httpexpect.DefaultAssertionHandler{
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

		e := httpexpect.WithConfig(httpexpect.Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			Reporter: reporter,
			AssertionHandler: &httpexpect.DefaultAssertionHandler{
				Formatter: &httpexpect.DefaultFormatter{
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

		e := httpexpect.WithConfig(httpexpect.Config{
			TestName: "formatter test",
			BaseURL:  server.URL,
			Reporter: reporter,
			AssertionHandler: &httpexpect.DefaultAssertionHandler{
				Formatter: &httpexpect.DefaultFormatter{
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

func TestE2EReport_RequestResponse(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"foo":{"bar":{"baz":[1,2,3]}}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cases := []struct {
		name      string
		formatter *httpexpect.DefaultFormatter
	}{
		{
			name: "request and response enabled",
			formatter: &httpexpect.DefaultFormatter{
				DisableRequests:  false,
				DisableResponses: false,
			},
		},
		{
			name: "request enabled, response disabled",
			formatter: &httpexpect.DefaultFormatter{
				DisableRequests:  false,
				DisableResponses: true,
			},
		},
		{
			name: "request disabled, response enabled",
			formatter: &httpexpect.DefaultFormatter{
				DisableRequests:  true,
				DisableResponses: false,
			},
		},
		{
			name: "request and response disabled",
			formatter: &httpexpect.DefaultFormatter{
				DisableRequests:  true,
				DisableResponses: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := &recordingReporter{}

			e := httpexpect.WithConfig(httpexpect.Config{
				TestName: "TestExample",
				BaseURL:  server.URL,
				AssertionHandler: &httpexpect.DefaultAssertionHandler{
					Formatter: tc.formatter,
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
				NotContainsAll(1)

			logs := rep.recorded

			if tc.formatter.DisableRequests {
				assert.NotContains(t,
					logs,
					"GET /test HTTP/1.1",
					"expected log output not to contain request information")
			} else {
				assert.Contains(t,
					logs,
					"GET /test HTTP/1.1",
					"expected log output to contain request information")
			}

			if tc.formatter.DisableResponses {
				assert.NotContains(t,
					logs,
					"HTTP/1.1 200 OK",
					"expected log output not to contain response information")
			} else {
				assert.Contains(t,
					logs,
					"HTTP/1.1 200 OK", "expected log output to contain response information")
			}
		})
	}
}
