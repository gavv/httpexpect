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
			name: "zero value",
			formatter: &DefaultFormatter{
				LineWidth: 0,
			},
			wantMaxLineWidth: 60,
		},
		{
			name: "small positive value",
			formatter: &DefaultFormatter{
				LineWidth: 15,
			},
			wantMaxLineWidth: 26,
		},
		{
			name: "very large positive value",
			formatter: &DefaultFormatter{
				LineWidth: 1000,
			},
			wantMaxLineWidth: 60,
		},
		{
			name: "negative value (no limit)",
			formatter: &DefaultFormatter{
				LineWidth: -1,
			},
			wantMaxLineWidth: 60,
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
				NotContainsValue(123)

			t.Logf("%s", rep.reported)

			var hasMaxWidth bool
			for _, s := range strings.Split(rep.reported, "\n") {
				width := len(s)
				if width >= tt.wantMaxLineWidth {
					hasMaxWidth = true
				}
				assert.LessOrEqual(t, width, tt.wantMaxLineWidth)
			}
			assert.True(t, hasMaxWidth)
		})
	}
}
