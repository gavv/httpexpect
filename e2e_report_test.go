package httpexpect

import (
	"fmt"
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

			t.Logf("%s", rep.reported)

			actualLongestLine := ""

			for _, s := range strings.Split(rep.reported, "\n") {
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
